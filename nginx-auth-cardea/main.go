package main

import "flag"
import "log"
import "net/http"
import "net/url"
import "os"
import "strconv"
import "text/template"

import "github.com/3ofcoins/cardea"

var cfg cardea.Config

var nginxConfigTemplateSource = `
set $cardea_server {{.ServerURL}};
set $cardea_handler {{.HandlerURL}};
set $cardea_location "/.-cardea-.";

location $cardea_location {
  proxy_pass $cardea_handler/;
  proxy_pass_request_body off;
  proxy_set_header Content-Length "";
  proxy_set_header X-Cardea-RequestInfo "$remote_addr $scheme://$host$request_uri";
  internal;
}

auth_request $cardea_location;
error_page 403 =301 $cardea_server?reason=$cardea_nonce&ref=$scheme://$host$request_uri;

auth_request_set $cardea_user $upstream_http_x_cardea_user;
auth_request_set $cardea_roles $upstream_http_x_cardea_roles;
auth_request_set $cardea_nonce $upstream_http_x_cardea_nonce;
`

type nginxConfigParameters struct {
	ServerURL  string
	HandlerURL string
}

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

	log.Println("Starting httpd on", *listen)

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
		} else {
			cfg.ServeHTTP(w, r)
		}
	})

	nginxConfigTemplate := template.New("nginx.conf")
	nginxConfigTemplate.Parse(nginxConfigTemplateSource)

	mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		var params nginxConfigParameters
		qp := r.URL.Query()

		if qp["server"] != nil {
			params.ServerURL = qp["server"][0]
		} else {
			params.ServerURL = "FIXME"
		}

		if qp["handler"] != nil {
			params.HandlerURL = qp["handler"][0]
		} else {
			u := &url.URL{"http", "", nil, r.Host, "", "", ""}
			if r.Header["X-Forwarded-Proto"] != nil {
				u.Scheme = r.Header["X-Forwarded-Proto"][0]
			}
			params.HandlerURL = u.String()
		}

		nginxConfigTemplate.Execute(w, params)
	})

	http.ListenAndServe(*listen, mux)
}
