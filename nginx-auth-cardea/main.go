package main

import "flag"
import "log"
import "net"
import "net/http"
import "os"
import "strconv"
import "time"

import "github.com/3ofcoins/cardea"

var cfg cardea.Config

func main() {
	default_expiration_sec := cardea.DEFAULT_EXPIRATION_SEC
	default_cookie_name := cardea.DEFAULT_COOKIE_NAME

	if env := os.Getenv("CARDEA_COOKIE"); env != "" {
		default_cookie_name = env
	}

	if env := os.Getenv("CARDEA_EXPIRATION_SEC"); env != "" {
		if n, err := strconv.ParseUint(env, 10, 64); err == nil {
			default_expiration_sec = n
		} else {
			log.Fatal(err)
		}
	}

	secret := flag.String("secret", os.Getenv("CARDEA_SECRET"),
		"(not recommended, use CARDEA_SECRET environment variable instead if possible)")
	flag.StringVar(&cfg.Cookie, "cookie", default_cookie_name,
		"Name of Cardea's cookie")
	flag.Uint64Var(&cfg.ExpirationSec, "expiration-sec", default_expiration_sec,
		"Cookie older than this many seconds will be considered expired")
	listen := flag.String("listen", ":8080", "ip:port to listen on")
	flag.Parse()

	if secret == nil || *secret == "" {
		log.Fatal("Need a secret to start; set CARDEA_SECRET environment variable")
	}

	cfg.Secret = []byte(*secret)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
		} else {
			cfg.ServeHTTP(w, r)
		}
	})

	http.HandleFunc("/config", HandleConfig)

	// We duplicate http.ListenAndServe here to intercept the actual
	// listener and print the actual listen address, so that we can do
	// listen on "127.0.0.1:0" and print the actual port to the log
	// stream.

	ln, err := net.Listen("tcp", *listen)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Listening on", ln.Addr())

	if err := http.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)}, nil); err != nil {
		log.Fatalln(err)
	}
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
