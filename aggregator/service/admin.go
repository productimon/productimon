package service

import (
	"context"

	cpb "git.yiad.am/productimon/proto/common"
	spb "git.yiad.am/productimon/proto/svc"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) isAdmin(ctx context.Context) bool {
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

func (s *Service) isRootAdmin(ctx context.Context) bool {
	uid, did, err := s.auther.AuthenticateRequest(ctx)
	if err != nil || did != -1 {
		return false
	}
	var admin bool
	var email string
	if s.db.QueryRow("SELECT email, admin FROM users WHERE id = ? LIMIT 1", uid).Scan(&email, &admin); err != nil {
		s.log.Error("failed to get admin status", zap.Error(err), zap.String("uid", uid))
		return false
	}
	return admin && email == flagFirstUser
}

func (s *Service) PromoteAccount(ctx context.Context, user *cpb.User) (*cpb.Empty, error) {
	if !s.isRootAdmin(ctx) {
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

func (s *Service) DemoteAccount(ctx context.Context, user *cpb.User) (*cpb.Empty, error) {
	if !s.isRootAdmin(ctx) {
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
		if user.Email == flagFirstUser {
			return nil, status.Error(codes.Internal, "Cannot demote root admin")
		}
		if _, err := s.db.Exec("UPDATE users SET admin = FALSE WHERE email = ?", user.Email); err != nil {
			s.log.Error("failed to demote account", zap.Error(err), zap.String("email", user.Email))
			return nil, status.Error(codes.Internal, "something went wrong")
		}
	} else {
		return nil, status.Error(codes.InvalidArgument, "demoted account uid/email both empty")
	}
	return &cpb.Empty{}, nil
}

func (s *Service) ListAdmins(ctx context.Context, req *cpb.Empty) (*spb.DataAggregatorListAdminsResponse, error) {
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
