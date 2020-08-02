package service

import (
	"context"
	"errors"

	cpb "git.yiad.am/productimon/proto/common"
	spb "git.yiad.am/productimon/proto/svc"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// get total time for user in interval for a given label/app
func (s *Service) getGoalDuration(uid string, isLabel bool, item, dFilter string, startTime, endTime int64) (duration int64, err error) {
	st := "SELECT COALESCE(SUM(MIN(endtime, ?)-MAX(starttime, ?)), 0) FROM intervals "
	defer func() {
		s.log.Debug("getGoalDuration", zap.String("sql", st), zap.Int64("duration", duration), zap.Error(err))
	}()
	if isLabel {
		st += "LEFT JOIN user_apps ON (user_apps.name = intervals.app AND user_apps.uid = intervals.uid) "
		st += "LEFT JOIN default_apps ON (user_apps.name IS NULL AND default_apps.name = intervals.app) "
	}
	st += "WHERE intervals.uid = ? AND endtime >= ? AND starttime <= ? "
	if isLabel {
		st += "AND COALESCE(user_apps.label, default_apps.label, 'Unknown') = ? "
	} else {
		st += "AND app = ? "
	}
	st += dFilter
	if err = s.db.QueryRow(st, endTime, startTime, uid, startTime, endTime, item).Scan(&duration); err != nil {
		return 0, err
	}
	return
}

// precompute goal
func (s *Service) initGoal(g *cpb.Goal) (isLabel bool, item string, isPercent bool, goalDuration, targetDuration, baseDuration int64, err error) {
	switch i := g.Item.(type) {
	case *cpb.Goal_Label:
		isLabel = true
		item = i.Label
	case *cpb.Goal_Application:
		isLabel = false
		item = i.Application
	}
	switch a := g.Amount.(type) {
	case *cpb.Goal_PercentAmount:
		isPercent = true
		if a.PercentAmount < -1 {
			goalDuration = -1000
		} else {
			goalDuration = int64(a.PercentAmount * 1000)
		}
	case *cpb.Goal_FixedAmount:
		isPercent = false
		goalDuration = a.FixedAmount
	}
	if g.CompareInterval == nil {
		if isPercent {
			err = errors.New("cannot set percent goal when compare interval is not set")
			return
		}
		baseDuration = 0
		targetDuration = goalDuration
		return
	}
	if g.CompareInterval.End.Nanos <= g.CompareInterval.Start.Nanos {
		err = errors.New("invalid compare interval")
		return
	}
	if baseDuration, err = s.getGoalDuration(g.Uid, isLabel, item, deviceFilters("intervals.did", g.GetDevices()), g.CompareInterval.Start.Nanos, g.CompareInterval.End.Nanos); err != nil {
		return
	}
	if g.CompareEqualized {
		ratio := float64(g.GoalInterval.End.Nanos-g.GoalInterval.Start.Nanos) / float64(g.CompareInterval.End.Nanos-g.CompareInterval.Start.Nanos)
		baseDuration = int64(float64(baseDuration) * ratio)
	}
	if isPercent {
		targetDuration = int64((1 + float32(goalDuration)/1000) * float32(baseDuration))
	} else {
		// NOTE(adamyi@): if equalized, i'm assuming goalDuration doesn't get scaled. only baseDuration gets scaled
		// please change this if this is not desired
		targetDuration = baseDuration + goalDuration
	}
	return
}

func (s *Service) getGoalProgress(uid, dFilter string, isLabel bool, item string, baseDuration, targetDuration, startTime, endTime int64) (int64, error) {
	actualDuration, err := s.getGoalDuration(uid, isLabel, item, dFilter, startTime, endTime)
	if err != nil {
		return 0, err
	}
	result := float64(actualDuration-baseDuration) / float64(targetDuration-baseDuration)
	if result < 0 {
		result = 0
	}
	if result > 1 {
		result = 1
	}
	return int64(result * 1000), nil
}

func (s *Service) AddGoal(ctx context.Context, goal *cpb.Goal) (*cpb.Empty, error) {
	uid, did, err := s.auther.AuthenticateRequest(ctx)
	if err != nil || did != -1 {
		return nil, status.Error(codes.Unauthenticated, "Invalid token")
	}
	goal.Uid = uid
	s.dbWLock.Lock()
	defer s.dbWLock.Unlock()
	if err = s.db.QueryRow("SELECT COALESCE(MAX(id), -1) FROM goals WHERE uid=?", uid).Scan(&goal.Id); err != nil {
		s.log.Error("error getting goal id", zap.String("uid", uid), zap.Error(err))
		return nil, status.Error(codes.Internal, "error adding goal")
	}
	goal.Id = goal.Id + 1
	isLabel, item, isPercent, goalDuration, targetDuration, baseDuration, err := s.initGoal(goal)
	if err != nil {
		s.log.Error("init goal error", zap.Error(err))
		return nil, status.Error(codes.Internal, "error adding goal")
	}
	progress, err := s.getGoalProgress(uid, deviceFilters("did", goal.GetDevices()), isLabel, item, baseDuration, targetDuration, goal.GoalInterval.Start.Nanos, goal.GoalInterval.End.Nanos)
	if err != nil {
		s.log.Error("error getting goal progress", zap.Error(err))
		return nil, status.Error(codes.Internal, "error adding goal")
	}
	if _, err = s.db.Exec("INSERT INTO goals VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ",
		goal.Uid, goal.Id, isLabel, item, isPercent, goalDuration, targetDuration, baseDuration, goal.GoalInterval.Start.Nanos, goal.GoalInterval.End.Nanos, goal.GetCompareInterval().GetStart().GetNanos(), goal.GetCompareInterval().GetEnd().GetNanos(), goal.DaysOfWeek, goal.CompareEqualized, progress); err != nil {
		s.log.Error("insert goal failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "error adding goal")
	}
	return &cpb.Empty{}, nil
}

func (s *Service) DeleteGoal(ctx context.Context, goal *cpb.Goal) (*cpb.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *Service) EditGoal(ctx context.Context, goal *cpb.Goal) (*cpb.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *Service) GetGoals(ctx context.Context, req *cpb.Empty) (*spb.DataAggregatorGetGoalsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}
