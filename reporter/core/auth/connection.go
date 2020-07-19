package auth

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Establish a connection to the server. The server string here should be an HTTP address, not gRPC.
func ConnectToServer(server string, cert tls.Certificate, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	rsp, err := http.Get(fmt.Sprintf("http://%s/rpc.json", server))
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()
	var rpcConfig struct {
		Port       int
		PublicKey  []byte
		ServerName string
	}
	j := json.NewDecoder(rsp.Body)
	if err := j.Decode(&rpcConfig); err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(rpcConfig.PublicKey)
	creds := credentials.NewTLS(&tls.Config{
		ServerName:   rpcConfig.ServerName,
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
	})
	return mtlsDial(server, rpcConfig.Port, creds, opts...)
}
