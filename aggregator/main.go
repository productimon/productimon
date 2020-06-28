package main

import (
	"database/sql"
	"flag"
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
	jsFilename            string
	mapFilename           string
)

func init() {
	flag.StringVar(&flagGRPCListenAddress, "grpc_listen_address", "0.0.0.0:4200", "gRPC listen address")
	flag.StringVar(&flagHTTPListenAddress, "http_listen_address", "0.0.0.0:4201", "HTTP listen address")
	flag.StringVar(&flagPublicKeyPath, "jwt_public_key", "jwtRS256.key.pub", "Path to JWT public key")
	flag.StringVar(&flagPrivateKeyPath, "jwt_private_key", "jwtRS256.key", "Path to JWT private key")
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
	grpcServer := grpc.NewServer()
	reflection.Register(grpcServer)
	s := &service{
		auther: auther,
		db:     db,
	}
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

	httpServer := &http.Server{Addr: flagHTTPListenAddress, Handler: mux}
	err = httpServer.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
