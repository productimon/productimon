package main

import "C"
import (
	"log"

	spb "git.yiad.am/productimon/proto/svc"
	"google.golang.org/grpc"
)

var mq chan string
var done chan chan bool

func runReporter(init chan bool, server string, username string, password string) {
	log.Printf("productimon core module initiating")
	log.Printf("using server: %s", server)
	creds := &Credentials{}

	// establish grpc connection
	conn, err := grpc.Dial(server, grpc.WithInsecure(), grpc.WithPerRPCCredentials(creds))
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
	init <- true

	// event loop
	mq = make(chan string)
	done = make(chan chan bool)
	for {
		select {
		case s := <-mq:
			log.Println(s)
		case c := <-done:
			log.Println("Shutting down...")
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

//export QuitReporter
func QuitReporter() {
	cleanup := make(chan bool)
	done <- cleanup
	<-cleanup
}

func main() {
	panic("You shouldn't run core as a separate binary!")
}
