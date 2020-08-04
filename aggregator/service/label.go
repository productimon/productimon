package service

import (
	"context"
	"database/sql"
	"time"

	"git.yiad.am/productimon/analyzer/nlp"
	cpb "git.yiad.am/productimon/proto/common"
	spb "git.yiad.am/productimon/proto/svc"
	lru "github.com/hashicorp/golang-lru"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	labelbuffersize      = 1024
	labelcachesize       = 1024
	labelDbCheckInterval = 5 * time.Minute
)

const (
	/* please make sure there are no special characters in the following constants */
	/* we're directly putting them in sql statement without escaping in some places */
	LABEL_UNCATEGORIZED = "Uncategorized" // not yet guessed
	LABEL_UNKNOWN       = "Unknown"       // can't guess
)

var labelChan chan string
var labelCache *lru.TwoQueueCache

// add label to queue on best effort
// directly return if queue is full
func (s *Service) addLabelToQueue(app string) {
	select {
	case labelChan <- app:
	default:
	}
}

// get system-wide label for app
// return LABEL_UNCATEGORIZED if it's not yet ready
func (s *Service) getDefaultLabel(app string, tx *sql.Tx) string {
	if label, ok := labelCache.Get(app); ok {
		return label.(string)
	}
	var label string
	if err := tx.QueryRow("SELECT label FROM default_apps WHERE name=? LIMIT 1", app).Scan(&label); err == nil && label != LABEL_UNCATEGORIZED {
		labelCache.Add(app, label)
		return label
	}
	s.addLabelToQueue(app)
	return LABEL_UNCATEGORIZED
}

func (s *Service) getLabel(uid, appname string, tx *sql.Tx) (label string) {
	if err := tx.QueryRow("SELECT label FROM user_apps WHERE uid=? AND name=? LIMIT 1", uid, appname).Scan(&label); err == nil {
		return
	}
	label = s.getDefaultLabel(appname, tx)
	// if the current label is UNKNOWN (which means we tried to guess it but failed),
	// and is later updated to a real label, this update is reflected to the user
	//
	// if the current label is not UNKNOWN but a valid tag and is later changed to a
	// different label, i don't want it updated for users who already saw this label
	// before (it's weird to change user data after they saw it)
	//
	// e.g. if zoom is unknown but later changed to videoconference, this is updated
	// for the user. if zoom is videoconference and the user already saw it, but
	// admin changes it to meeting later, we don't want this change to take place
	// automatically for old users. but for new users who never used it before, they
	// have the new label
	if label != LABEL_UNKNOWN && label != LABEL_UNCATEGORIZED {
		if _, err := tx.Exec("INSERT INTO user_apps (uid, name, label) VALUES(?, ?, ?)", uid, appname, label); err != nil {
			s.log.Error("failed to insert to user_apps", zap.Error(err), zap.String("uid", uid), zap.String("appname", appname), zap.String("label", label))
		}
	}
	return
}

// blocking call to ensure app is labelled
func (s *Service) ensureLabel(app string) {
	var label string
	if err := s.db.QueryRow("SELECT label FROM default_apps WHERE name=? LIMIT 1", app).Scan(&label); err == nil && label != LABEL_UNCATEGORIZED {
		labelCache.Add(app, label)
		return
	}
	label = nlp.GuessLabel(app)
	s.log.Info("guessed label", zap.String("app", app), zap.String("label", label))
	s.dbWLock.Lock()
	if _, err := s.db.Exec("INSERT INTO default_apps (name, label) VALUES (?, ?)", app, label); err != nil {
		s.log.Error("cannot insert label into default_apps", zap.Error(err), zap.String("app", app), zap.String("label", label))
	}
	s.dbWLock.Unlock()
}

// scan db for any remaining uncatogorized apps and add them to queue on best effort
func (s *Service) scanDbToLabelQueue() {
	if rows, err := s.db.Query("SELECT name FROM default_apps WHERE label = ''"); err == nil {
		defer rows.Close()
		for rows.Next() {
			var app string
			rows.Scan(&app)
			s.addLabelToQueue(app)
		}
	}
}

