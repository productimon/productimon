package service

import (
	"database/sql"
	"time"

	"git.yiad.am/productimon/aggregator/nlp"
	lru "github.com/hashicorp/golang-lru"
	"go.uber.org/zap"
)

const (
	labelbuffersize      = 1024
	labelcachesize       = 1024
	labelDbCheckInterval = 5 * time.Minute
	LABEL_UNCATEGORIZED  = "Uncatogorized" // not yet guessed
	LABEL_UNKNOWN        = "Unknown"       // can't guess
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
