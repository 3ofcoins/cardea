require 'spec_helper'

require 'net/http'
require 'uri'

module Cardea
  describe 'nginx-auth-cardea' do
    let(:token) { Cardea::Token.new('ladmin', g: %w(foo bar)) }
    let(:authreq) { Spec::Endpoint[:authreq] }

    def expect_success(res)
      expect { res.code == '200' }
      expect { res['x-cardea-user'] == 'ladmin' }
      expect { res.get_fields('x-cardea-groups') == %w(foo bar) }
      expect { authreq.log.read.include?('ALLOW ladmin[foo,bar]') }
    end

    it 'lets modern cookie through' do
      expect_success authreq.get('/',
        'Cookie' => "ca=#{token.cookie(secret, 'MiniTest')}",
        'X-Cardea-Hmac-Extra' => 'MiniTest')
    end

    it 'lets legacy cookie through' do
      expect_success authreq.get('/',
        'Cookie' => "ca=#{token.legacy.cookie(secret, 'MiniTest')}",
        'X-Cardea-Hmac-Extra' => 'MiniTest')
    end

    it 'lets modern http auth through' do
      expect_success authreq.get('/',
        'Authorization' => token.basic_auth(secret, 'MiniTest'),
        'X-Cardea-Hmac-Extra' => 'MiniTest')
    end
  end
end
