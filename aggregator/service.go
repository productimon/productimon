package main

import (
	"context"

	cpb "git.yiad.am/productimon/proto/common"
	spb "git.yiad.am/productimon/proto/svc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type service struct {
	auther *Authenticator
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
		return nil, status.Errorf(codes.Unauthenticated, "invalid email/password")
	}
	return &spb.DataAggregatorLoginResponse{
		Token: token,
	}, nil
}

func (s *service) Login(ctx context.Context, req *spb.DataAggregatorLoginRequest) (*spb.DataAggregatorLoginResponse, error) {
	return s.returnToken(ctx, req.Email) // TODO(adamyi): use uid instead of email and verify password in db
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
	ret := &spb.DataAggregatorUserDetailsResponse{
		// TODO(adamyi): populate with details from db
		User: &cpb.User{
			Id:    uid,
			Email: uid,
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
