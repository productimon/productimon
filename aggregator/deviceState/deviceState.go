package deviceState

import (
	"database/sql"
	"fmt"
	"sync"

	cpb "git.yiad.am/productimon/proto/common"
	"go.uber.org/zap"
)

type DeviceState struct {
	uid        string
	did        int64
	app        string
	startTime  int64
	activeTime int64
	running    bool
	evq        OrderedEventQueue
}

type LazyInitEidHandler func(uid string, did int64) (int64, error)

type DsMap struct {
	states         map[string]*DeviceState
	initEidHandler LazyInitEidHandler
	log            *zap.Logger
}

func idsToKey(uid string, did int64) string {
	return fmt.Sprintf("%s-%d", uid, did)
}

func (ds *DeviceState) switchApp(db *sql.DB, dblock *sync.Mutex, log *zap.Logger, app string, timestamp int64) {
	ds.clearState(db, dblock, log, timestamp)
	ds.running = true
	ds.app = app
	ds.startTime = timestamp
	ds.activeTime = 0
}

func (ds *DeviceState) clearState(db *sql.DB, dblock *sync.Mutex, log *zap.Logger, timestamp int64) {
	if ds.running {
		ds.running = false
		dblock.Lock()
		if _, err := db.Exec("INSERT INTO intervals (uid, did, starttime, endtime, activetime, app) VALUES(?, ?, ?, ?, ?, ?)", ds.uid, ds.did, ds.startTime, timestamp, ds.activeTime, ds.app); err != nil {
			log.Error("error in clearState", zap.Error(err))
		}
		dblock.Unlock()
	}
}

func (ds *DeviceState) setActive(db *sql.DB, dblock *sync.Mutex, log *zap.Logger, timestart, timeend int64) {
	ds.activeTime += timeend - timestart
}

func switchApp(app string, timestamp int64) func(ds *DeviceState, db *sql.DB, dblock *sync.Mutex, log *zap.Logger) {
	return func(ds *DeviceState, db *sql.DB, dblock *sync.Mutex, log *zap.Logger) {
		ds.switchApp(db, dblock, log, app, timestamp)
	}
}

func clearState(timestamp int64) func(ds *DeviceState, db *sql.DB, dblock *sync.Mutex, log *zap.Logger) {
	return func(ds *DeviceState, db *sql.DB, dblock *sync.Mutex, log *zap.Logger) {
		ds.clearState(db, dblock, log, timestamp)
	}
}

func setActive(timestart, timeend int64) func(ds *DeviceState, db *sql.DB, dblock *sync.Mutex, log *zap.Logger) {
	return func(ds *DeviceState, db *sql.DB, dblock *sync.Mutex, log *zap.Logger) {
		ds.setActive(db, dblock, log, timestart, timeend)
	}
}

func SwitchApp(e *cpb.Event) func(ds *DeviceState, db *sql.DB, dblock *sync.Mutex, log *zap.Logger) {
	return switchApp(e.GetAppSwitchEvent().AppName, e.Timeinterval.Start.Nanos)
}

func ClearState(e *cpb.Event) func(ds *DeviceState, db *sql.DB, dblock *sync.Mutex, log *zap.Logger) {
	return clearState(e.Timeinterval.Start.Nanos)
}

func SetActive(e *cpb.Event) func(ds *DeviceState, db *sql.DB, dblock *sync.Mutex, log *zap.Logger) {
	return setActive(e.Timeinterval.Start.Nanos, e.Timeinterval.End.Nanos)
}

func Nop(e *cpb.Event) func(ds *DeviceState, db *sql.DB, dblock *sync.Mutex, log *zap.Logger) {
	return func(ds *DeviceState, db *sql.DB, dblock *sync.Mutex, log *zap.Logger) {}
}

func NewDsMap(initEidHandler LazyInitEidHandler, logger *zap.Logger) *DsMap {
	return &DsMap{
		initEidHandler: initEidHandler,
		states:         make(map[string]*DeviceState),
		log:            logger,
	}
}

func (dsm *DsMap) RunEvent(db *sql.DB, dblock *sync.Mutex, log *zap.Logger, uid string, did, eid int64, evf func(ds *DeviceState, db *sql.DB, dblock *sync.Mutex, log *zap.Logger)) error {
	key := idsToKey(uid, did)
	ds, ok := dsm.states[key]
	if !ok {
		initEid, err := dsm.initEidHandler(uid, did)
		if err != nil {
			dsm.log.Error("initEidHandler failed", zap.Error(err))
		}
		// event is added to db before calling this
		// otherwise first event is always discarded
		if eid == initEid {
			initEid -= 1
		}
		dsm.log.Debug("lazy init eid", zap.String("uid", uid), zap.Int64("did", did), zap.Int64("initEid", initEid), zap.Int64("eid", eid))
		ds = &DeviceState{
			uid: uid,
			did: did,
			evq: OrderedEventQueue{lastid: initEid, log: dsm.log},
		}
		dsm.states[key] = ds
	}
	return ds.evq.Push(eid, func() { evf(ds, db, dblock, log) })
}
