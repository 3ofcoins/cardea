require 'spec_helper'

require 'net/http'
require 'uri'

module Cardea
  describe 'nginx integration' do
    let(:token) { Cardea::Token.new('ladmin', g: %w(foo bar)) }

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

    def expect_success(res)
      expect { res.code == '200' }
      expect { res.body.include? 'She is the goddess of the hinge' }
      expect { res['x-cardea-user'] == 'ladmin' }
      expect { authreq.log.include?('ALLOW ladmin[foo,bar]') }
    end

    it 'accepts modern cookie' do
      res = nginx.get('/',
        'Cookie' => "ca=#{token.cookie('SWORDFISH', 'MiniTest')}",
        'User-Agent' => 'MiniTest')
      expect_success(res)
    end

    it 'acepts legacy cookie' do
      expect_success nginx.get('/',
        'Cookie' => "ca=#{token.legacy.cookie('SWORDFISH', 'MiniTest')}",
        'User-Agent' => 'MiniTest',)
    end

    it 'accepts modern http auth' do
      expect_success nginx.get('/',
        'Authorization' => token.basic_auth('SWORDFISH', 'MiniTest'),
        'User-Agent' => 'MiniTest')
    end
  end
end
