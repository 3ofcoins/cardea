require 'cgi'
require 'time'

require 'hashie'

require 'cardea/helpers'

module Cardea
  class Token < Hashie::Mash
    attr_reader :user

    def initialize(user, *args, &block)
      if user !~ /^[a-z0-9_.-]+$/
        fail ArgumentError, "Invalid username #{user.inspect}"
      end
      @user = user
      super(*args, &block)
      self['g'] = Array(self['g'])
      self.t = self['t'] || Time.utcnow
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

    def to_odin
      [Helpers.b64(user), Helpers.b64(g.join(',')), t.to_s].join(',')
    end
  end
end
