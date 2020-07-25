package reporter

import (
	"context"
	"log"
	"sync"

	cpb "git.yiad.am/productimon/proto/common"
	spb "git.yiad.am/productimon/proto/svc"
	"git.yiad.am/productimon/reporter/core/auth"
	"git.yiad.am/productimon/reporter/core/config"
)

const ChannelBufferSize = 4096

// Productimon DataReporter struct
type Reporter struct {
	Config *config.Config

	eq   chan *cpb.Event
	done chan chan bool

	nClicks         int64
	nKeystrokes     int64
	inputStatsMutex sync.Mutex

	eid      int64
	eidMutex sync.Mutex

	inputTrackingDone chan bool
	reportInputStats  chan chan bool

	isTracking bool
	stateMutex sync.RWMutex

	currentApp        string
	windowSwitchMutex sync.Mutex
}

// Create a new Reporter with config
func NewReporter(config *config.Config) *Reporter {
	return &Reporter{
		Config:            config,
		inputTrackingDone: make(chan bool),
		reportInputStats:  make(chan chan bool),
	}
}

// Get next sequential event ID, called when you generate a new event.
func (r *Reporter) getEid() (id int64) {
	r.eidMutex.Lock()
	r.eid++
	id = r.eid
	r.eidMutex.Unlock()
	return
}

// Initiate reporter and go into the main event loop
//
// If it successfully authenticates, a True is sent to the init channel before
// going into the event loop. This blocks the current goroutine and doesn't quit
// until you call r.Quit()
//
// If it fails, a False is sent to init channel and this function returns.
func (r *Reporter) run(init chan bool) {
	log.Printf("productimon core module initiating")

	// establish grpc connection
	// TODO think about what happens if the user started with offline env
	conn, err := auth.ConnectToServer(r.Config.Server, r.Config.Cert())
	if err != nil {
		log.Printf("cannot dial: %v", err)
		init <- false
		return
	}
	client := spb.NewDataAggregatorClient(conn)

	r.eid = r.Config.LastEid

	rsp, err := client.UserDetails(context.Background(), &cpb.Empty{})
	if err != nil {
		log.Printf("Failed to get user details %v", err)
		init <- false
		conn.Close()
		return
	}
	log.Printf("User details: %v", rsp)
	if rsp.LastEid > r.eid {
		r.eid = rsp.LastEid
		log.Printf("Using more recent eid from server: %v", r.eid)
	}

	eventStream, err := client.PushEvent(context.Background())
	if err != nil {
		init <- false
		log.Printf("Cannot get event stream: %v", err)
		conn.Close()
		return
	}

	r.eq = make(chan *cpb.Event, ChannelBufferSize)
	r.done = make(chan chan bool)

	init <- true

	r.eventLoop(conn, client, eventStream)

}

// Returns if the configuration certificate is valid by sending a GetUserDetails
// rpc call.
func (r *Reporter) IsLoggedIn() bool {
	return auth.IsLoggedIn(r.Config.Server, r.Config.Cert())
}

// Returns if reporter is tracking
func (r *Reporter) IsTracking() bool {
	r.stateMutex.RLock()
	defer r.stateMutex.RUnlock()
	return r.isTracking
}

// Login and register as a new device. Certificate is stored in r.Config
func (r *Reporter) Login(server, username, password, deviceName string) bool {
	key, cert, err := auth.Login(server, username, password, deviceName)
	if err != nil {
		return false
	}
	r.Config.Server = server
	r.Config.Key = key
	r.Config.Certificate = cert
	r.Config.ReloadCert()
	r.Config.Save()
	return true
}

// Initiate reporter and go into the main event loop
//
// If it successfully authenticates, a goroutine is created to run event loop,
// and True is returned. Otherwise it returns false.
func (r *Reporter) Run() bool {
	init := make(chan bool)
	go r.run(init)
	return <-init
}

// Clean up and completely exit the reporter (to be differented from StopTracking).
func (r *Reporter) Quit() {
	r.StopTracking()
	cleanup := make(chan bool)
	r.done <- cleanup
	<-cleanup
}
