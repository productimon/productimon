package main

import "C"
import (
	"context"
	"log"
	"sync"
	"time"

	cpb "git.yiad.am/productimon/proto/common"
	spb "git.yiad.am/productimon/proto/svc"
	"google.golang.org/grpc"
)

const ChannelBufferSize = 4096

var (
	build string

	MaxInputReportingInterval time.Duration

	mq   chan string
	eq   chan *cpb.Event
	done chan chan bool

	nClicks     int64
	nKeystrokes int64
	statsMutex  sync.Mutex

	eid      int64
	eidMutex sync.Mutex

	inputTrackingDone chan bool
	reportInputStats  chan chan bool
)

func connectToServer(server string, creds *Credentials, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.Dial(server, append([]grpc.DialOption{grpc.WithInsecure(), grpc.WithPerRPCCredentials(creds)}, opts...)...)
}

func nowInterval() *cpb.Interval {
	ts := time.Now().UnixNano()
	return &cpb.Interval{Start: &cpb.Timestamp{Nanos: ts}, End: &cpb.Timestamp{Nanos: ts}}
}

func getEid() (id int64) {
	eidMutex.Lock()
	eid++
	id = eid
	eidMutex.Unlock()
	return
}

// TODO didn't thoroughly test all scenarios
// would be worth to test all network conditions
func retrySendEvent(event *cpb.Event, server string, creds *Credentials) (*grpc.ClientConn, spb.DataAggregatorClient, spb.DataAggregator_PushEventClient, error) {
	var conn *grpc.ClientConn
	var err error
	// NOTE this blocks this goroutine until connection to the server
	// is established again
	for {
		conn, err = connectToServer(server, creds, grpc.WithBlock())
		if err == nil {
			break
		}
	}
	client := spb.NewDataAggregatorClient(conn)
	eventStream, err := client.PushEvent(context.Background())
	if err != nil {
		log.Printf("Failed to get stream to push events: %v", err)
		return nil, nil, nil, err
	}
	log.Println("Reconnected to server\n")
	eq <- event // event will be sent out of order, but they're all timestamped
	return conn, client, eventStream, nil
}

func runReporter(init chan bool, server string, username string, password string) {
	log.Printf("productimon core module initiating")
	log.Printf("using server: %s", server)
	creds := &Credentials{}

	// establish grpc connection
	// TODO think about what happens if the user started with offline env
	conn, err := connectToServer(server, creds)
	if err != nil {
		log.Printf("cannot dial: %v", err)
		init <- false
		return
	}
	defer func() {
		conn.Close()
	}()
	client := spb.NewDataAggregatorClient(conn)

	// login
	if err = creds.Login(client, username, password); err != nil {
		init <- false
		log.Printf("cannot login: %v", err)
		return
	}

	eventStream, err := client.PushEvent(context.Background())
	if err != nil {
		init <- false
		log.Printf("Cannot get event stream: %v", err)
		return
	}

	init <- true

	// event loop
	mq = make(chan string)
	eq = make(chan *cpb.Event, ChannelBufferSize)
	done = make(chan chan bool)
	for {
		select {
		case s := <-mq:
			log.Println(s)
		case e := <-eq:
			log.Println("Sending event", e)
			err := eventStream.Send(e)
			if err != nil {
				log.Printf("Got err %v, reconnecting to the server", err)
				conn.Close()
				for err != nil {
					conn, client, eventStream, err = retrySendEvent(e, server, creds)
				}
			}
		case c := <-done:
			log.Println("Shutting down...")
			eventStream.CloseAndRecv()
			c <- true
			return
		}
	}
}

//export InitReporter
func InitReporter(server *C.char, username *C.char, password *C.char) bool {
	if build == "DEBUG" {
		log.Println("Running DEBUG build!")
		MaxInputReportingInterval = 5 * time.Second
	} else {
		MaxInputReportingInterval = 60 * time.Second
	}
	log.Printf("MaxInputReportingInterval: %v", MaxInputReportingInterval)
	init := make(chan bool)
	go runReporter(init, C.GoString(server), C.GoString(username), C.GoString(password))
	return <-init
}

