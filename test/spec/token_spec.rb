require 'spec_helper'

# rubocop:disable Metrics/LineLength

module Cardea
  describe Token do
    it 'rejects invalid username' do
      expect { rescuing { Token.new('!@#$%^&*()') }.to_s.include? 'Invalid username' }
    end

    it 'serializes to modern and legacy format' do
      token = Token.new('john.doe', g: %w(foo bar), t: 23)
      expect { token.to_s == 'john.doe:g=foo&g=bar&t=23' }
      expect { token.to_odin == 'am9obi5kb2U,Zm9vLGJhcg,23' }
    end
  end
end
