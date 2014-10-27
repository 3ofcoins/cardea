require 'tempfile'
require 'timeout'

require 'childprocess'

ChildProcess.posix_spawn = true

module Cardea
  module Spec
    def childprocess(*commandline)
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

    def nginx_auth_cardea
      spawn_nginx_auth_cardea_handler unless @nginx_auth_cardea
      @nginx_auth_cardea
    end

    def nginx_auth_cardea_handler
      @nginx_auth_cardea_handler ||= spawn_nginx_auth_cardea_handler
    end

    private

    def spawn_nginx_auth_cardea_handler
      process = childprocess('./target/nginx-auth-cardea',
        '-listen', '127.0.0.1:0',
        '-secret', 'SWORDFISH',
        '-literal-secret')
      Timeout.timeout 10 do
        loop do
          fail RuntimeError, 'nginx-auth-cardea died' unless process.alive?
          if File.read(process.io.stderr.path) =~ /Listening on (.*:\d+)$/
            @nginx_auth_cardea = "http://#{$1}/"
            break
          else
            sleep 0.1
          end
        end
      end
      require 'pry' ; binding.pry
      process
    end
  end
end
