require 'cardea/gatekeeper'

class Gatekeeper < Cardea::Gatekeeper::App
  set :cardea_secret, 'SWORDFISH'

  omniauth do
    provider :developer, :uid_field => :name
  end

  set :login_href, '/auth/developer'
  set :login_text, 'Pretend to log in'
  set :company_name, 'Three of Coins'
  set :company_url, 'http://3ofcoins.net/'
end

run Gatekeeper.new
