package service

import (
	"context"
	"database/sql"

	cpb "git.yiad.am/productimon/proto/common"
	spb "git.yiad.am/productimon/proto/svc"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) findActiveTime(uid, dFilter string, stime, etime int64, tx *sql.Tx) int64 {
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

// TODO
func (s *Service) isActive(keystrokes, mouseclicks, starttime, endtime int64) bool {
	return true
}

func (s *Service) GetTime(ctx context.Context, req *spb.DataAggregatorGetTimeRequest) (*spb.DataAggregatorGetTimeResponse, error) {
	uid, did, err := s.auther.AuthenticateRequest(ctx)
	if err != nil || did != -1 {
		return nil, status.Error(codes.Unauthenticated, "Invalid token")
	}

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

	devices := req.GetDevices()
	dFilter := deviceFilters("did", devices)
	idFilter := deviceFilters("intervals.did", devices)
	s.log.Debug("using device filter", zap.String("dFilter", dFilter))

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
					atime = s.findActiveTime(uid, idFilter, in.Start.Nanos, in.End.Nanos, tx)
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
