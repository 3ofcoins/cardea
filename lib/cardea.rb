require 'cardea/version'
require 'cardea/helpers'
require 'cardea/token'

module Cardea
  def self.token(*args)
    Token.new(*args)
  end
end
