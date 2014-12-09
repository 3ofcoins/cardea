require 'spec_helper'

require 'net/http'
require 'uri'

module Cardea
  describe 'nginx integration' do
    let(:token) { Cardea::Token.new('ladmin', g: %w(foo bar)) }
    let(:nginx) { Spec::Endpoint[:nginx] }

    it 'bounces unauthenticated connections' do
      res = nginx.get('/')
      expect { res.code == '302' }
      expect { res['location'].start_with?('http://example.com/') }
    end

    it 'allows exempt locations' do
      res = nginx.get('/config/gatekeeper.rb')
      expect { res.code == '200' }
      expect { res.body.include? 'CARDEA_SECRET' }
      expect { res['x-cardea-user'].nil? }
    end

    it 'accepts modern cookie' do
      res = nginx.get('/',
        'Cookie' => "ca=#{token.cookie(secret, 'MiniTest')}",
        'User-Agent' => 'MiniTest')
      expect { res.code == '200' }
      expect { res.body.include? 'She is the goddess of the hinge' }
      expect { res['x-cardea-user'] == 'ladmin' }
    end

    # def expect_success(res)
    #   expect { res.code == '200' }
    #   expect { res['x-cardea-user'] == 'ladmin' }
    #   expect { res.get_fields('x-cardea-groups') == %w(foo bar) }
    #   expect { authreq_log.include?('ALLOW ladmin[foo,bar]') }
    # end

    # it 'lets modern cookie through' do
    #   expect_success authreq(
    #     'Cookie' => "ca=#{token.cookie(secret, 'MiniTest')}",
    #     'X-Cardea-Hmac-Extra' => 'MiniTest',)
    # end

    # it 'lets legacy cookie through' do
    #   expect_success authreq(
    #     'Cookie' => "ca=#{token.legacy.cookie(secret, 'MiniTest')}",
    #     'X-Cardea-Hmac-Extra' => 'MiniTest')
    # end

    # it 'lets modern http auth through' do
    #   expect_success authreq(
    #     'Authorization' => token.basic_auth(secret, 'MiniTest'),
    #     'X-Cardea-Hmac-Extra' => 'MiniTest')
    # end
  end
end
