package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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
		st += "AND COALESCE(user_apps.label, default_apps.label, '" + LABEL_UNCATEGORIZED + "') = ? "
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

// calculate and update goal progress
func (s *Service) UpdateGoal(uid string, gid int64) {
	var isLabel bool
	var item string
	var baseDuration, targetDuration, startTime, endTime, oldProgress, progress int64
	var err error
	s.dbWLock.Lock()
	defer s.dbWLock.Unlock()
	if err = s.db.QueryRow("SELECT is_label, item, base_duration, target_duration, starttime, endtime, progress FROM goals WHERE uid = ? AND id = ?", uid, gid).Scan(&isLabel, &item, &baseDuration, &targetDuration, &startTime, &endTime, &oldProgress); err != nil {
		s.log.Error("Error updating goal", zap.Error(err), zap.String("uid", uid), zap.Int64("gid", gid))
		return
	}
	if progress, err = s.getGoalProgress(uid, "" /* TODO: get device filter */, isLabel, item, baseDuration, targetDuration, startTime, endTime); err != nil {
		s.log.Error("error getting goal progress", zap.Error(err), zap.String("uid", uid), zap.Int64("gid", gid))
		return
	}
	if _, err = s.db.Exec("UPDATE goals SET progress = ? WHERE uid = ? AND id = ?", progress, uid, gid); err != nil {
		s.log.Error("error setting goal progress", zap.Error(err), zap.String("uid", uid), zap.Int64("gid", gid))
		return
	}
	// TODO: differentiate aspiring/limiting goals, and make the message fancier
	var msg string
	switch {
	// TODO: this looks like a nice place to allow configuration/templating. If not on a user-level, at least on a server-level
	case oldProgress < 1000 && progress >= 1000:
		msg = fmt.Sprintf("Congrats! You've achieved your goal in using %s from %s to %s. Check %s for more details.", item, time.Unix(0, startTime).String(), time.Unix(0, endTime).String(), s.domain)
	case oldProgress < 850 && progress >= 850:
		msg = fmt.Sprintf("You're almost there! You've finished %.1f%% of your goal in using %s from %s to %s. Check %s for more details.", float32(progress)/10, item, time.Unix(0, startTime).String(), time.Unix(0, endTime).String(), s.domain)
	}
	// TODO: add non-email notifiers
	if len(msg) > 0 {
		var email string
		if err = s.db.QueryRow("SELECT email FROM users WHERE id = ? LIMIT 1", uid).Scan(&email); err != nil {
			s.log.Error("error getting email", zap.Error(err), zap.String("uid", uid))
			return
		}
		if err = s.Notify("email", email, msg); err != nil {
			s.log.Error("error sending goal notification email", zap.Error(err), zap.String("email", email))
			return
		}
	}
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
	// TODO: store devices to db
	if _, err = s.db.Exec("INSERT INTO goals VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ",
		goal.Uid, goal.Id, isLabel, item, isPercent, goalDuration, targetDuration, baseDuration, goal.GoalInterval.Start.Nanos, goal.GoalInterval.End.Nanos, goal.GetCompareInterval().GetStart().GetNanos(), goal.GetCompareInterval().GetEnd().GetNanos(), goal.DaysOfWeek, goal.CompareEqualized, progress); err != nil {
		s.log.Error("insert goal failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "error adding goal")
	}
	return &cpb.Empty{}, nil
}

func (s *Service) DeleteGoal(ctx context.Context, goal *cpb.Goal) (*cpb.Empty, error) {
	uid, did, err := s.auther.AuthenticateRequest(ctx)
	if err != nil || did != -1 {
		return nil, status.Error(codes.Unauthenticated, "Invalid token")
	}
	// once again we're happily settled with silent failure if id doesn't exist
	_, err = s.db.Exec("DELETE FROM goals WHERE uid = ? AND id = ?", uid, goal.Id)
	if err != nil {
		s.log.Error("delete goal failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "something went wrong")
	}
	return &cpb.Empty{}, nil
}

func (s *Service) EditGoal(ctx context.Context, goal *cpb.Goal) (*cpb.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (s *Service) GetGoals(ctx context.Context, req *cpb.Empty) (*spb.DataAggregatorGetGoalsResponse, error) {
	uid, did, err := s.auther.AuthenticateRequest(ctx)
	if err != nil || did != -1 {
		return nil, status.Error(codes.Unauthenticated, "Invalid token")
	}

	rows, err := s.db.Query("SELECT id, is_label, item, is_percent, goal_duration, starttime, endtime, compare_starttime, compare_endtime, equalized, progress FROM goals WHERE uid = ?", uid)

	rsp := &spb.DataAggregatorGetGoalsResponse{}
	switch {
	case err == nil:
		defer rows.Close()
		for rows.Next() {
			var id, goalDuration, starttime, endtime, compareStarttime, compareEndtime, progress int64
			var isLabel, isPercent, equalized bool
			var item string
			if err = rows.Scan(&id, &isLabel, &item, &isPercent, &goalDuration, &starttime, &endtime, &compareStarttime, &compareEndtime, &equalized, &progress); err != nil {
				s.log.Error("failed to scan goal", zap.Error(err))
				continue
			}
			goal := &cpb.Goal{
				Uid: uid,
				Id:  id,
				GoalInterval: &cpb.Interval{
					Start: &cpb.Timestamp{Nanos: starttime},
					End:   &cpb.Timestamp{Nanos: endtime},
				},
				CompareInterval: &cpb.Interval{
					Start: &cpb.Timestamp{Nanos: compareStarttime},
					End:   &cpb.Timestamp{Nanos: compareEndtime},
				},
				CompareEqualized: equalized,
				Completed:        endtime < time.Now().UnixNano(),
				Progress:         float32(progress) / 1000,
			}
			if isPercent {
				goal.Amount = &cpb.Goal_PercentAmount{PercentAmount: float32(goalDuration) / 1000}
			} else {
				goal.Amount = &cpb.Goal_FixedAmount{FixedAmount: goalDuration}
			}
			if isLabel {
				goal.Item = &cpb.Goal_Label{Label: item}
			} else {
				goal.Item = &cpb.Goal_Application{Application: item}
			}
			rsp.Goals = append(rsp.Goals, goal)
		}
	case err == sql.ErrNoRows:
	default:
		s.log.Error("Failed to get goals", zap.Error(err))
		return nil, status.Error(codes.Internal, "something went wrong")
	}

	return rsp, nil
}
