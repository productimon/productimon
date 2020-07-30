package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"sync"

	"git.yiad.am/productimon/aggregator/deviceState"
	"git.yiad.am/productimon/aggregator/notifications"
	cpb "git.yiad.am/productimon/proto/common"
	spb "git.yiad.am/productimon/proto/svc"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const bcryptStrength = 12

type service struct {
	auther    *Authenticator
	dbWLock   *sync.Mutex
	db        *sql.DB
	log       *zap.Logger
	notifiers map[string]notifications.Notifier

	ds *deviceState.DsMap
}

// https://github.com/badoux/checkmail/blob/f9f80cb795fa32891c4f3556822e179796031549/checkmail.go#L37
var rxEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// TODO: what if we recoreded out-of-order events in db and are waiting for an old event when we shutdown server
func (s *service) lazyInitEidHandler(uid string, did int64) (int64, error) {
	var eid int64
	err := s.db.QueryRow("SELECT MAX(id) FROM events WHERE uid=? AND did=?", uid, did).Scan(&eid)
	switch {
	case err == sql.ErrNoRows:
		eid = 0

	case err != nil:
		return -1, err
	}
	return eid, nil
}

func NewService(auther *Authenticator, db *sql.DB, logger *zap.Logger) (s *service) {
	s = &service{
		auther:    auther,
		db:        db,
		dbWLock:   &sync.Mutex{},
		log:       logger,
		notifiers: make(map[string]notifications.Notifier),
	}
	s.ds = deviceState.NewDsMap(s.lazyInitEidHandler, logger)
	return
}

func (s *service) RegisterNotifier(n notifications.Notifier) {
	s.notifiers[n.Name()] = n
}

func (s *service) Notify(kind, recipient, message string) error {
	n := s.notifiers[kind]
	if n == nil {
		return notifications.ErrNotRegistered
	}
	return n.Notify(recipient, message)
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
		s.log.Error("can't sign token", zap.Error(err), zap.String("uid", uid))
		return nil, status.Error(codes.Internal, "something went wrong with signing token")
	}
	return &spb.DataAggregatorLoginResponse{
		Token: token,
	}, nil
}

func (s *service) Login(ctx context.Context, req *spb.DataAggregatorLoginRequest) (*spb.DataAggregatorLoginResponse, error) {
	var uid, storedPassword string
	var verified bool
	err := s.db.QueryRow("SELECT id, password, verified FROM users WHERE email = ? LIMIT 1", req.Email).Scan(&uid, &storedPassword, &verified)
	if err != nil {
		s.log.Debug("error logging in", zap.Error(err), zap.String("email", req.Email))
		return nil, status.Error(codes.Unauthenticated, "invalid email/password")
	}
	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(req.Password))
	if err != nil {
		s.log.Debug("wrong password", zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "invalid email/password")
	}
	if !verified {
		return nil, status.Error(codes.Unauthenticated, "account not verified, please check your email")
	}
	s.log.Info("logged in", zap.String("uid", uid))
	return s.returnToken(ctx, uid)
}

func (s *service) DeviceSignin(ctx context.Context, req *spb.DataAggregatorDeviceSigninRequest) (*spb.DataAggregatorDeviceSigninResponse, error) {
	uid, did, err := s.auther.AuthenticateRequest(ctx)
	if err != nil || did != -1 {
		return nil, status.Error(codes.Unauthenticated, "Invalid token")
	}
	err = s.db.QueryRow("SELECT MAX(id) FROM devices WHERE uid=?", uid).Scan(&did)
	if err != nil {
		s.log.Debug("DeviceSignin: MAX(id) FROM devices failed", zap.Error(err))
		did = 0
	} else {
		did += 1
	}
	s.log.Info("DeviceSignin: signing cert", zap.String("uid", uid), zap.Int64("did", did))
	s.dbWLock.Lock()
	defer s.dbWLock.Unlock()
	_, err = s.db.Exec("INSERT INTO devices(uid, id, name, kind) VALUES(?, ?, ?, ?)", uid, did, req.Device.Name, req.Device.DeviceType)
	if err != nil {
		s.log.Error("can't insert device", zap.Error(err), zap.String("uid", uid), zap.Int64("did", did), zap.String("device_name", req.Device.Name))
		return nil, status.Error(codes.Internal, "something went wrong")
	}
	cert, key, err := s.auther.SignDeviceCert(uid, did)
	if err != nil {
		s.log.Error("can't sign device cert", zap.Error(err), zap.String("uid", uid), zap.Int64("did", did), zap.String("device_name", req.Device.Name))
		return nil, status.Error(codes.Internal, "something went wrong")
	}
	return &spb.DataAggregatorDeviceSigninResponse{
		Cert: cert,
		Key:  key,
	}, nil
}

