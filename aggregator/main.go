package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"

	spb "git.yiad.am/productimon/proto/svc"
	"git.yiad.am/productimon/viewer/webfe"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	flagGRPCListenAddress string
	flagHTTPListenAddress string
	flagPublicKeyPath     string
	flagPrivateKeyPath    string
	flagDBFilePath        string
	flagGRPCPublicPort    int
	jsFilename            string
	mapFilename           string
)

func init() {
	flag.StringVar(&flagGRPCListenAddress, "grpc_listen_address", "0.0.0.0:4200", "gRPC listen address")
	flag.StringVar(&flagHTTPListenAddress, "http_listen_address", "0.0.0.0:4201", "HTTP listen address (TODO: HTTPS only)")
	flag.IntVar(&flagGRPCPublicPort, "grpc_public_port", 4200, "gRPC public-facing port (this usually needs to be the same port as grpc_listen_address, unless you have some fancy NAT infra)")
	flag.StringVar(&flagPublicKeyPath, "ca_cert", "ca.pem", "Path to CA cert")
	flag.StringVar(&flagPrivateKeyPath, "ca_key", "ca.key", "Path to CA key")
	flag.StringVar(&flagDBFilePath, "db_path", "db.sqlite3", "Path to SQLite3 database file")
}

func main() {
	flag.Parse()
	auther, err := NewAuthenticator(flagPublicKeyPath, flagPrivateKeyPath)
	if err != nil {
		panic(err)
	}
	db, err := sql.Open("sqlite3", flagDBFilePath)
	if err != nil {
		panic(err)
	}
	lis, err := net.Listen("tcp", flagGRPCListenAddress)
	if err != nil {
		panic(err)
	}
	grpcCreds, err := auther.GrpcCreds()
	if err != nil {
		panic(err)
	}
	grpcServer := grpc.NewServer(grpcCreds)
	reflection.Register(grpcServer)
	s := NewService(auther, db)
	spb.RegisterDataAggregatorServer(grpcServer, s)
	go func() {
		gerr := grpcServer.Serve(lis)
		if gerr != nil {
			panic(gerr)
		}
	}()
	wrappedGrpc := grpcweb.WrapServer(grpcServer)

	mux := http.NewServeMux()
	mux.Handle("/rpc/", http.StripPrefix("/rpc", http.HandlerFunc(wrappedGrpc.ServeHTTP)))

	mux.HandleFunc("/app.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/javascript")
		w.Write(webfe.Data[jsFilename])
	})

	mux.HandleFunc("/app.dev.js.map", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(webfe.Data[mapFilename])
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(webfe.Data["index.html"])
	})

	mux.HandleFunc("/rpc.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		j := json.NewEncoder(w)
		err := j.Encode(struct {
			Port       int
			PublicKey  []byte
			ServerName string
		}{
			flagGRPCPublicPort,
			auther.certPEM,
			"api.productimon.com",
		})
		if err != nil {
			log.Println(err)
		}
	})

	httpServer := &http.Server{Addr: flagHTTPListenAddress, Handler: mux}
	err = httpServer.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
