require 'omniauth-google-oauth2'
require 'cardea/gatekeeper'

class Gatekeeper < Cardea::Gatekeeper::App
  set :cardea_secret, 'SWORDFISH'

  omniauth :google_oauth2,
           'GOOGLE_CLIENT_ID',
           'GOOGLE_CLIENT_SECRET'

  def omniauth_username(auth)
    # Let in only users @example.com. Use username part of their email
    # address as SSO username.
    return $1 if auth['info']['email'] =~ /^(.*)@example\.com$/
  end
end

run Gatekeeper.new
