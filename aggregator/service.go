package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"

	cpb "git.yiad.am/productimon/proto/common"
	spb "git.yiad.am/productimon/proto/svc"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type service struct {
	auther *Authenticator
	db     *sql.DB
}

func (s *service) Ping(ctx context.Context, req *spb.DataAggregatorPingRequest) (*spb.DataAggregatorPingResponse, error) {
	rsp := &spb.DataAggregatorPingResponse{
		Payload: req.Payload,
	}
	return rsp, nil
}

func (s *service) returnToken(ctx context.Context, uid string) (*spb.DataAggregatorLoginResponse, error) {
	token, err := s.auther.SignToken(uid)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "something went wrong with signing token")
	}
	return &spb.DataAggregatorLoginResponse{
		Token: token,
	}, nil
}

func (s *service) Login(ctx context.Context, req *spb.DataAggregatorLoginRequest) (*spb.DataAggregatorLoginResponse, error) {
	var uid, storedPassword string
	err := s.db.QueryRow("SELECT id, password FROM users WHERE email = ?", req.Email).Scan(&uid, &storedPassword)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Unauthenticated, "invalid email/password")
	}
	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(req.Password))
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Unauthenticated, "invalid email/password")
	}
	return s.returnToken(ctx, uid)
}

func (s *service) ExtendToken(ctx context.Context, req *cpb.Empty) (*spb.DataAggregatorLoginResponse, error) {
	uid, err := s.auther.AuthenticateRequest(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid token")
	}
	return s.returnToken(ctx, uid)
}

func (s *service) UserDetails(ctx context.Context, req *cpb.Empty) (*spb.DataAggregatorUserDetailsResponse, error) {
	uid, err := s.auther.AuthenticateRequest(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid token")
	}
	var email string
	err = s.db.QueryRow("SELECT email FROM users WHERE id = ?", uid).Scan(&email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "User missing from db")
	}
	ret := &spb.DataAggregatorUserDetailsResponse{
		User: &cpb.User{
			Id:    uid,
			Email: email,
		},
	}
	return ret, nil
}

// add a event to events table in a new transaction and return that transaction
// no need to rollback tx if err != nil
func (s *service) addGeneralEvent(uid string, did int, e *cpb.Event, kind cpb.EventType) (tx *sql.Tx, err error) {
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

func (s *service) AddEvent(uid string, did int, e *cpb.Event) error {
	switch k := e.Kind.(type) {
	case *cpb.Event_AppSwitchEvent:
		tx, err := s.addGeneralEvent(uid, did, e, cpb.EventType_APP_SWITCH_EVENT)
		if err != nil {
			return err
		}
		if _, err := tx.Exec("INSERT INTO app_switch_events(uid, did, id, app) VALUES(?, ?, ?, ?)",
			uid, did, e.Id, k.AppSwitchEvent.AppName); err != nil {
			tx.Rollback()
			return err
		}
		tx.Commit()
	case *cpb.Event_StartTrackingEvent:
		tx, err := s.addGeneralEvent(uid, did, e, cpb.EventType_START_TRACKING_EVENT)
		if err != nil {
			return err
		}
		tx.Commit()
	case *cpb.Event_StopTrackingEvent:
		tx, err := s.addGeneralEvent(uid, did, e, cpb.EventType_STOP_TRACKING_EVENT)
		if err != nil {
			return err
		}
		tx.Commit()
	case *cpb.Event_KeyStrokeEvent:
		tx, err := s.addGeneralEvent(uid, did, e, cpb.EventType_KEY_STROKE_EVENT)
		if err != nil {
			return err
		}
		if _, err := tx.Exec("INSERT INTO key_stroke_events(uid, did, id, keystrokes) VALUES(?, ?, ?, ?)",
			uid, did, e.Id, k.KeyStrokeEvent.Keystrokes); err != nil {
			tx.Rollback()
			return err
		}
		tx.Commit()
	case *cpb.Event_MouseClickEvent:
		tx, err := s.addGeneralEvent(uid, did, e, cpb.EventType_MOUSE_CLICK_EVENT)
		if err != nil {
			return err
		}
		if _, err := tx.Exec("INSERT INTO mouse_click_events(uid, did, id, mouseclicks) VALUES(?, ?, ?, ?)",
			uid, did, e.Id, k.MouseClickEvent.Mouseclicks); err != nil {
			tx.Rollback()
			return err
		}
		tx.Commit()
	case nil:
		return errors.New("event not set")
	default:
		return fmt.Errorf("unknown event type %T", k)
	}

	return nil
}

func (s *service) PushEvent(server spb.DataAggregator_PushEventServer) error {
	log.Println("Started pushEvent stream")
	ctx := server.Context()
	uid, err := s.auther.AuthenticateRequest(ctx)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "Invalid token")
	}
	did := 0 // TODO: fill in device id

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		event, err := server.Recv()
		switch {
		case err == io.EOF:
			log.Println("Client closed the stream, we're closing too")
			return nil
		case err != nil:
			log.Printf("receive error %v", err)
			continue
		}
		log.Printf("received event %v", event)

		if err = s.AddEvent(uid, did, event); err != nil {
			log.Printf("error adding event: %v", err)
		}
	}
	return nil
}

