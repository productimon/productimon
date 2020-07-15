package main

import "C"
import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	cpb "git.yiad.am/productimon/proto/common"
	spb "git.yiad.am/productimon/proto/svc"
	"golang.org/x/crypto/ssh/terminal"
	"google.golang.org/grpc"
)

type Config struct {
	workDir     string
	Server      string
	Key         []byte
	Certificate []byte
	cert        tls.Certificate
	LastEid     int64
}

var config Config

// get default work dir
// defaults to $HOME/.productimon
// if unavailable, fall back to cwd
// if still unavailable, return empty string
func defaultWorkDir() string {
	user, err := user.Current()
	if err != nil {
		log.Println(err)
		path, err := os.Getwd()
		if err != nil {
			log.Println(err)
		}
		return path
	}
	if user.HomeDir == "" {
		return ""
	}
	return filepath.Join(user.HomeDir, ".productimon")
}

func init() {
	flag.StringVar(&config.workDir, "work_dir", defaultWorkDir(), "Path to productimon working dir")
	flag.StringVar(&config.Server, "server", "127.0.0.1:4201", "Server Address (this will get overriden by config file, if exists)")
}

//export ReadConfig
func ReadConfig() {
	flag.Parse()
	if len(config.workDir) == 0 {
		panic("Cannot determine default working directory, please specify manually via --work_dir flag")
	}
	config.workDir = filepath.Clean(config.workDir)
	if err := os.MkdirAll(config.workDir, 0700); err != nil {
		panic(err)
	}
	file, err := os.Open(filepath.Join(config.workDir, "config.json"))
	if err == nil {
		defer file.Close()
		decoder := json.NewDecoder(file)
		if err = decoder.Decode(&config); err != nil {
			panic(err)
		}
	} else {
		log.Printf("Can't open config.json: %v", err)
	}
	if len(config.Certificate) > 0 {
		if config.cert, err = tls.X509KeyPair(config.Certificate, config.Key); err != nil {
			panic(err)
		}
	}
}

func interactiveLogin() bool {
	var username, password, deviceName string
	fmt.Printf("username? ")
	fmt.Scanln(&username)
	fmt.Printf("password? ")
	bytePassword, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	if err == nil {
		password = string(bytePassword)
		fmt.Printf("\n")
	} else {
		// see https://github.com/golang/go/issues/11914#issuecomment-613715787
		// Cygwin/mintty/git-bash can't reach down to OS api on Windows
		// and returns "handle is invalid" error
		fmt.Scanln(&password)
	}
	fmt.Printf("deviceName? ")
	fmt.Scanln(&deviceName)
	return login(config.Server, username, password, deviceName)
}

func login(server string, username string, password string, deviceName string) bool {
	var err error
	creds := &Credentials{}

	conn, err := ConnectToServer(server, tls.Certificate{}, grpc.WithPerRPCCredentials(creds))
	if err != nil {
		log.Printf("cannot dial: %v", err)
		return false
	}
	defer conn.Close()

	client := spb.NewDataAggregatorClient(conn)

	// login
	if err = creds.Login(client, username, password); err != nil {
		log.Printf("cannot login: %v", err)
		return false
	}

	rsp, err := client.DeviceSignin(context.Background(), &spb.DataAggregatorDeviceSigninRequest{
		Device: &cpb.Device{
			Name: deviceName,
		},
	})

	if err != nil {
		log.Println(err)
		return false
	}

	config.Server = server
	config.Key = rsp.Key
	config.Certificate = rsp.Cert
	config.cert, err = tls.X509KeyPair(config.Certificate, config.Key)
	config.LastEid = 0
	if err != nil {
		log.Println(err)
		return false
	}

	if err = SaveConfig(); err != nil {
		log.Println(err)
		return false
	}
	return true

}

func SaveConfig() error {
	file, err := os.OpenFile(filepath.Join(config.workDir, "config.json"), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		log.Printf("Failed to save config: %v\n", err)
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err = encoder.Encode(config); err != nil {
		log.Printf("Failed to encode config.json: %v\n", err)
		return err
	}
	log.Println("Config saved")
	return nil
}
