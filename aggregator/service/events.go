package service

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"sync"

	"git.yiad.am/productimon/aggregator/deviceState"
	cpb "git.yiad.am/productimon/proto/common"
	spb "git.yiad.am/productimon/proto/svc"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TODO: what if we recoreded out-of-order events in db and are waiting for an old event when we shutdown server
func (s *Service) lazyInitEidHandler(uid string, did int64) (int64, error) {
	var eid int64
	if err := s.db.QueryRow("SELECT COALESCE(MAX(id), 0) FROM events WHERE uid=? AND did=?", uid, did).Scan(&eid); err != nil {
		return -1, err
	}
	return eid, nil
}

// add a event to events table in a new transaction and return that transaction
// no need to rollback tx if err != nil
func (s *Service) addGeneralEvent(uid string, did int64, e *cpb.Event, kind cpb.EventType) (tx *sql.Tx, err error) {
	tx, err = s.db.Begin()
	if err != nil {
		return
	}
	_, err = tx.Exec("INSERT INTO events (uid, did, id, kind, starttime, endtime) VALUES(?, ?, ?, ?, ?, ?)",
		uid, did, e.Id, kind, e.Timeinterval.Start.Nanos, e.Timeinterval.End.Nanos)
	if err != nil {
		tx.Rollback()
	}
	return
}

func (s *Service) eventUpdateState(uid string, did int64, e *cpb.Event, eg func(e *cpb.Event) func(ds *deviceState.DeviceState, db *sql.DB, dblock *sync.Mutex, logger *zap.Logger)) error {
	err := s.ds.RunEvent(s.db, s.dbWLock, s.log, uid, did, e.Id, eg(e))
	if err != nil {
		s.log.Error("RunEvent error", zap.Error(err), zap.String("uid", uid), zap.Int64("did", did), zap.Int64("eid", e.Id))
	}
	return err // this is currently ignored by AddEvent
}

func (s *Service) AddEvent(uid string, did int64, e *cpb.Event) error {
	switch k := e.Kind.(type) {
	case *cpb.Event_AppSwitchEvent:
		s.dbWLock.Lock()
		tx, err := s.addGeneralEvent(uid, did, e, cpb.EventType_APP_SWITCH_EVENT)
		if err != nil {
			s.dbWLock.Unlock()
			return err
		}
		if _, err := tx.Exec("INSERT INTO app_switch_events(uid, did, id, app) VALUES(?, ?, ?, ?)",
			uid, did, e.Id, k.AppSwitchEvent.AppName); err != nil {
			tx.Rollback()
			s.dbWLock.Unlock()
			return err
		}
		s.getDefaultLabel(k.AppSwitchEvent.AppName, tx) // either it exists or we add it to queue
		tx.Commit()
		s.dbWLock.Unlock()
		s.eventUpdateState(uid, did, e, deviceState.SwitchApp)

	case *cpb.Event_StartTrackingEvent:
		s.dbWLock.Lock()
		tx, err := s.addGeneralEvent(uid, did, e, cpb.EventType_START_TRACKING_EVENT)
		if err != nil {
			s.dbWLock.Unlock()
			return err
		}
		tx.Commit()
		s.dbWLock.Unlock()
		s.eventUpdateState(uid, did, e, deviceState.ClearState)

	case *cpb.Event_StopTrackingEvent:
		s.dbWLock.Lock()
		tx, err := s.addGeneralEvent(uid, did, e, cpb.EventType_STOP_TRACKING_EVENT)
		if err != nil {
			s.dbWLock.Unlock()
			return err
		}
		tx.Commit()
		s.dbWLock.Unlock()
		s.eventUpdateState(uid, did, e, deviceState.ClearState)

	case *cpb.Event_ActivityEvent:
		s.dbWLock.Lock()
		tx, err := s.addGeneralEvent(uid, did, e, cpb.EventType_ACTIVITY_EVENT)
		if err != nil {
			s.dbWLock.Unlock()
			return err
		}
		if _, err := tx.Exec("INSERT INTO activity_events(uid, did, id, keystrokes, mouseclicks) VALUES(?, ?, ?, ?, ?)",
			uid, did, e.Id, k.ActivityEvent.Keystrokes, k.ActivityEvent.Mouseclicks); err != nil {
			tx.Rollback()
			s.dbWLock.Unlock()
			return err
		}
		tx.Commit()
		s.dbWLock.Unlock()
		if s.isActive(k.ActivityEvent.Keystrokes, k.ActivityEvent.Mouseclicks, e.Timeinterval.Start.Nanos, e.Timeinterval.End.Nanos) {
			s.eventUpdateState(uid, did, e, deviceState.SetActive)
		} else {
			s.eventUpdateState(uid, did, e, deviceState.Nop)
		}

	case nil:
		return errors.New("event not set")

	default:
		return fmt.Errorf("unknown event type %T", k)
	}

	return nil
}

func (s *Service) PushEvent(server spb.DataAggregator_PushEventServer) error {
	s.log.Info("Started pushEvent stream")
	ctx := server.Context()
	uid, did, err := s.auther.AuthenticateRequest(ctx)
	if err != nil || did == -1 {
		s.log.Error("Failed to authenticate pushEvent", zap.Error(err), zap.String("uid", uid), zap.Int64("did", did))
		return status.Error(codes.Unauthenticated, "Invalid token")
	}

	for {
		select {
		case <-ctx.Done():
			s.log.Warn("context cancelled", zap.Error(ctx.Err()))
			return ctx.Err()
		default:
		}

		event, err := server.Recv()
		switch {
		case err == io.EOF:
			s.log.Info("Client closed the stream, we're closing too")
			return nil
		case err != nil:
			s.log.Error("receive error", zap.Error(err))
			continue
		}
		s.log.Sugar().Info("received event ", event)

		if err = s.AddEvent(uid, did, event); err != nil {
			s.log.Error("Failed to add event", zap.Error(err), zap.String("uid", uid), zap.Int64("did", did), zap.Int64("eid", event.Id))
		}
	}
	return nil
}

func (s *Service) GetEvent(req *spb.DataAggregatorGetEventRequest, server spb.DataAggregator_GetEventServer) error {
	return status.Error(codes.Unimplemented, "not implemented")
}
