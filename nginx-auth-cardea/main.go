package main

import "flag"
import "log"
import "net/http"
import "os"
import "strconv"

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

	flag.StringVar(&cfg.Secret, "secret", os.Getenv("CARDEA_SECRET"),
		"(not recommended, use CARDEA_SECRET environment variable instead if possible)")
	flag.StringVar(&cfg.Cookie, "cookie", default_cookie_name,
		"Name of Cardea's cookie")
	flag.Uint64Var(&cfg.ExpirationSec, "expiration-sec", default_expiration_sec,
		"Cookie older than this many seconds will be considered expired")
	listen := flag.String("listen", ":8080", "ip:port to listen on")
	flag.Parse()

	if cfg.Secret == "" {
		log.Fatal("Need a secret to start; set CARDEA_SECRET environment variable")
	}

	log.Println("Starting httpd on", *listen)
	http.ListenAndServe(*listen, &cfg)
}
