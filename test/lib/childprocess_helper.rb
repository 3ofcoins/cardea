require 'socket'
require 'tempfile'
require 'timeout'

require 'childprocess'

# rubocop:disable Style/GlobalVars

ChildProcess.posix_spawn = true

module Cardea
  module Spec
    class Endpoint
      class << self
        def define(name, *args, &block)
          (@defs ||= {})[name] = [ args, block ]
          nil
        end

        def [](name)
          (@procs ||= {})[name] ||=
            begin
              args, block = @defs[name]
              new(name, *args, &block)
            end
        end

        def each(&block)
          @procs.values.each(&block) if @procs
        end
      end

      attr_reader :name, :process, :opts, :port, :log

      def initialize(name, *args, &block)
        @name = name
        @args = args
        instance_eval(&block)

        @process = ChildProcess.build(*@command)
        @process.io.stdout = Tempfile.new("childprocess-#{name}-stdout")
        @process.io.stderr = Tempfile.new("childprocess-#{name}-stderr")
        @log = File.open(@process.io.stderr.path)

        # assert tcp is not listening yet (no leftover processes)
        begin
          sock = TCPSocket.new('localhost', port)
        rescue Errno::ECONNREFUSED
          # this is good
        else
          sock.close
          raise "#{name}: port #{port} already listening"
        end

        process.start

        Timeout.timeout 10 do
          loop do
            fail 'process died' unless process.alive?
            begin
              sock = TCPSocket.new('localhost', port)
            rescue Errno::ECONNREFUSED
              sleep 0.1
            else
              sock.close
              break
            end
          end
        end

        log_ff

        Minitest.after_run { self.cleanup! }
      end

      def cleanup!
        process.stop if process.alive?
        @log.close
        [
          process.io.stdout.path,
          process.io.stderr.path,
          *Array(@remove_files),
        ].each do |f|
          File.unlink(f) rescue nil
        end
      end

      def log_ff
        log.seek(0, :END)
      end

      def http
        @http ||= Net::HTTP.start('127.0.0.1', port)
      end

      def get(path, headers={})
        req = Net::HTTP::Get.new(path)
        headers.each do |hdr, val|
          req[hdr] = val
        end
        http.request(req)
      end
    end

    def setup
      super
      Endpoint.each { |p| p.log_ff }
    end
    CARDEA_SECRET = 'SWORDFISH'

    Endpoint.define :authreq do
      @port = ENV['AUTHREQ_PORT'] ? ENV['AUTHREQ_PORT'].to_i : 4000
      @command = [
        './target/nginx-auth-cardea',
        '-listen', "127.0.0.1:#{port}",
        '-secret', CARDEA_SECRET,
        '-literal-secret'
      ]
    end

    Endpoint.define :nginx do
      @port = ENV['NGINX_PORT'] ? ENV['NGINX_PORT'].to_i : 4001

      # get config snippet from authreq
      res = Endpoint[:authreq].http.get('/config?server=http://example.com/')
      fail unless res.code == '200'
      cardea_snippet = res.body

      nginx_conf = Tempfile.new('nginx.conf.')
      nginx_conf.write <<EOF
worker_processes 1;
error_log stderr;
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
    listen #{@port} default;
    server_name _;
    root #{File.expand_path(File.join(File.dirname(__FILE__), '../..'))};
    #{cardea_snippet}
    add_header X-Cardea-User $cardea_user;
    location /config {
      auth_request off;
    }
  }
}
EOF
      nginx_conf.close
      @remove_files = [nginx_conf.path]
      @command = ['nginx', '-c', nginx_conf.path]
    end
  end
end
