package reporter

import (
	"context"
	"log"
	"time"

	cpb "git.yiad.am/productimon/proto/common"
	spb "git.yiad.am/productimon/proto/svc"
	"git.yiad.am/productimon/reporter/core/auth"
	"google.golang.org/grpc"
)

// Block current goroutine until it can connect to the server
// and send event
func (r *Reporter) retrySendEvent(event *cpb.Event) (*grpc.ClientConn, spb.DataAggregatorClient, spb.DataAggregator_PushEventClient, error) {
	// TODO didn't thoroughly test all scenarios
	// would be worth to test all network conditions
	var conn *grpc.ClientConn
	var err error
	// NOTE this blocks this goroutine until connection to the server
	// is established again
	for {
		if conn, err = auth.ConnectToServer(r.Config.Server, r.Config.Cert(), grpc.WithBlock()); err == nil {
			break
		}
	}
	client := spb.NewDataAggregatorClient(conn)
	eventStream, err := client.PushEvent(context.Background())
	if err != nil {
		log.Printf("Failed to get stream to push events: %v", err)
		return nil, nil, nil, err
	}
	log.Println("Reconnected to server")
	r.eq <- event // event will be sent out of order, but they're all timestamped
	return conn, client, eventStream, nil
}

// main event loop, blocking call
func (r *Reporter) eventLoop(conn *grpc.ClientConn, client spb.DataAggregatorClient, eventStream spb.DataAggregator_PushEventClient) {
	defer func() {
		conn.Close()
	}()
	for {
		select {
		case e := <-r.eq:
			log.Println("Sending event", e)
			err := eventStream.Send(e)
			if err != nil {
				log.Printf("Got err %v, reconnecting to the server", err)
				conn.Close()
				for err != nil {
					conn, client, eventStream, err = r.retrySendEvent(e)
				}
			}
		case c := <-r.done:
			log.Println("Shutting down...")
			eventStream.CloseAndRecv()
			c <- true
			return
		}
	}
}

// Call this in platform specific code when you detect that the user
// switches window.
func (r *Reporter) SwitchWindow(programName string) {
	r.stateMutex.RLock()
	defer r.stateMutex.RUnlock()
	if !r.isTracking {
		return
	}
	done := make(chan bool)
	r.reportInputStats <- done
	<-done // wait until input event has been sent because we need entire stats interval to be for the same app
	event := &cpb.Event{
		Id:           r.getEid(),
		Timeinterval: nowInterval(),
		Kind:         &cpb.Event_AppSwitchEvent{&cpb.AppSwitchEvent{AppName: programName}},
	}
	// TODO do some magic here in case the channel buffer is full
	r.eq <- event
}

// Call this to start tracking.
func (r *Reporter) StartTracking() {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()
	if r.isTracking {
		return
	}
	r.isTracking = true
	go r.runInputTracking()
	event := &cpb.Event{
		Id:           r.getEid(),
		Timeinterval: nowInterval(),
		Kind:         &cpb.Event_StartTrackingEvent{},
	}
	r.eq <- event
}

// Call this to stop tracking.
func (r *Reporter) StopTracking() {
	r.stateMutex.Lock()
	defer r.stateMutex.Unlock()
	if !r.isTracking {
		return
	}
	r.isTracking = false
	done := make(chan bool)
	r.reportInputStats <- done
	<-done // wait until input event has been sent
	event := &cpb.Event{
		Id:           r.getEid(),
		Timeinterval: nowInterval(),
		Kind:         &cpb.Event_StopTrackingEvent{},
	}
	r.eq <- event
	r.inputTrackingDone <- true
	r.Config.LastEid = r.eid
	r.Config.Save()
}

// Call this in platform specific code to register a mouse click.
func (r *Reporter) HandleMouseClick() {
	r.inputStatsMutex.Lock()
	r.nClicks++
	r.inputStatsMutex.Unlock()
}

// Call this in platform specific code to register a keystroke.
func (r *Reporter) HandleKeystroke() {
	r.inputStatsMutex.Lock()
	r.nKeystrokes++
	r.inputStatsMutex.Unlock()
}

// Check current click and keystroke counters and generate an ActivityEvent
// if needed.
// It is caller's responsibility to make sure r.isTracking is true
func (r *Reporter) sendInputStats(start, end int64) {
	r.inputStatsMutex.Lock()
	if r.nKeystrokes > 0 || r.nClicks > 0 {
		event := &cpb.Event{
			Id:           r.getEid(),
			Timeinterval: &cpb.Interval{Start: &cpb.Timestamp{Nanos: start}, End: &cpb.Timestamp{Nanos: end}},
			Kind:         &cpb.Event_ActivityEvent{&cpb.ActivityEvent{Keystrokes: r.nKeystrokes, Mouseclicks: r.nClicks}},
		}
		r.eq <- event
		r.nKeystrokes = 0
		r.nClicks = 0
	}
	r.inputStatsMutex.Unlock()
}

// blocking goroutine for generating ActivityEvents with Config.MaxInputReportingInterval
// or before every window switch.
func (r *Reporter) runInputTracking() {
	timer := time.NewTicker(r.Config.MaxInputReportingInterval)
	for {
		start := time.Now().UnixNano()
		select {
		case <-r.inputTrackingDone: // quit InputTracking goroutine
			timer.Stop()
			return
		case done := <-r.reportInputStats: // report stats before switching
			// we don't lock stateMutex here because this is called upon WindowSwitch or StopTracking
			// both of them would acquire stateMutex
			end := time.Now().UnixNano()
			r.sendInputStats(start, end)
			done <- true
			timer.Stop()
			timer = time.NewTicker(r.Config.MaxInputReportingInterval) // restart the timer for new program
		case <-timer.C: // tick for max interval
			r.stateMutex.RLock()
			if r.isTracking {
				end := time.Now().UnixNano()
				r.sendInputStats(start, end)
			}
			r.stateMutex.RUnlock()
		}
	}
}
