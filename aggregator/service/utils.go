package service

import (
	"fmt"

	cpb "git.yiad.am/productimon/proto/common"
)

func deviceFilters(fieldname string, devices []*cpb.Device) string {
	dFilter := ""
	if len(devices) > 0 {
		dFilter = " AND " + fieldname + " IN ("
		pfx := ""
		for _, dev := range devices {
			// this works because device ID is integer
			// so this is not prone to injection
			dFilter += fmt.Sprintf("%s%d", pfx, dev.Id)
			pfx = ", "
		}
		dFilter += ")"
	}
	return dFilter
}
