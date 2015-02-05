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
      set :cookie_max_age, 18*3600
      set :login_href, 'https://github.com/3ofcoins/cardea/'
      set :login_text, 'Configure Me, Please'
      set :odin_cookie_name, nil
      set :debug_auth, false
      set :links, {}

      def request_token
        return unless request.cookies.include?(settings.cookie_name)
        @token ||= Cardea.parse(request.cookies[settings.cookie_name], settings.cardea_secret, *hmac_extras)
      rescue => e
        # FIXME: proper log?
        puts "ERROR parsing cookie: #{e}"
      end

      def return_to
        return params[:return_to] if params[:return_to] && params[:return_to] != ''
        return params[:ref] if settings.odin_cookie_name && params[:ref] && params[:ref] != ''
      end

      get '/login' do
        redirect(return_to || url('/')) if request_token
        session[:return_to] = return_to
        haml(:login, locals: { logout: false })
      end

      get '/' do
        if !request_token
          login_url = '/login'
          login_url << '?return_to=#{CGI::escape(return_to)}' if return_to
          redirect to(login_url)
        end
        haml(:index)
      end

      get '/logout' do
        logout
      end

      def cookie_parameters
        {
          domain: settings.cookie_domain || ".#{request.host}",
          path: '/',
          secure: settings.cookie_secure.nil? ? request.ssl? : settings.cookie_secure,
          httponly: true
        }
      end
      private :cookie_parameters

      def hmac_extras
        [ request.env['HTTP_USER_AGENT'] ]
      end

      def login(username, meta={})
        logout "Authorization Failure" unless username # FIXME: show something?

        tk = Cardea.token(username, meta)
        params = cookie_parameters.merge(
          max_age: settings.cookie_max_age,
          expires: Time.now + settings.cookie_max_age)
        response.set_cookie(settings.cookie_name, params.merge(
            value: tk.cookie(settings.cardea_secret, *hmac_extras)))
        if settings.odin_cookie_name
          response.set_cookie(settings.odin_cookie_name, params.merge(
              value: tk.legacy.cookie(settings.cardea_secret, *hmac_extras)))
        end

        redirect session.delete(:return_to) || url('/'), 303
      end

      def logout(message=nil)
        response.delete_cookie(settings.cookie_name, cookie_parameters)
        response.delete_cookie(settings.odin_cookie_name, cookie_parameters) if settings.odin_cookie_name
        session.delete(:return_to)
        halt haml(:login, locals: { message: message, logout: true })
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

        route :get, :post, '/auth/failure' do
          # ?message=csrf_detected&strategy=google_oauth2
          logout "Authentication Failure"
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
