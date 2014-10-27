require 'spec_helper'

# rubocop:disable Metrics/LineLength

module Cardea
  describe Config do
    it 'generates sane HMAC' do
      config = Config.new('swordfish')
      expect { config.hmac_for('YV91c2Vy,c29tZSxncm91cHM,23', 'GoConvey', ',') == '686a4945039bda5e0a1e3be1f6e53d5aceee104446fd2f5f47d86dec21421dca' }
      expect { config.hmac_for('maciej:g=admin&t=1396349947', 'For instance, User-Agent header') == 'bac9f78c8b06ea96295d976d90e5094c378bad2539693b081a4da22880068ba4' }
      expect { config.hmac_for('maciej:t=23', %w(foo bar)) == '177b7eaf3b29fdb09e15beb964304d86c269b26775d0a90b925d0913eb264eca' }
    end

    it 'serializes to modern and legacy format' do
      token = Token.new('john.doe', g: %w(foo bar), t: 23)
      expect { token.to_s == 'john.doe:g=foo&g=bar&t=23' }
      expect { token.to_odin == 'am9obi5kb2U,Zm9vLGJhcg,23' }
    end
  end
end
