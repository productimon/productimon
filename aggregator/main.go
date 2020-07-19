package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	spb "git.yiad.am/productimon/proto/svc"
	"git.yiad.am/productimon/viewer/webfe"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	_ "github.com/mattn/go-sqlite3"
	"github.com/productimon/wasmws"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"nhooyr.io/websocket"
)

var (
	flagGRPCListenAddress string
	flagHTTPListenAddress string
	flagPublicKeyPath     string
	flagPrivateKeyPath    string
	flagDBFilePath        string
	flagGRPCPublicPort    int
	flagDebug             bool
	jsFilename            string
	mapFilename           string
	logger                *zap.Logger
)

func init() {
	flag.StringVar(&flagGRPCListenAddress, "grpc_listen_address", "0.0.0.0:4200", "gRPC listen address")
	flag.StringVar(&flagHTTPListenAddress, "http_listen_address", "0.0.0.0:4201", "HTTP listen address (TODO: HTTPS only)")
	flag.IntVar(&flagGRPCPublicPort, "grpc_public_port", 4200, "gRPC public-facing port (this usually needs to be the same port as grpc_listen_address, unless you have some fancy NAT infra)")
	flag.StringVar(&flagPublicKeyPath, "ca_cert", "ca.pem", "Path to CA cert")
	flag.StringVar(&flagPrivateKeyPath, "ca_key", "ca.key", "Path to CA key")
	flag.StringVar(&flagDBFilePath, "db_path", "db.sqlite3", "Path to SQLite3 database file")
	flag.BoolVar(&flagDebug, "debug", false, "enable debug logging")
}

func main() {
	var err error
	flag.Parse()
	if flagDebug {
		logconfig := zap.NewDevelopmentConfig()
		logconfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		logger, err = logconfig.Build()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		panic(err)
	}
	logger.Info("Productimon aggregator starting up...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	auther, err := NewAuthenticator(flagPublicKeyPath, flagPrivateKeyPath)
	if err != nil {
		logger.Fatal("can't create authenticator", zap.Error(err))
	}
	db, err := sql.Open("sqlite3", flagDBFilePath+"?_journal_mode=wal&_txlock=immediate&_busy_timeout=5000")
	if err != nil {
		logger.Fatal("can't open database", zap.Error(err))
	}
	// db.SetMaxOpenConns(1)
	lis, err := net.Listen("tcp", flagGRPCListenAddress)
	if err != nil {
		logger.Fatal("can't listen on grpc address", zap.Error(err), zap.String("grpc_listen_address", flagGRPCListenAddress))
	}
	grpcCreds, err := auther.GrpcCreds()
	if err != nil {
		logger.Fatal("can't create grpc credentials", zap.Error(err))
	}
	grpcServer := grpc.NewServer(grpcCreds)
	reflection.Register(grpcServer)
	s := NewService(auther, db, logger)
	spb.RegisterDataAggregatorServer(grpcServer, s)
	wrappedGrpc := grpcweb.WrapServer(grpcServer)

	mux := http.NewServeMux()
	mux.Handle("/rpc/", http.StripPrefix("/rpc", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		wrappedGrpc.ServeHTTP(w, r)
	})))

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
		w.Header().Set("Access-Control-Allow-Origin", "*")
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
			logger.Error("can't encode /rpc.json", zap.Error(err))
		}
	})
	// InsecureSkipVerify means not to verify origin header
	// because when a Chrome extension visits us, it doesn't attach the Origin header
	// since we're doing mTLS handshake on top of raw websocket, this is safe against CSRF.
	wsl := wasmws.NewWebSocketListener(ctx, &websocket.AcceptOptions{InsecureSkipVerify: true})
	mux.HandleFunc("/ws", wsl.ServeHTTP)

	if flagDebug {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	httpServer := &http.Server{Addr: flagHTTPListenAddress, Handler: mux}

	go func() {
		defer cancel()
		if herr := httpServer.ListenAndServe(); herr != nil {
			logger.Error("can't listen http server", zap.Error(herr), zap.String("address", flagHTTPListenAddress))
		}
	}()
	go func() {
		defer cancel()
		if gerr := grpcServer.Serve(wsl); gerr != nil {
			logger.Error("can't serve grpc server", zap.Error(err))
		}
	}()
	go func() {
		defer cancel()
		if gerr := grpcServer.Serve(lis); gerr != nil {
			logger.Error("can't serve grpc server", zap.Error(err))
		}
	}()
	go func() {
		defer cancel()
		s.runLabelRoutine()
	}()

	// Handle signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		logger.Sugar().Info("Received shutdown signal: ", <-sigs)
		cancel()
	}()

	// Shutdown
	<-ctx.Done()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Second*2)
	defer shutdownCancel()

	grpcShutdown := make(chan struct{}, 1)
	go func() {
		grpcServer.GracefulStop()
		grpcShutdown <- struct{}{}
	}()

	httpServer.Shutdown(shutdownCtx)
	select {
	case <-grpcShutdown:
	case <-shutdownCtx.Done():
		grpcServer.Stop()
	}
}
