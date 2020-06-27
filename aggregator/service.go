package main

import (
	"context"
	"database/sql"
	"log"

	cpb "git.yiad.am/productimon/proto/common"
	spb "git.yiad.am/productimon/proto/svc"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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

func (s *service) ExtendToken(ctx context.Context, req *emptypb.Empty) (*spb.DataAggregatorLoginResponse, error) {
	uid, err := s.auther.AuthenticateRequest(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid token")
	}
	return s.returnToken(ctx, uid)
}

func (s *service) UserDetails(ctx context.Context, req *emptypb.Empty) (*spb.DataAggregatorUserDetailsResponse, error) {
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

func (s *service) PushEvent(server spb.DataAggregator_PushEventServer) error {
	return nil
}

func (s *service) GetEvent(req *spb.DataAggregatorGetEventRequest, server spb.DataAggregator_GetEventServer) error {
	return nil
}
