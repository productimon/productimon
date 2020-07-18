package main

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
func (s *service) addLabelToQueue(app string) {
	select {
	case labelChan <- app:
	default:
	}
}

// get system-wide label for app
// return LABEL_UNCATEGORIZED if it's not yet ready
func (s *service) getDefaultLabel(app string, tx *sql.Tx) string {
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

// blocking call to ensure app is labelled
func (s *service) ensureLabel(app string) {
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
func (s *service) scanDbToLabelQueue() {
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
func (s *service) runLabelRoutine() {
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
