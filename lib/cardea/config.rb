require 'openssl'

require 'cardea/helpers'

module Cardea
  class Config
    def initialize(secret)
      @secret = secret
    end

    def hmac_for(obj, hmac_extras, glue = '#')
      message = [
        obj.to_s,
        process_extras(hmac_extras)
      ].join(glue)
      OpenSSL::HMAC.hexdigest(OpenSSL::Digest::SHA256.new, secret, message)
    end

    private

    attr_reader :secret

    def process_extras(hmac_extras)
      extras_s = if hmac_extras.respond_to?(:map)
                   hmac_extras.map(&:to_s).join("\r\n")
                 else
                   hmac_extras.to_s
                 end
      Helpers.b64(extras_s)
    end
  end
end
