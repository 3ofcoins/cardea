require 'cardea/gatekeeper'

class Gatekeeper < Cardea::Gatekeeper::App
  set :cardea_secret, 'SWORDFISH'
  omniauth :developer, :uid_field => :name
end

run Gatekeeper.new
