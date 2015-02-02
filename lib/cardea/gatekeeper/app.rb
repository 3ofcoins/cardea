require 'erubis'
require 'sinatra'

module Cardea
  module Gatekeeper
    class App < Sinatra::Base
      set :erb, escape_html: true
      set :root, Config.root_dir

      def config
        @config ||= Config.load
      end

      def ref
        ref = params[:ref]
        ref = url('/') if !ref || ref == ''
        ref
      end

      get '/' do
        erb :landing
      end

      get '/login' do
        erb :login
      end

      post '/login' do
        # TODO: expires (smart); ¿max_age?
        response.set_cookie config.cardea_cookie,
                            value: "You are #{params[:username]}",
                            domain: ".#{request.host}",
                            secure: config.cookie_secure,
                            httponly: true
        if config.odin_compatible
          response.set_cookie config.odin_cookie,
                              value: "You are #{params[:username]}",
                              domain: ".#{request.host}",
                              secure: config.cookie_secure,
                              httponly: true
        end
        redirect ref, 303
      end
    end
  end
end
