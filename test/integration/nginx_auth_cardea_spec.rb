require 'spec_helper'

require 'net/http'
require 'uri'

module Cardea
  describe 'nginx-auth-cardea' do
    let(:token) { Cardea::Token.new('ladmin', g: %w(foo bar)) }

    def authreq(headers = {})
      req = Net::HTTP::Get.new('/')
      headers.each do |hdr, val|
        req[hdr] = val
      end
      authreq_http.request(req)
    end

    def expect_success(res)
      expect { res.code == '200' }
      expect { res['x-cardea-user'] == 'ladmin' }
      expect { res.get_fields('x-cardea-groups') == %w(foo bar) }
      expect { authreq_log.include?('ALLOW ladmin[foo,bar]') }
    end

    it 'lets modern cookie through' do
      expect_success authreq(
        'Cookie' => "ca=#{token.cookie(secret, 'MiniTest')}",
        'X-Cardea-Hmac-Extra' => 'MiniTest',)
    end

    it 'lets legacy cookie through' do
      expect_success authreq(
        'Cookie' => "ca=#{token.legacy.cookie(secret, 'MiniTest')}",
        'X-Cardea-Hmac-Extra' => 'MiniTest')
    end

    it 'lets modern http auth through' do
      expect_success authreq(
        'Authorization' => token.basic_auth(secret, 'MiniTest'),
        'X-Cardea-Hmac-Extra' => 'MiniTest')
    end
  end
end
