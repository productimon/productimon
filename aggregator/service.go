package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"

	"git.yiad.am/productimon/aggregator/deviceState"
	cpb "git.yiad.am/productimon/proto/common"
	spb "git.yiad.am/productimon/proto/svc"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const bcryptStrength = 12

type service struct {
	auther *Authenticator
	db     *sql.DB

	ds *deviceState.DsMap
}

// TODO: deal with ds recovery upon server rebooting
func (s *service) lazyInitEidHandler(uid string, did int) (int64, error) {
	return 0, nil
}

func NewService(auther *Authenticator, db *sql.DB) (s *service) {
	s = &service{
		auther: auther,
		db:     db,
	}
	s.ds = deviceState.NewDsMap(s.lazyInitEidHandler)
	return
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
	err := s.db.QueryRow("SELECT id, password FROM users WHERE email = ? LIMIT 1", req.Email).Scan(&uid, &storedPassword)
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

func (s *service) Signup(ctx context.Context, req *spb.DataAggregatorSignupRequest) (*spb.DataAggregatorLoginResponse, error) {
	var tmp int64
	err := s.db.QueryRow("SELECT 1 FROM users WHERE email = ? LIMIT 1", req.User.Email).Scan(&tmp)
	switch {
	case err != nil && err != sql.ErrNoRows:
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "something went wrong")
	case err == nil:
		return nil, status.Errorf(codes.AlreadyExists, "user already exists")
	}
	pwd, err := bcrypt.GenerateFromPassword([]byte(req.User.Password), bcryptStrength)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "something went wrong")
	}
	uid := uuid.New().String()
	_, err = s.db.Exec("INSERT INTO users (id, email, password) VALUES (?, ?, ?)", uid, req.User.Email, pwd)
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "something went wrong")
	}
	// TODO: have separate api to create device. This should be under one transaction but we will remove this eventually so
	// we don't have a transaction here
	s.db.Exec("INSERT INTO devices VALUES(?, 0, 'test device (Linux)', 1)", uid)
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
	err = s.db.QueryRow("SELECT email FROM users WHERE id = ? LIMIT 1", uid).Scan(&email)
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

func (s *service) eventUpdateState(uid string, did int, e *cpb.Event, eg func(e *cpb.Event) func(ds *deviceState.DeviceState, db *sql.DB)) error {
	err := s.ds.RunEvent(s.db, uid, did, e.Id, eg(e))
	if err != nil {
		log.Println(err)
	}
	return err // this is currently ignored by AddEvent
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
		s.eventUpdateState(uid, did, e, deviceState.SwitchApp)

	case *cpb.Event_StartTrackingEvent:
		tx, err := s.addGeneralEvent(uid, did, e, cpb.EventType_START_TRACKING_EVENT)
		if err != nil {
			return err
		}
		tx.Commit()
		s.eventUpdateState(uid, did, e, deviceState.ClearState)

	case *cpb.Event_StopTrackingEvent:
		tx, err := s.addGeneralEvent(uid, did, e, cpb.EventType_STOP_TRACKING_EVENT)
		if err != nil {
			return err
		}
		tx.Commit()
		s.eventUpdateState(uid, did, e, deviceState.ClearState)

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
		s.eventUpdateState(uid, did, e, deviceState.SetActive)

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
		s.eventUpdateState(uid, did, e, deviceState.SetActive)

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

	for _, in := range intervals {
		rd := &spb.DataAggregatorGetTimeResponse_RangeData{Interval: in}
		rows, err := s.db.Query("SELECT MAX(starttime, ?), MIN(endtime, ?), app FROM intervals WHERE uid = ? AND endtime >= ? AND starttime <= ?"+dFilter, in.Start.Nanos, in.End.Nanos, uid, in.Start.Nanos, in.End.Nanos)

		if err == nil {
			data := make(map[string]int64)
			for rows.Next() {
				var stime, etime int64
				var appname string
				rows.Scan(&stime, &etime, &appname)
				data[transform(appname)] += etime - stime
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
		} else {
			log.Println(err)
		}

		rsp.Data = append(rsp.Data, rd)
	}

	return rsp, nil
}
