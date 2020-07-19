package reporter

import (
	"time"

	cpb "git.yiad.am/productimon/proto/common"
)

// Returns an interval with starting and ending time to be current timestamp.
func nowInterval() *cpb.Interval {
	ts := time.Now().UnixNano()
	return &cpb.Interval{Start: &cpb.Timestamp{Nanos: ts}, End: &cpb.Timestamp{Nanos: ts}}
}