func (s *service) Signup(ctx context.Context, req *spb.DataAggregatorSignupRequest) (*spb.DataAggregatorLoginResponse, error) {
	if len(req.User.Email) > 254 || !rxEmail.MatchString(req.User.Email) {
		return nil, status.Error(codes.InvalidArgument, "invalid email address")
	}
	var tmp int64
	err := s.db.QueryRow("SELECT 1 FROM users WHERE email = ? LIMIT 1", req.User.Email).Scan(&tmp)
	switch {
	case err != nil && err != sql.ErrNoRows:
		s.log.Error("error checking user existence for signup", zap.Error(err), zap.String("email", req.User.Email))
		return nil, status.Error(codes.Internal, "something went wrong")
	case err == nil:
		return nil, status.Error(codes.AlreadyExists, "user already exists")
	}
	pwd, err := bcrypt.GenerateFromPassword([]byte(req.User.Password), bcryptStrength)
	if err != nil {
		s.log.Error("error encrypting password", zap.Error(err), zap.String("email", req.User.Email))
		return nil, status.Error(codes.Internal, "something went wrong")
	}
	uid := uuid.New().String()
	s.dbWLock.Lock()
	defer s.dbWLock.Unlock()
	verified := false
	vtoken, err := s.auther.SignVerificationToken(req.User.Email)
	if err != nil {
		s.log.Error("can't sign verification token", zap.Error(err), zap.String("email", req.User.Email))
		return nil, status.Error(codes.Internal, "something went wrong with signing token")
	}
	if err = s.Notify("email", req.User.Email, fmt.Sprintf(
		"Hi there! Verify your productimon email here: http://%s/verify?token=%s", flagDomain, url.QueryEscape(vtoken))); err != nil {
		switch err {
		case notifications.ErrNotRegistered:
			verified = true
		default:
			s.log.Error("error sending verification email", zap.Error(err), zap.String("email", req.User.Email))
			return nil, status.Error(codes.Internal, "something went wrong")
		}
	}
	_, err = s.db.Exec("INSERT INTO users (id, email, password, verified) VALUES (?, ?, ?, ?)", uid, req.User.Email, pwd, verified)
	if err != nil {
		s.log.Error("error inserting user for signup", zap.Error(err), zap.String("uid", uid), zap.String("email", req.User.Email))
		return nil, status.Error(codes.Internal, "something went wrong")
	}
	if !verified {
		// TODO: instead of using an error, have proper proto types for this
		return nil, status.Error(codes.Internal, "Please check your email and click the verification link!")
	}
	return s.returnToken(ctx, uid)
}

func (s *service) ExtendToken(ctx context.Context, req *cpb.Empty) (*spb.DataAggregatorLoginResponse, error) {
	uid, did, err := s.auther.AuthenticateRequest(ctx)
	if err != nil || did != -1 {
		return nil, status.Error(codes.Unauthenticated, "Invalid token")
	}
	return s.returnToken(ctx, uid)
}

