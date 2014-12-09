require 'spec_helper'

require 'net/http'
require 'uri'

module Cardea
  describe 'nginx-auth-cardea' do
    def expect_success(res)
      expect { res.code == '200' }
      expect { res['x-cardea-user'] == 'ladmin' }
      expect { res.get_fields('x-cardea-groups') == %w(foo bar) }
      expect { authreq.log.include?('ALLOW ladmin[foo,bar]') }
    end

    it 'lets modern cookie through' do
      expect_success authreq.get('/',
        'Cookie' => "ca=#{token.cookie('SWORDFISH', 'MiniTest')}",
        'X-Cardea-Hmac-Extra' => 'MiniTest')
    end

    it 'lets legacy cookie through' do
      expect_success authreq.get('/',
        'Cookie' => "ca=#{token.legacy.cookie('SWORDFISH', 'MiniTest')}",
        'X-Cardea-Hmac-Extra' => 'MiniTest')
    end

    it 'lets modern http auth through' do
      expect_success authreq.get('/',
        'Authorization' => token.basic_auth('SWORDFISH', 'MiniTest'),
        'X-Cardea-Hmac-Extra' => 'MiniTest')
    end
  end
end
