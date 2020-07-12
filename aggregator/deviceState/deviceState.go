package deviceState

import (
	"database/sql"
	"fmt"
	"log"

	cpb "git.yiad.am/productimon/proto/common"
)

type DeviceState struct {
	uid        string
	did        int
	app        string
	startTime  int64
	activeTime int64
	running    bool
	evq        OrderedEventQueue
}

type LazyInitEidHandler func(uid string, did int) (int64, error)

type DsMap struct {
	states         map[string]*DeviceState
	initEidHandler LazyInitEidHandler
}

func idsToKey(uid string, did int) string {
	return fmt.Sprintf("%s-%d", uid, did)
}

func (ds *DeviceState) switchApp(db *sql.DB, app string, timestamp int64) {
	ds.clearState(db, timestamp)
	ds.running = true
	ds.app = app
	ds.startTime = timestamp
	ds.activeTime = 0
}

func (ds *DeviceState) clearState(db *sql.DB, timestamp int64) {
	if ds.running {
		ds.running = false
		if _, err := db.Exec("INSERT INTO intervals (uid, did, starttime, endtime, activetime, app) VALUES(?, ?, ?, ?, ?, ?)", ds.uid, ds.did, ds.startTime, timestamp, ds.activeTime, ds.app); err != nil {
			log.Printf("error in clearState: %v", err)
		}
	}
}

func (ds *DeviceState) setActive(db *sql.DB, timestart, timeend int64) {
	ds.activeTime += timeend - timestart
}

func switchApp(app string, timestamp int64) func(ds *DeviceState, db *sql.DB) {
	return func(ds *DeviceState, db *sql.DB) {
		ds.switchApp(db, app, timestamp)
	}
}

func clearState(timestamp int64) func(ds *DeviceState, db *sql.DB) {
	return func(ds *DeviceState, db *sql.DB) {
		ds.clearState(db, timestamp)
	}
}

func setActive(timestart, timeend int64) func(ds *DeviceState, db *sql.DB) {
	return func(ds *DeviceState, db *sql.DB) {
		ds.setActive(db, timestart, timeend)
	}
}

func SwitchApp(e *cpb.Event) func(ds *DeviceState, db *sql.DB) {
	return switchApp(e.GetAppSwitchEvent().AppName, e.Timeinterval.Start.Nanos)
}

func ClearState(e *cpb.Event) func(ds *DeviceState, db *sql.DB) {
	return clearState(e.Timeinterval.Start.Nanos)
}

func SetActive(e *cpb.Event) func(ds *DeviceState, db *sql.DB) {
	return setActive(e.Timeinterval.Start.Nanos, e.Timeinterval.End.Nanos)
}

func Nop(e *cpb.Event) func(ds *DeviceState, db *sql.DB) {
	return func(ds *DeviceState, db *sql.DB) {}
}

func NewDsMap(initEidHandler LazyInitEidHandler) *DsMap {
	return &DsMap{
		initEidHandler: initEidHandler,
		states:         make(map[string]*DeviceState),
	}
}

func (dsm *DsMap) RunEvent(db *sql.DB, uid string, did int, eid int64, evf func(ds *DeviceState, db *sql.DB)) error {
	key := idsToKey(uid, did)
	ds, ok := dsm.states[key]
	if !ok {
		initEid, err := dsm.initEidHandler(uid, did)
		if err != nil {
			return err
		}
		ds = &DeviceState{
			uid: uid,
			did: did,
			evq: OrderedEventQueue{lastid: initEid},
		}
		dsm.states[key] = ds
	}
	return ds.evq.Push(eid, func() { evf(ds, db) })
}
