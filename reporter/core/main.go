package main

import "C"
import (
	"context"
	cpb "git.yiad.am/productimon/proto/common"
	spb "git.yiad.am/productimon/proto/svc"
	"google.golang.org/grpc"
	"log"
	"time"
)

const ChannelBufferSize = 4096

var mq chan string
var eq chan *cpb.Event
var done chan chan bool

func connectToServer(server string, creds *Credentials, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.Dial(server, append([]grpc.DialOption{grpc.WithInsecure(), grpc.WithPerRPCCredentials(creds)}, opts...)...)
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
	defer conn.Close()
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
	event := &cpb.Event{
		Device:    &cpb.Device{Id: "test-dev", DeviceType: 0},
		Id:        0,
		Timestamp: &cpb.Timestamp{Nanos: time.Now().UnixNano()},
		Kind:      &cpb.Event_AppSwitchEvent{&cpb.AppSwitchEvent{AppName: C.GoString(programName)}},
	}
	// TODO do some magic here in case the channel buffer is full
	eq <- event
}

//export SendStartTrackingEvent
func SendStartTrackingEvent() {
	event := &cpb.Event{
		Device:    &cpb.Device{Id: "test-dev", DeviceType: 0},
		Id:        0,
		Timestamp: &cpb.Timestamp{Nanos: time.Now().UnixNano()},
		Kind:      &cpb.Event_StartTrackingEvent{},
	}
	eq <- event
}

//export SendStopTrackingEvent
func SendStopTrackingEvent() {
	event := &cpb.Event{
		Device:    &cpb.Device{Id: "test-dev", DeviceType: 0},
		Id:        0,
		Timestamp: &cpb.Timestamp{Nanos: time.Now().UnixNano()},
		Kind:      &cpb.Event_StopTrackingEvent{},
	}
	eq <- event
}

//export QuitReporter
func QuitReporter() {
	cleanup := make(chan bool)
	done <- cleanup
	<-cleanup
}

func main() {
	panic("You shouldn't run core as a separate binary!")
}
