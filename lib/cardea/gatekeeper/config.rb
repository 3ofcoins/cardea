require 'tinyconfig'

module Cardea
  module Gatekeeper
    class Config < TinyConfig
      option :secret, nil
      option :cardea_cookie, 'ca'
      option :odin_cookie, 'oa'
      option :odin_compatible, false

      def self.root_dir
        ::File.expand_path(::File.join(
            ::File.dirname(__FILE__), '../../..'))
      end

      def self.expand_path(path)
        ::File.join(root_dir, path)
      end

      def self.load
        config = new
        config.load(expand_path('config/gatekeeper.rb'))
        config.load(::ENV['CARDEA_CONFIG']) if ::ENV['CARDEA_CONFIG']
        fail unless config.secret # FIXME: descriptive error
        config
      end
    end
  end
end
