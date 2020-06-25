package main

import (
	"flag"
	"net"

	spb "git.yiad.am/productimon/proto/svc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	flagListenAddress string
	flagDebugAddress  string
)

func init() {
	flag.StringVar(&flagListenAddress, "listen_address", "127.0.0.1:4200", "gRPC listen address")
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", flagListenAddress)
	if err != nil {
		panic(err)
	}
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	s := &service{}
	spb.RegisterDataAggregatorServer(grpcServer, s)
	grpcServer.Serve(lis)

}
