require 'spec_helper'

# rubocop:disable Metrics/LineLength

module Cardea
  describe Token do
    let(:token) { Cardea.token('john.doe', g: %w(foo bar), t: 23) }

    it 'rejects invalid username' do
      expect { rescuing { Cardea.token('!@#$%^&*()') }.to_s.include? 'Invalid username' }
    end

    it 'serializes to modern and legacy format' do
      expect { token.to_s == 'john.doe:g=foo&g=bar&t=23' }
      expect { token.legacy.to_s == 'am9obi5kb2U,Zm9vLGJhcg,23' }
    end

    it 'generates HMAC for tokens' do
      expect { token.hmac(secret, 'MiniTest') == 'wnZvtUl1Z_STsuQcIYuH6nht9MBSHEQi5ORlXpMvFSM' }
      expect { token.legacy.hmac(secret, 'MiniTest') == 'dc3ebee54bf2975dd98782af9b5065d9d4bee364c9c1ef4cf74fc38ced34488b' }
    end

    it 'generates modern auth cookies' do
      expect { token.cookie(secret, 'MiniTest') == 'john.doe:g=foo&g=bar&t=23#wnZvtUl1Z_STsuQcIYuH6nht9MBSHEQi5ORlXpMvFSM' }
      expect { token.legacy.cookie(secret, 'MiniTest') == 'am9obi5kb2U,Zm9vLGJhcg,23,dc3ebee54bf2975dd98782af9b5065d9d4bee364c9c1ef4cf74fc38ced34488b' }
    end

    it 'generates basic auth header' do
      expect { token.basic_auth(secret, 'MiniTest') == 'Basic am9obi5kb2U6Zz1mb28mZz1iYXImdD0yMyN3blp2dFVsMVpfU1RzdVFjSVl1SDZuaHQ5TUJTSEVRaTVPUmxYcE12RlNN' }
      expect { rescuing { token.legacy.basic_auth(secret, 'MiniTest') }.is_a? NotImplementedError }
    end
  end
end
