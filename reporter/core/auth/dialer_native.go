// +build !js

package auth

import (
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func mtlsDial(server string, port int, creds credentials.TransportCredentials, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	grpcServer := fmt.Sprintf("%s:%d", strings.Split(server, ":")[0], port)
	return grpc.Dial(grpcServer, append([]grpc.DialOption{grpc.WithTransportCredentials(creds)}, opts...)...)
}