// resolve local queue and scan db periodically for uncategorized apps(in case queue if full or we reboot server)
// to be run in its own goroutine
func (s *Service) RunLabelRoutine() {
	var err error
	labelChan = make(chan string, labelbuffersize)
	if labelCache, err = lru.New2Q(labelcachesize); err != nil {
		panic(err)
	}
	s.scanDbToLabelQueue()
	timer := time.NewTicker(labelDbCheckInterval)
	for {
		select {
		case app := <-labelChan:
			s.ensureLabel(app)
		case <-timer.C:
			s.scanDbToLabelQueue()
		}
	}
}

func (s *Service) GetLabels(ctx context.Context, req *spb.DataAggregatorGetLabelsRequest) (*spb.DataAggregatorGetLabelsResponse, error) {
	uid, did, err := s.auther.AuthenticateRequest(ctx)
	if err != nil || did != -1 {
		return nil, status.Error(codes.Unauthenticated, "Invalid token")
	}

	var rows *sql.Rows
	if req.AllLabels {
		if !s.isAdmin(ctx) {
			return nil, status.Error(codes.Unauthenticated, "You must be an admin to change system-level labels")
		}
		rows, err = s.db.Query("SELECT a.name, a.label, COUNT(DISTINCT i.uid) FROM default_apps a, intervals i WHERE a.name = i.app GROUP BY a.name ORDER BY a.name COLLATE NOCASE ASC")
	} else {
		rows, err = s.db.Query("SELECT i.app, COALESCE(u.label, d.label, ?), 0 FROM (SELECT DISTINCT app FROM intervals WHERE uid = ?) i LEFT JOIN user_apps u ON i.app = u.name AND u.uid = ? LEFT JOIN default_apps d ON i.app = d.name ORDER BY i.app COLLATE NOCASE ASC", LABEL_UNCATEGORIZED, uid, uid)
	}
	rsp := &spb.DataAggregatorGetLabelsResponse{}

	switch {
	case err == nil:
		defer rows.Close()
		for rows.Next() {
			label := &cpb.Label{}
			if err = rows.Scan(&label.App, &label.Label, &label.UsedBy); err != nil {
				s.log.Error("failed to scan label", zap.Error(err))
				continue
			}
			rsp.Labels = append(rsp.Labels, label)
		}
	case err == sql.ErrNoRows:
	default:
		s.log.Error("Failed to get labels", zap.Error(err))
		return nil, status.Error(codes.Internal, "something went wrong")
	}
	return rsp, nil
}

func (s *Service) UpdateLabel(ctx context.Context, req *spb.DataAggregatorUpdateLabelRequest) (*cpb.Empty, error) {
	uid, did, err := s.auther.AuthenticateRequest(ctx)
	if err != nil || did != -1 {
		return nil, status.Error(codes.Unauthenticated, "Invalid token")
	}

	if req.AllLabels {
		if !s.isAdmin(ctx) {
			return nil, status.Error(codes.Unauthenticated, "You must be an admin to change system-level labels")
		}
		_, err = s.db.Exec("UPDATE default_apps SET label = ? WHERE name = ?", req.Label.Label, req.Label.App)
		labelCache.Remove(req.Label.App)
	} else {
		var result sql.Result
		result, err = s.db.Exec("UPDATE user_apps SET label = ? WHERE name = ? AND uid = ?", req.Label.Label, req.Label.App, uid)
		if err == nil {
			var rows int64
			if rows, _ = result.RowsAffected(); rows == 0 {
				_, err = s.db.Exec("INSERT INTO user_apps (label, name, uid) VALUES (?, ?, ?)", req.Label.Label, req.Label.App, uid)
			}
		}
	}

	if err != nil {
		s.log.Error("failed to update label", zap.Error(err))
		return nil, status.Error(codes.Internal, "something went wrong")
	}

	return &cpb.Empty{}, nil
}
