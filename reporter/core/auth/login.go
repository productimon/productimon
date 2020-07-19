package auth

import (
	"context"
	"crypto/tls"
	"log"

	cpb "git.yiad.am/productimon/proto/common"
	spb "git.yiad.am/productimon/proto/svc"
	"google.golang.org/grpc"
)

// Login and register a new device, returning the signed certificate for mTLS
func Login(server string, username string, password string, deviceName string) (key, cert []byte, err error) {
	creds := &Credentials{}

	conn, err := ConnectToServer(server, tls.Certificate{}, grpc.WithPerRPCCredentials(creds))
	if err != nil {
		log.Printf("cannot dial: %v", err)
		return nil, nil, err
	}
	defer conn.Close()

	client := spb.NewDataAggregatorClient(conn)

	if err = creds.Login(client, username, password); err != nil {
		log.Printf("cannot login: %v", err)
		return nil, nil, err
	}

	rsp, err := client.DeviceSignin(context.Background(), &spb.DataAggregatorDeviceSigninRequest{
		Device: &cpb.Device{
			Name: deviceName,
		},
	})

	if err != nil {
		log.Println(err)
		return nil, nil, err
	}

	return rsp.Key, rsp.Cert, nil

}

// check if cert is valid against server
func IsLoggedIn(server string, cert tls.Certificate) bool {
	if len(cert.Certificate) == 0 {
		return false
	}
	conn, err := ConnectToServer(server, cert)
	if err != nil {
		log.Printf("cannot dial: %v", err)
		return false
	}
	defer conn.Close()
	client := spb.NewDataAggregatorClient(conn)

	if _, err = client.UserDetails(context.Background(), &cpb.Empty{}); err != nil {
		log.Printf("Failed to get user details %v", err)
		return false
	}
	return true
}
