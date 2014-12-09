require 'cgi'
require 'time'

require 'hashie'

require 'cardea/helpers'

module Cardea
  class Token < Hashie::Mash
    attr_reader :user

    def initialize(user, *args, &block)
      if user.is_a?(Token)
        # Token <-> Token::Odin conversion
        fail ArgumentError, 'too much args' unless args.empty?
        args = [user]
        user = user.user
      end

      if user !~ /^[a-z0-9_.-]+$/
        fail ArgumentError, "Invalid username #{user.inspect}"
      end
      @user = user
      super(*args, &block)
      self['g'] = Array(self['g'])
      self.t = self['t'] || Time.now.to_i
    end

    def t=(value)
      self['t'] = value.to_i
    end

    def payload
      params = []
      keys.sort.each do |k|
        case v = self[k]
        when Enumerable
          v.each do |v1|
            params << "#{k}=#{CGI.escape(v1.to_s)}"
          end
        else
          params << "#{k}=#{CGI.escape(v.to_s)}"
        end
      end
      params.join('&')
    end

    def to_s
      "#{user}:#{payload}"
    end

    def cardea_glue
      '#'
    end

    def hmac(secret, *hmac_extras)
      encode_hmac(
        OpenSSL::HMAC.digest(OpenSSL::Digest::SHA256.new, secret, [
          to_s,
          Helpers.b64(hmac_extras.map(&:to_s).join("\r\n")),
        ].join(cardea_glue)))
    end

    def encode_hmac(bin_hmac)
      Helpers.b64(bin_hmac)
    end

    def cookie(secret, *hmac_extras)
      [to_s, hmac(secret, *hmac_extras)].join(cardea_glue)
    end

    def basic_auth(secret, hmac_extras)
      "Basic #{Base64.strict_encode64(cookie(secret, hmac_extras))}"
    end

    def legacy
      Odin[self]
    end

    class Odin < Token
      def to_s
        [Helpers.b64(user), Helpers.b64(g.join(',')), t.to_s].join(',')
      end

      def encode_hmac(bin_hmac)
        Digest.hexencode(bin_hmac)
      end

      def cardea_glue
        ','
      end

      def basic_auth(_secret, _hmac_extras)
        fail NotImplementedError
      end
    end
  end
end
