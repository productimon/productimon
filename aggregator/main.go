package main

import (
	"context"
	"database/sql"
	"flag"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"git.yiad.am/productimon/aggregator/authenticator"
	"git.yiad.am/productimon/aggregator/notifications"
	"git.yiad.am/productimon/aggregator/service"
	"git.yiad.am/productimon/internal"
	spb "git.yiad.am/productimon/proto/svc"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	flagGRPCListenAddress string
	flagPublicKeyPath     string
	flagPrivateKeyPath    string
	flagDBFilePath        string
	flagDomain            string
	flagGRPCPublicPort    int
	flagDebug             bool
	flagSMTPServer        string
	flagSMTPUsername      string
	flagSMTPPasswordFile  string
	flagSMTPSender        string
)

var logger *zap.Logger

func init() {
	flag.StringVar(&flagGRPCListenAddress, "grpc_listen_address", "0.0.0.0:4200", "gRPC listen address")
	flag.StringVar(&flagDomain, "domain", "my.productimon.com", "public-facing server domain")
	flag.IntVar(&flagGRPCPublicPort, "grpc_public_port", 4200, "gRPC public-facing port (this usually needs to be the same port as grpc_listen_address, unless you have some fancy NAT infra)")
	flag.StringVar(&flagPublicKeyPath, "ca_cert", "ca.pem", "Path to auth token CA cert (auto-generated if not present)")
	flag.StringVar(&flagPrivateKeyPath, "ca_key", "ca.key", "Path to auth token CA key (auto-generated if not present)")
	flag.StringVar(&flagDBFilePath, "db_path", "db.sqlite3", "Path to SQLite3 database file (will be automatically created for first time)")
	flag.StringVar(&flagSMTPServer, "smtp_server", "", "SMTP server for sending email (leave empty to disable SMTP")
	flag.StringVar(&flagSMTPUsername, "smtp_username", "", "SMTP username for authentication (this is usually the same as sender address, leave empty to disable authentication)")
	flag.StringVar(&flagSMTPPasswordFile, "smtp_password_file", "", "Path to SMTP password file")
	flag.StringVar(&flagSMTPSender, "smtp_sender", "", "SMTP sender address for sending emails")
	flag.BoolVar(&flagDebug, "debug", false, "enable debug logging")
}

func main() {
	var err error
	internal.ParseFlags()
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

	auther, err := authenticator.NewAuthenticator(flagPublicKeyPath, flagPrivateKeyPath, strings.Split(flagDomain, ":")[0])
	if err != nil {
		logger.Fatal("can't create authenticator", zap.Error(err))
	}

	db, err := sql.Open("sqlite3", flagDBFilePath+"?_journal_mode=wal&_txlock=immediate&_busy_timeout=5000&_foreign_keys=1")
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

	s, err := service.NewService(flagDomain, auther, db, logger)
	if err != nil {
		logger.Fatal("can't create service", zap.Error(err))
	}

	spb.RegisterDataAggregatorServer(grpcServer, s)
	wrappedGrpc := grpcweb.WrapServer(grpcServer)
	httpServer, httpsServer, grpcListener := NewHTTPServer(ctx, s, auther, wrappedGrpc)

	if len(flagSMTPServer) > 0 {
		if len(flagSMTPUsername) > 0 {
			smtpPwdBytes, err := ioutil.ReadFile(flagSMTPPasswordFile)
			if err != nil {
				logger.Fatal("failed to read SMTP password", zap.Error(err))
			}
			smtpPwd := strings.TrimSpace(string(smtpPwdBytes))
			s.RegisterNotifier(notifications.NewEmailNotifier(flagSMTPServer, flagSMTPUsername, smtpPwd, flagSMTPSender))
		} else {
			s.RegisterNotifier(notifications.NewEmailNoAuthNotifier(flagSMTPServer, flagSMTPSender))
		}
	}

	go func() {
		defer cancel()
		if herr := httpServer.ListenAndServe(); herr != nil {
			logger.Error("can't listen http server", zap.Error(herr), zap.String("address", flagHTTPListenAddress))
		}
	}()
	go func() {
		defer cancel()
		if herr := httpsServer.ListenAndServeTLS("", ""); herr != nil {
			logger.Error("can't listen https server", zap.Error(herr), zap.String("address", flagHTTPSListenAddress))
		}
	}()
	go func() {
		defer cancel()
		if gerr := grpcServer.Serve(grpcListener); gerr != nil {
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
		s.RunLabelRoutine()
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
