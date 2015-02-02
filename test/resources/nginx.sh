#!/bin/sh
set -e

curl -S -o test/resources/nginx.authreq.conf --retry 5 http://127.0.0.1:$(($PORT-100))/config

cat <<EOF >test/resources/nginx.conf
worker_processes 1;
error_log stderr info;
daemon off;

events {
  worker_connections 32;
  accept_mutex off;
  use kqueue;
}

http {
  default_type application/octet-stream;
  access_log off;
  sendfile on;
  index index.html README.md;
  server {
    listen $PORT default;
    server_name _;
    root "$(pwd)";

    include nginx.authreq.conf;

    add_header X-Cardea-User \$cardea_user;
    location /config {
      auth_request off;
    }
  }
}
EOF

exec nginx -c $(pwd)/test/resources/nginx.conf "${@}"
