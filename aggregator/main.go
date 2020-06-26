package main

import (
	"flag"
	"net"

	spb "git.yiad.am/productimon/proto/svc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	flagListenAddress  string
	flagPublicKeyPath  string
	flagPrivateKeyPath string
)

func init() {
	flag.StringVar(&flagListenAddress, "listen_address", "127.0.0.1:4200", "gRPC listen address")
	flag.StringVar(&flagPublicKeyPath, "jwt_public_key", "jwtRS256.key.pub", "Path to JWT public key")
	flag.StringVar(&flagPrivateKeyPath, "jwt_private_key", "jwtRS256.key", "Path to JWT private key")
}

func main() {
	flag.Parse()
	auther, err := NewAuthenticator(flagPublicKeyPath, flagPrivateKeyPath)
	if err != nil {
		panic(err)
	}
	lis, err := net.Listen("tcp", flagListenAddress)
	if err != nil {
		panic(err)
	}
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	s := &service{
		auther: auther,
	}
	spb.RegisterDataAggregatorServer(grpcServer, s)
	grpcServer.Serve(lis)
}
