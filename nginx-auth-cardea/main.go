package main

// import "flag"
import "io/ioutil"
import "log"
import "net"
import "net/http"

import "time"
import "github.com/3ofcoins/cardea"
import "github.com/mpasternacki/flag"	// forked from github.com/namsral/flag

func fatalOnError(err error) {
	if err != nil {
		log.Fatalln("FATAL:", err)
	}
}

type handler struct {
	*cardea.Config
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		h.Config.ServeHTTP(w, r)
	case "/config":
		HandleConfig(w, r)
	default:
		http.NotFound(w, r)
	}
}

func main() {
	var secret string
	var literal_secret bool
	var cookie string
	var listen string
	var expires uint64

	flag.StringVar(&secret, "secret", "PATH", "File containing the secret")
	flag.BoolVar(&literal_secret, "literal-secret", false,
		"Use -secret option as a literal secret string, not file name (DANGEROUS)")
	flag.StringVar(&cookie, "cookie", cardea.DEFAULT_COOKIE_NAME, "Name of authentication cookie")
	flag.StringVar(&listen, "listen", ":8080", "ip:port to listen on")
	flag.Uint64Var(&expires, "expires", cardea.DEFAULT_EXPIRATION_SEC,
		"Cookie older than this many seconds will be considered expired")

	flag.String("config", "PATH", "load configuration defaults from file")

	flag.CommandLine.SetEnvPrefix("CARDEA")
	flag.Parse()

	if secret == "PATH" {
		log.Fatalln("FATAL: Secret not supplied")
	}

	if literal_secret {
		log.Println("WARNING: using secret value literally. Hope you know what you are doing.")
	} else {
		secretBytes, err := ioutil.ReadFile(secret)
		fatalOnError(err)
		secret = string(secretBytes)
	}

	cfg := &handler{cardea.NewConfig(secret)}
	cfg.Cookie = cookie
	cfg.ExpirationSec = expires

	// We duplicate http.ListenAndServe here to intercept the actual
	// listener and print the actual listen address, so that we can do
	// listen on "127.0.0.1:0" and print the actual port to the log
	// stream.

	ln, err := net.Listen("tcp", listen)
	fatalOnError(err)

	log.Println("Listening on", ln.Addr())

	fatalOnError(http.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)}, cfg))
}

// Remaining part of copy-paste from net/http's server.go
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}
