package main

import "net/http"
import "net/url"
import "text/template"

var configTemplate = template.New("nginx.conf")

func init() {
	configTemplate.Parse(`
set $cardea_server {{.ServerURL}};
set $cardea_handler {{.HandlerURL}};
set $cardea_location "/.-cardea-.";

location $cardea_location {
  proxy_pass $cardea_handler/;
  proxy_pass_request_body off;
  proxy_set_header Content-Length "";
  proxy_set_header X-Cardea-RequestInfo "$remote_addr $scheme://$host$request_uri";
  proxy_set_header X-Cardea-HMAC-Extra "$http_user_agent";
  # proxy_set_header X-Cardea-HMAC-Extra "$remote_addr";
  internal;
}

auth_request $cardea_location;
error_page 403 =301 $cardea_server?reason=$cardea_nonce&ref=$scheme://$host$request_uri;

auth_request_set $cardea_user $upstream_http_x_cardea_user;
auth_request_set $cardea_roles $upstream_http_x_cardea_roles;
auth_request_set $cardea_nonce $upstream_http_x_cardea_nonce;
`)
}

type ConfigParameters struct {
	ServerURL  string
	HandlerURL string
}

func HandleConfig(w http.ResponseWriter, r *http.Request) {
	params := &ConfigParameters{"FIXME", "FIXME"}
	qry := r.URL.Query()

	if qry["server"] != nil {
		params.ServerURL = qry["server"][0]
	}

	if qry["handler"] != nil {
		params.HandlerURL = qry["handler"][0]
	} else {
		u := &url.URL{Scheme: "http", Host: r.Host}
		if r.Header["X-Forwarded-Proto"] != nil {
			u.Scheme = r.Header["X-Forwarded-Proto"][0]
		}
		params.HandlerURL = u.String()
	}

	configTemplate.Execute(w, params)
}
