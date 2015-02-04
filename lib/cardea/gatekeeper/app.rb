require 'cgi'
require 'haml'
require 'sinatra'

module Cardea
  module Gatekeeper
    class App < Sinatra::Base
      # General Sinatra settings

      enable :sessions

      set :root, File.expand_path(File.join(
          File.dirname(__FILE__), '../../..'))

      # Cardea settings

      set :cardea_secret, ENV['CARDEA_SECRET']
      set :company_name, nil
      set :company_url, nil
      set :cookie_domain, nil
      set :cookie_name, 'ca'
      set :cookie_secure, nil
      set :login_href, 'https://github.com/3ofcoins/cardea/'
      set :login_text, 'Configure Me, Please'
      set :odin_cookie_name, nil
      set :debug_auth, false

      def request_token
        return unless request.cookies.include?(settings.cookie_name)
        @token ||= Cardea.parse(request.cookies[settings.cookie_name], settings.cardea_secret, *hmac_extras)
      rescue => e
        # FIXME: proper log?
        puts "ERROR parsing cookie: #{e}"
      end

      get '/' do
        if request_token
          haml :landing
        else
          login_url = '/login'
          if params[:return_to] && params[:return_to] != ''
            login_url << "?return_to=#{CGI.escape(params[:return_to])}"
          end
          redirect url(login_url), 302
        end
      end

      get '/logout' do
        logout
      end

      get '/login' do
        if params[:return_to] && params[:return_to] != ''
          session[:return_to] = params[:return_to]
        elsif settings.odin_cookie_name && params[:ref] && params[:ref] != ''
          session[:return_to] = params[:ref]
        else
          session.delete(:return_to)
        end
        haml :login
      end

      def cookie_parameters
        {
          domain: settings.cookie_domain || ".#{request.host}",
          path: '/',
          secure: settings.cookie_secure.nil? ? request.ssl? : settings.cookie_secure,
          httponly: true,
        }
      end
      private :cookie_parameters

      def hmac_extras
        [ request.env['HTTP_USER_AGENT'] ]
      end

      def login(username, meta={})
        return logout unless username # FIXME: show something?
        tk = Cardea.token(username, meta)
        # TODO: expires (smart); ¿max_age?
        response.set_cookie(settings.cookie_name, cookie_parameters.merge(value: tk.cookie(settings.cardea_secret, *hmac_extras)))
        response.set_cookie(settings.odin_cookie_name, cookie_parameters.merge(value: tk.legacy.cookie(settings.cardea_secret, *hmac_extras))) if settings.odin_cookie_name
        redirect session.delete(:return_to) || url('/'), 303
      end

      def logout
        response.delete_cookie(settings.cookie_name, cookie_parameters)
        response.delete_cookie(settings.odin_cookie_name, cookie_parameters) if settings.odin_cookie_name
        session.delete(:return_to)
        redirect to('/login'), 303
      end

      # OmniAuth Integration

      def omniauth_username(auth)
        auth['uid']
      end

      def omniauth_extras(auth)
        {}
      end

      def self.omniauth_builder(&block)
        require 'omniauth'
        require 'sinatra/multi_route'

        use OmniAuth::Builder, &block
        register Sinatra::MultiRoute

        route :get, :post, '/auth/:name/callback' do
          auth = request.env['omniauth.auth']
          if settings.debug_auth
            require 'pp'
            pp auth
          end
          login(omniauth_username(auth), omniauth_extras(auth))
        end
      end

      def self.omniauth(provider_name, *args)
        omniauth_builder do
          provider provider_name, *args
        end
        set :login_href, "/auth/#{provider_name}"
        set :login_text, "Sign in"
      end
    end
  end
end
