package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"strings"

	"git.yiad.am/productimon/aggregator/authenticator"
	"git.yiad.am/productimon/aggregator/service"
	"git.yiad.am/productimon/internal"
	"git.yiad.am/productimon/viewer/webfe"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/productimon/wasmws"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
	"nhooyr.io/websocket"
)

var (
	flagHTTPListenAddress  string
	flagHTTPSListenAddress string
	flagCertDir            string
	flagTOSAccepted        bool
)

var (
	jsFilename  string
	mapFilename string
)

func init() {
	flag.StringVar(&flagHTTPListenAddress, "http_listen_address", "0.0.0.0:80", "HTTP listen address")
	flag.StringVar(&flagHTTPSListenAddress, "https_listen_address", "0.0.0.0:443", "HTTPS listen address")
	flag.StringVar(&flagCertDir, "cert_cache_dir", ".certs", "Path to directory to store HTTPS certificates. Concatenate your key and cert to a single file and put it in this directory with your domain name as filename without any extra file extension (e.g., .certs/my.productimon.com). If you don't provide a certificate, one will be provisioned automatically for you via Let's Encrypt")
	flag.BoolVar(&flagTOSAccepted, "accept_acme_tos", false, "Accept Let's Encrypt Terms of Service (you don't have to pass this if you provided your own certificate)")
}

func webfeServeStaticFile(contentType, filename string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.Write(webfe.Data[filename])
	}
}

func redirectSSL(w http.ResponseWriter, r *http.Request) {
	target := "https://" + r.Host + r.URL.Path
	if len(r.URL.RawQuery) > 0 {
		target += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, target,
		http.StatusTemporaryRedirect)
}

func buildHandlers(ctx context.Context, s *service.Service, auther *authenticator.Authenticator, wrappedGrpc *grpcweb.WrappedGrpcServer) (http.Handler, net.Listener) {
	mux := http.NewServeMux()

	mux.Handle("/rpc/", http.StripPrefix("/rpc", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		wrappedGrpc.ServeHTTP(w, r)
	})))

	mux.HandleFunc("/app.js", webfeServeStaticFile("text/javascript", jsFilename))
	mux.HandleFunc("/app.dev.js.map", webfeServeStaticFile("application/json", mapFilename))
	mux.HandleFunc("/", webfeServeStaticFile("text/html", "index.html"))
	mux.HandleFunc("/favicon.ico", webfeServeStaticFile("image/x-icon", "favicon.ico"))
	mux.HandleFunc("/logo.svg", webfeServeStaticFile("image/svg+xml", "productimon.svg"))
	mux.HandleFunc("/logo-white.svg", webfeServeStaticFile("image/svg+xml", "productimon_white.svg"))

	mux.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		tokens := r.Form["token"]
		if len(tokens) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("token missing"))
			return
		}

		if err := s.VerifyAccount(tokens[0]); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write([]byte("Account verified! You may login now"))
	})

	mux.HandleFunc("/rpc.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		j := json.NewEncoder(w)
		err := j.Encode(struct {
			Port          int
			PublicKey     []byte
			ServerName    string
			ServerVersion string
		}{
			flagGRPCPublicPort,
			auther.CertPEM(),
			strings.Split(flagDomain, ":")[0],
			internal.GitVersion,
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

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "productimon")
		mux.ServeHTTP(w, r)
	})

	return handler, wsl
}

func acceptTOS(tosURL string) bool {
	if flagTOSAccepted {
		return true
	}
	fmt.Println("You didn't provide your own HTTPS certificate. Don't worry we can automatically provision one for you for free from Let's Encrypt")
	fmt.Println("You need to accept Let's Encrypt's terms of service before we can contonue")
	fmt.Printf("You can find the terms of service here: %s\n", tosURL)
	fmt.Println("If you agree to the terms of service, please restart the server with `-accept_acme_tos` flag")
	return false
}

func NewHTTPServer(ctx context.Context, s *service.Service, auther *authenticator.Authenticator, wrappedGrpc *grpcweb.WrappedGrpcServer) (httpServer *http.Server, httpsServer *http.Server, grpcListener net.Listener) {
	handler, grpcListener := buildHandlers(ctx, s, auther, wrappedGrpc)

	// in case public-facing address is not on port 443.
	domain := strings.Split(flagDomain, ":")[0]

	certManager := &autocert.Manager{
		Cache:      autocert.DirCache(flagCertDir),
		Prompt:     acceptTOS,
		HostPolicy: autocert.HostWhitelist(domain),
	}

	httpServer = &http.Server{
		Addr:    flagHTTPListenAddress,
		Handler: certManager.HTTPHandler(http.HandlerFunc(redirectSSL)),
	}

	httpsServer = &http.Server{
		Addr:      flagHTTPSListenAddress,
		Handler:   handler,
		TLSConfig: certManager.TLSConfig(),
	}

	return
}
