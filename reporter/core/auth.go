package main

import (
	"context"
	"log"

	cpb "git.yiad.am/productimon/proto/common"
	spb "git.yiad.am/productimon/proto/svc"
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
	return false
}