func (s *service) GetEvent(req *spb.DataAggregatorGetEventRequest, server spb.DataAggregator_GetEventServer) error {
	return nil
}

func (s *service) getLabel(appname string) string {
	// TODO: this needs to be fancy
	return appname
}

func (s *service) GetTime(ctx context.Context, req *spb.DataAggregatorGetTimeRequest) (*spb.DataAggregatorGetTimeResponse, error) {
	uid, err := s.auther.AuthenticateRequest(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid token")
	}
	devices := req.GetDevices()
	dFilter := ""
	if len(devices) > 0 {
		dFilter = " AND events.did IN ("
		pfx := ""
		for _, dev := range devices {
			dFilter += fmt.Sprintf("%s%d", pfx, dev.Id)
			pfx = ", "
		}
		dFilter += ")"
	}
	log.Println(dFilter)

	intervals := req.GetIntervals()

	var transform func(string) string

	switch req.GroupBy {
	case spb.DataAggregatorGetTimeRequest_APPLICATION:
		transform = func(x string) string { return x }
	case spb.DataAggregatorGetTimeRequest_LABEL:
		transform = func(x string) string { return s.getLabel(x) }
	default:
		return nil, status.Errorf(codes.InvalidArgument, "i don't recognize that GroupBy param, is earth flat now?")
	}

	rsp := &spb.DataAggregatorGetTimeResponse{}

	// TODO: optimize this terribly written code written at 3AM
	//       that does not consider performance whatsoever
	for _, in := range intervals {
		rd := &spb.DataAggregatorGetTimeResponse_RangeData{Interval: in}
		func() {
			var sid, eid int64
			if err := s.db.QueryRow("SELECT MIN(id), MAX(id) FROM events WHERE starttime >= ? AND endtime <= ? AND uid = ?"+dFilter, in.Start.Nanos, in.End.Nanos, uid).Scan(&sid, &eid); err != nil {
				fmt.Println(err)
				return
			}
			rows, err := s.db.Query("SELECT events.id, events.kind, events.starttime, app_switch_events.app FROM events LEFT JOIN app_switch_events ON app_switch_events.id = events.id AND app_switch_events.uid = events.uid AND app_switch_events.did = events.did WHERE events.id >= ? AND events.id <= ? AND events.uid = ? AND events.kind IN (?,?,?)"+dFilter, sid-1, eid, uid, cpb.EventType_APP_SWITCH_EVENT, cpb.EventType_START_TRACKING_EVENT, cpb.EventType_STOP_TRACKING_EVENT)
			if err != nil {
				fmt.Println(err)
				return
			}
			currApp := ""
			lastTime := in.Start.Nanos
			data := make(map[string]int64)
			for rows.Next() {
				var eid, ekind, etime int64
				var appname string
				// for all event types we are interested, starttime=endtime
				rows.Scan(&eid, &ekind, &etime, &appname)
				if etime > lastTime {
					if currApp != "" {
						data[currApp] += etime - lastTime
						log.Printf("%s - %d - %d [%d ns]", currApp, lastTime, etime, etime-lastTime)
					}
					lastTime = etime
				}
				if appname != "" {
					currApp = transform(appname)
				} else {
					currApp = ""
				}
			}
			for k, v := range data {
				switch req.GroupBy {
				case spb.DataAggregatorGetTimeRequest_APPLICATION:
					rd.Data = append(rd.Data, &spb.DataAggregatorGetTimeResponse_RangeData_DataPoint{
						App:   k,
						Label: s.getLabel(k),
						Time:  v,
					})
				case spb.DataAggregatorGetTimeRequest_LABEL:
					rd.Data = append(rd.Data, &spb.DataAggregatorGetTimeResponse_RangeData_DataPoint{
						Label: k,
						Time:  v,
					})
				}
			}
		}()
		rsp.Data = append(rsp.Data, rd)
	}

	return rsp, nil
}
