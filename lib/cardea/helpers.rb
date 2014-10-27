require 'base64'

module Cardea
  module Helpers
    def self.b64(str)
      Base64.urlsafe_encode64(str).sub(/=*$/, '')
    end
  end
end
