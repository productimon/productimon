// +build js

package auth

import (
	"fmt"

	"github.com/productimon/wasmws"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func mtlsDial(server string, port int, creds credentials.TransportCredentials, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	grpcServer := fmt.Sprintf("passthrough:///ws://%s/ws", server)
	return grpc.Dial(grpcServer, append([]grpc.DialOption{grpc.WithContextDialer(wasmws.GRPCDialer), grpc.WithTransportCredentials(creds)}, opts...)...)
}