func (s *service) UserDetails(ctx context.Context, req *cpb.Empty) (*spb.DataAggregatorUserDetailsResponse, error) {
	uid, did, err := s.auther.AuthenticateRequest(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Invalid token")
	}
	var email string
	var admin bool
	err = s.db.QueryRow("SELECT email, admin FROM users WHERE id = ? LIMIT 1", uid).Scan(&email, &admin)
	if err != nil {
		return nil, status.Error(codes.Internal, "User missing from db")
	}
	var lastEid int64
	if did != -1 {
		if err = s.db.QueryRow("SELECT max(id) FROM events WHERE uid = ? AND did=?", uid, did).Scan(&lastEid); err != nil {
			s.log.Error("Failed to get last eid", zap.Error(err), zap.String("uid", uid), zap.Int64("did", did))
		}
	}
	ret := &spb.DataAggregatorUserDetailsResponse{
		User: &cpb.User{
			Id:    uid,
			Email: email,
			Admin: admin,
		},
		LastEid: lastEid,
		Device: &cpb.Device{
			Id: did,
		},
	}
	return ret, nil
}

// add a event to events table in a new transaction and return that transaction
// no need to rollback tx if err != nil
func (s *service) addGeneralEvent(uid string, did int64, e *cpb.Event, kind cpb.EventType) (tx *sql.Tx, err error) {
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

func (s *service) eventUpdateState(uid string, did int64, e *cpb.Event, eg func(e *cpb.Event) func(ds *deviceState.DeviceState, db *sql.DB, dblock *sync.Mutex, logger *zap.Logger)) error {
	err := s.ds.RunEvent(s.db, s.dbWLock, s.log, uid, did, e.Id, eg(e))
	if err != nil {
		s.log.Error("RunEvent error", zap.Error(err), zap.String("uid", uid), zap.Int64("did", did), zap.Int64("eid", e.Id))
	}
	return err // this is currently ignored by AddEvent
}

func (s *service) AddEvent(uid string, did int64, e *cpb.Event) error {
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

func (s *service) PushEvent(server spb.DataAggregator_PushEventServer) error {
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

func (s *service) GetEvent(req *spb.DataAggregatorGetEventRequest, server spb.DataAggregator_GetEventServer) error {
	return status.Error(codes.Unimplemented, "not implemented")
}

// TODO
func (s *service) isActive(keystrokes, mouseclicks, starttime, endtime int64) bool {
	return true
}

func (s *service) getLabel(uid, appname string, tx *sql.Tx) (label string) {
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

func (s *service) findActiveTime(uid, dFilter string, stime, etime int64, tx *sql.Tx) int64 {
	rows, err := tx.Query("SELECT I.starttime, I.endtime, A.keystrokes, A.mouseclicks FROM intervals I JOIN activity_events A ON I.id=A.id WHERE I.uid = ? AND I.endtime >= ? AND I.starttime <= ? AND I.kind=?"+dFilter, uid, stime, etime, cpb.EventType_ACTIVITY_EVENT)
	var ret int64
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var st, et, ks, mc int64
			rows.Scan(&st, &et, &ks, &mc)
			if s.isActive(ks, mc, st, et) {
				if st < stime {
					st = stime
				}
				if et > etime {
					et = etime
				}
				ret += et - st
			}
		}
	}
	return ret
}

func (s *service) GetTime(ctx context.Context, req *spb.DataAggregatorGetTimeRequest) (*spb.DataAggregatorGetTimeResponse, error) {
	uid, did, err := s.auther.AuthenticateRequest(ctx)
	if err != nil || did != -1 {
		return nil, status.Error(codes.Unauthenticated, "Invalid token")
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
	s.log.Debug("using device filter", zap.String("dFilter", dFilter))

	intervals := req.GetIntervals()

	var merge func(app string, tottime, acttime int64, result map[string]*spb.DataAggregatorGetTimeResponse_RangeData_DataPoint, tx *sql.Tx)

	switch req.GroupBy {
	case spb.DataAggregatorGetTimeRequest_APPLICATION:
		merge = func(app string, tottime, acttime int64, result map[string]*spb.DataAggregatorGetTimeResponse_RangeData_DataPoint, tx *sql.Tx) {
			dp, ok := result[app]
			if !ok {
				dp = &spb.DataAggregatorGetTimeResponse_RangeData_DataPoint{
					App:   app,
					Label: s.getLabel(uid, app, tx),
				}
				result[app] = dp
			}
			dp.Time += tottime
			dp.Activetime += acttime
		}
	case spb.DataAggregatorGetTimeRequest_LABEL:
		merge = func(app string, tottime, acttime int64, result map[string]*spb.DataAggregatorGetTimeResponse_RangeData_DataPoint, tx *sql.Tx) {
			label := s.getLabel(uid, app, tx)
			dp, ok := result[label]
			if !ok {
				dp = &spb.DataAggregatorGetTimeResponse_RangeData_DataPoint{
					Label: label,
				}
				result[label] = dp
			}
			dp.Time += tottime
			dp.Activetime += acttime
		}
	default:
		return nil, status.Error(codes.InvalidArgument, "i don't recognize that GroupBy param, is earth flat now?")
	}

	rsp := &spb.DataAggregatorGetTimeResponse{}

	for _, in := range intervals {
		rd := &spb.DataAggregatorGetTimeResponse_RangeData{Interval: in}
		s.dbWLock.Lock()
		func() {
			tx, err := s.db.Begin()
			if err != nil {
				s.log.Error("can't begin transaction", zap.Error(err))
				return
			}
			defer tx.Commit()
			rows, err := tx.Query("SELECT MAX(starttime, ?), MIN(endtime, ?), activetime, app FROM intervals WHERE uid = ? AND endtime >= ? AND starttime <= ?"+dFilter, in.Start.Nanos, in.End.Nanos, uid, in.Start.Nanos, in.End.Nanos)

			if err != nil {
				s.log.Error("error querying for GetTime", zap.Error(err))
				return
			}
			result := make(map[string]*spb.DataAggregatorGetTimeResponse_RangeData_DataPoint)
			for rows.Next() {
				var stime, etime, atime int64
				var appname string
				rows.Scan(&stime, &etime, &atime, &appname)
				if stime == in.Start.Nanos || etime == in.End.Nanos {
					atime = s.findActiveTime(uid, dFilter, in.Start.Nanos, in.End.Nanos, tx)
				}
				merge(appname, etime-stime, atime, result, tx)
			}
			rows.Close()
			for _, v := range result {
				rd.Data = append(rd.Data, v)
			}
		}()
		s.dbWLock.Unlock()

		rsp.Data = append(rsp.Data, rd)
	}

	return rsp, nil
}

func (s *service) VerifyAccount(token string) error {
	email, err := s.auther.VerifyVerificationToken(token)
	if err != nil {
		return errors.New("invalid token")
	}
	// NOTE: if account is already verified, we just have a silent failure. Might want to tell user
	// in the future
	if _, err = s.db.Exec("UPDATE users SET verified = TRUE WHERE email = ?", email); err != nil {
		s.log.Error("failed to verify account", zap.Error(err))
		return errors.New("something went wrong verifying your account")
	}
	return nil
}

func (s *service) AddGoal(ctx context.Context, goal *cpb.Goal) (*cpb.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *service) DeleteGoal(ctx context.Context, goal *cpb.Goal) (*cpb.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *service) EditGoal(ctx context.Context, goal *cpb.Goal) (*cpb.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *service) GetGoals(ctx context.Context, req *cpb.Empty) (*spb.DataAggregatorGetGoalsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *service) DeleteAccount(ctx context.Context, req *cpb.Empty) (*cpb.Empty, error) {
	uid, did, err := s.auther.AuthenticateRequest(ctx)
	if err != nil || did != -1 {
		return nil, status.Error(codes.Unauthenticated, "Invalid token")
	}

	s.dbWLock.Lock()
	defer s.dbWLock.Unlock()

	if _, err = s.db.Exec("DELETE FROM users WHERE id = ?", uid); err != nil {
		s.log.Error("failed to delete user", zap.Error(err), zap.String("uid", uid))
		return nil, status.Error(codes.Internal, "error deleting user")
	}

	s.log.Info("deleted user", zap.String("uid", uid))
	return &cpb.Empty{}, nil
}

func (s *service) isAdmin(ctx context.Context) bool {
	uid, did, err := s.auther.AuthenticateRequest(ctx)
	if err != nil || did != -1 {
		return false
	}
	var admin bool
	// we don't want to put admin to JWT because user could be demoted but that token won't expire for a while
	if s.db.QueryRow("SELECT admin FROM users WHERE id = ? LIMIT 1", uid).Scan(&admin); err != nil {
		s.log.Error("failed to get admin status", zap.Error(err), zap.String("uid", uid))
		return false
	}
	return admin
}

func (s *service) PromoteAccount(ctx context.Context, user *cpb.User) (*cpb.Empty, error) {
	if !s.isAdmin(ctx) {
		return nil, status.Error(codes.Unauthenticated, "Unauthorized")
	}
	s.dbWLock.Lock()
	defer s.dbWLock.Unlock()
	// if user is already an admin, we just return success
	if user.Id != "" {
		res, err := s.db.Exec("UPDATE users SET admin = TRUE WHERE id = ?", user.Id)
		if err != nil {
			s.log.Error("failed to promote account", zap.Error(err), zap.String("uid", user.Id))
			return nil, status.Error(codes.Internal, "something went wrong")
		}
		if rows, err := res.RowsAffected(); err != nil || rows == 0 {
			return nil, status.Error(codes.InvalidArgument, "user doesn't exist")
		}
	} else if user.Email != "" {
		res, err := s.db.Exec("UPDATE users SET admin = TRUE WHERE email = ?", user.Email)
		if err != nil {
			s.log.Error("failed to promote account", zap.Error(err), zap.String("email", user.Email))
			return nil, status.Error(codes.Internal, "something went wrong")
		}
		if rows, err := res.RowsAffected(); err != nil || rows == 0 {
			return nil, status.Error(codes.InvalidArgument, "user doesn't exist")
		}
	} else {
		return nil, status.Error(codes.InvalidArgument, "promoted account uid/email both empty")
	}
	return &cpb.Empty{}, nil
}

func (s *service) DemoteAccount(ctx context.Context, user *cpb.User) (*cpb.Empty, error) {
	if !s.isAdmin(ctx) {
		return nil, status.Error(codes.Unauthenticated, "Unauthorized")
	}
	s.dbWLock.Lock()
	defer s.dbWLock.Unlock()
	// if user doesn't exist or isn't an admin, we just don't do anything without returning an error
	if user.Id != "" {
		if _, err := s.db.Exec("UPDATE users SET admin = FALSE WHERE id = ?", user.Id); err != nil {
			s.log.Error("failed to demote account", zap.Error(err), zap.String("uid", user.Id))
			return nil, status.Error(codes.Internal, "something went wrong")
		}
	} else if user.Email != "" {
		if _, err := s.db.Exec("UPDATE users SET admin = FALSE WHERE email = ?", user.Email); err != nil {
			s.log.Error("failed to demote account", zap.Error(err), zap.String("email", user.Email))
			return nil, status.Error(codes.Internal, "something went wrong")
		}
	} else {
		return nil, status.Error(codes.InvalidArgument, "demoted account uid/email both empty")
	}
	return &cpb.Empty{}, nil
}

func (s *service) ListAdmins(ctx context.Context, req *cpb.Empty) (*spb.DataAggregatorListAdminsResponse, error) {
	if !s.isAdmin(ctx) {
		return nil, status.Error(codes.Unauthenticated, "Unauthorized")
	}
	rsp := &spb.DataAggregatorListAdminsResponse{}
	rows, err := s.db.Query("SELECT id, email FROM users WHERE admin = TRUE")
	if err != nil {
		s.log.Error("failed to query admins", zap.Error(err))
		return nil, status.Error(codes.Internal, "something went wrong")
	}
	defer rows.Close()
	for rows.Next() {
		user := &cpb.User{Admin: true}
		if err = rows.Scan(&user.Id, &user.Email); err != nil {
			s.log.Error("failed to query admins", zap.Error(err))
			return nil, status.Error(codes.Internal, "something went wrong")
		}
		rsp.Admins = append(rsp.Admins, user)
	}
	return rsp, nil
}
