require 'socket'
require 'tempfile'
require 'timeout'

require 'childprocess'

# rubocop:disable Style/GlobalVars

ChildProcess.posix_spawn = true

module Cardea
  module Spec
    CARDEA_AUTHREQ_PORT =
      ENV['CARDEA_AUTHREQ_PORT'] ? ENV['CARDEA_AUTHREQ_PORT'].to_i : 4000
    CARDEA_NGINX_PORT =
      ENV['CARDEA_NGINX_PORT'] ? ENV['CARDEA_NGINX_PORT'].to_i : 4001

    def childprocess(*commandline) # rubocop:disable Metrics/AbcSize
      if commandline.last.is_a? Hash
        opts = commandline.pop
      else
        opts = {}
      end

      process = ChildProcess.build(*commandline)
      process.io.stdout = Tempfile.new('childprocess-stdout')
      process.io.stderr = Tempfile.new('childprocess-stdout')
      process.environment.merge! opts[:env] if opts[:env]
      process.start unless opts[:dont_start]

      Minitest.after_run do
        process.stop if process.alive?
        File.unlink process.io.stdout.path,
                    process.io.stderr.path
      end

      process
    end

    def authreq_handler
      return $authreq_handler if $authreq_handler

      process = childprocess('./target/nginx-auth-cardea',
                             '-listen', "127.0.0.1:#{CARDEA_AUTHREQ_PORT}",
                             '-secret', secret,
                             '-literal-secret')
      Timeout.timeout 10 do
        loop do
          fail 'nginx-auth-cardea died' unless process.alive?
          begin
            sock = TCPSocket.new('localhost', CARDEA_AUTHREQ_PORT)
          rescue Errno::ECONNREFUSED
            sleep 0.1
          else
            sock.close
            break
          end
        end
      end
      $authreq_log = File.open(process.io.stderr.path)
      $authreq_log.seek(0, :END)
      $authreq_handler = process
    end

    def authreq_log
      $authreq_log.read if $authreq_log
    end

    def authreq_http
      @authreq_http ||=
        begin
          authreq_handler # for side effects
          Net::HTTP.start('127.0.0.1', Spec::CARDEA_AUTHREQ_PORT)
        end
    end

    def teardown
      $authreq_log.seek(0, :END) if $authreq_log
      @authreq_http.finish if @authreq_http
      @authreq_http = nil
    end

    private
  end
end
