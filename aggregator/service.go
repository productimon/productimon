package main

import (
	"context"

	spb "git.yiad.am/productimon/proto/svc"
)

type service struct{}

func (s *service) Ping(ctx context.Context, req *spb.DataAggregatorPingRequest) (*spb.DataAggregatorPingResponse, error) {
	rsp := &spb.DataAggregatorPingResponse{
		Payload: req.Payload,
	}
	return rsp, nil
}
func (s *service) PushEvent(server spb.DataAggregator_PushEventServer) error {
	return nil
}
func (s *service) GetEvent(req *spb.DataAggregatorGetEventRequest, server spb.DataAggregator_GetEventServer) error {
	return nil
}