//export SendMessage
func SendMessage(msg *C.char) {
	mq <- C.GoString(msg)
}

//export SendWindowSwitchEvent
func SendWindowSwitchEvent(programName *C.char) {
	done := make(chan bool)
	reportInputStats <- done
	<-done // wait until input event has been sent
	event := &cpb.Event{
		Id:           getEid(),
		Timeinterval: nowInterval(),
		Kind:         &cpb.Event_AppSwitchEvent{&cpb.AppSwitchEvent{AppName: C.GoString(programName)}},
	}
	// TODO do some magic here in case the channel buffer is full
	eq <- event
}

//export SendStartTrackingEvent
func SendStartTrackingEvent() {
	inputTrackingDone = make(chan bool)
	reportInputStats = make(chan chan bool)
	go runInputTracking(reportInputStats, inputTrackingDone)
	event := &cpb.Event{
		Id:           getEid(),
		Timeinterval: nowInterval(),
		Kind:         &cpb.Event_StartTrackingEvent{},
	}
	eq <- event
}

//export SendStopTrackingEvent
func SendStopTrackingEvent() {
	done := make(chan bool)
	reportInputStats <- done
	<-done // wait until input event has been sent
	event := &cpb.Event{
		Id:           getEid(),
		Timeinterval: nowInterval(),
		Kind:         &cpb.Event_StopTrackingEvent{},
	}
	eq <- event
	inputTrackingDone <- true
}

//export HandleMouseClick
func HandleMouseClick() {
	statsMutex.Lock()
	nClicks++
	statsMutex.Unlock()
}

//export HandleKeystroke
func HandleKeystroke() {
	statsMutex.Lock()
	nKeystrokes++
	statsMutex.Unlock()
}

func sendInputStats(start, end int64) {
	statsMutex.Lock()
	defer statsMutex.Unlock()
	if nKeystrokes > 0 {
		event := &cpb.Event{
			Id:           getEid(),
			Timeinterval: &cpb.Interval{Start: &cpb.Timestamp{Nanos: start}, End: &cpb.Timestamp{Nanos: end}},
			Kind:         &cpb.Event_KeyStrokeEvent{&cpb.KeyStrokeEvent{Keystrokes: nKeystrokes}},
		}
		eq <- event
		nKeystrokes = 0
	}
	if nClicks > 0 {
		event := &cpb.Event{
			Id:           getEid(),
			Timeinterval: &cpb.Interval{Start: &cpb.Timestamp{Nanos: start}, End: &cpb.Timestamp{Nanos: end}},
			Kind:         &cpb.Event_MouseClickEvent{&cpb.MouseClickEvent{Mouseclicks: nClicks}},
		}
		eq <- event
		nClicks = 0
	}
}

func runInputTracking(reportInputStats chan chan bool, finish chan bool) {
	timer := time.NewTicker(MaxInputReportingInterval)
	for {
		start := time.Now().UnixNano()
		select {
		case <-finish: // quit InputTracking goroutine
			timer.Stop()
			return
		case done := <-reportInputStats: // report stats before switching
			end := time.Now().UnixNano()
			sendInputStats(start, end)
			done <- true
			timer.Stop()
			timer = time.NewTicker(MaxInputReportingInterval) // restart the timer for new program
		case <-timer.C: // tick for max interval
			end := time.Now().UnixNano()
			sendInputStats(start, end)
		}
	}
}

//export QuitReporter
func QuitReporter(isTracking C.char) {
	if isTracking != 0 {
		SendStopTrackingEvent()
	}
	cleanup := make(chan bool)
	done <- cleanup
	<-cleanup
}

func main() {
	panic("You shouldn't run core as a separate binary!")
}
