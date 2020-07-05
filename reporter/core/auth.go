package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	cpb "git.yiad.am/productimon/proto/common"
	spb "git.yiad.am/productimon/proto/svc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Credentials struct {
	token string
}

func (c *Credentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	ret := make(map[string]string)
	ret["Authorization"] = c.token
	return ret, nil
}

func (c *Credentials) Login(client spb.DataAggregatorClient, username, password string) error {
	token, err := client.Login(context.Background(), &spb.DataAggregatorLoginRequest{Email: username, Password: password})
	if err != nil {
		return err
	}
	c.token = token.Token
	log.Printf("auth token: %s", token.Token)
	user, err := client.UserDetails(context.Background(), &cpb.Empty{})
	if err != nil {
		c.token = ""
		return err
	}
	log.Printf("User details: %v", user)
	return nil
}

func (c *Credentials) RequireTransportSecurity() bool {
	return true
}

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
	grpcServer := fmt.Sprintf("%s:%d", strings.Split(server, ":")[0], rpcConfig.Port)
	return grpc.Dial(grpcServer, append([]grpc.DialOption{grpc.WithTransportCredentials(creds)}, opts...)...)
}
