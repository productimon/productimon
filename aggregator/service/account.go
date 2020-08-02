package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"regexp"

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

// https://github.com/badoux/checkmail/blob/f9f80cb795fa32891c4f3556822e179796031549/checkmail.go#L37
var rxEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func (s *Service) Login(ctx context.Context, req *spb.DataAggregatorLoginRequest) (*spb.DataAggregatorLoginResponse, error) {
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

func (s *Service) DeviceSignin(ctx context.Context, req *spb.DataAggregatorDeviceSigninRequest) (*spb.DataAggregatorDeviceSigninResponse, error) {
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

func (s *Service) Signup(ctx context.Context, req *spb.DataAggregatorSignupRequest) (*spb.DataAggregatorLoginResponse, error) {
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
		"Hi there! Verify your productimon email here: http://%s/verify?token=%s", s.domain, url.QueryEscape(vtoken))); err != nil {
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

func (s *Service) ExtendToken(ctx context.Context, req *cpb.Empty) (*spb.DataAggregatorLoginResponse, error) {
	uid, did, err := s.auther.AuthenticateRequest(ctx)
	if err != nil || did != -1 {
		return nil, status.Error(codes.Unauthenticated, "Invalid token")
	}
	return s.returnToken(ctx, uid)
}

func (s *Service) UserDetails(ctx context.Context, req *cpb.Empty) (*spb.DataAggregatorUserDetailsResponse, error) {
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
		if err = s.db.QueryRow("SELECT COALESCE(MAX(id), 0) FROM events WHERE uid = ? AND did=?", uid, did).Scan(&lastEid); err != nil {
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

func (s *Service) returnToken(ctx context.Context, uid string) (*spb.DataAggregatorLoginResponse, error) {
	token, err := s.auther.SignToken(uid)
	if err != nil {
		s.log.Error("can't sign token", zap.Error(err), zap.String("uid", uid))
		return nil, status.Error(codes.Internal, "something went wrong with signing token")
	}

	var admin bool
	var email string
	if err := s.db.QueryRow("SELECT email, admin FROM users WHERE id = ? LIMIT 1", uid).Scan(&email, &admin); err != nil {
		return nil, status.Error(codes.Internal, "something went wrong")
	}
	return &spb.DataAggregatorLoginResponse{
		Token: token,
		User: &cpb.User{
			Id:    uid,
			Email: email,
			Admin: admin,
		},
	}, nil
}

func (s *Service) VerifyAccount(token string) error {
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

func (s *Service) DeleteAccount(ctx context.Context, req *cpb.Empty) (*cpb.Empty, error) {
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
