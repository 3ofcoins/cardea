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
        "Would log in as <strong>#{params[:username]}</strong> and redirect to <strong>#{ref}</strong>"
      end
    end
  end
end
