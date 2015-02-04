# coding: utf-8
lib = File.expand_path('../lib', __FILE__)
$LOAD_PATH.unshift(lib) unless $LOAD_PATH.include?(lib)
require 'cardea/version'

Gem::Specification.new do |spec|
  spec.name          = 'cardea-gatekeeper'
  spec.version       = Cardea::VERSION
  spec.authors       = ['Maciej Pasternacki']
  spec.email         = ['maciej@3ofcoins.net']
  spec.summary       = 'A cookie-based single sign-on system'
  # spec.description   = %q{TODO: Write a longer description. Optional.}
  spec.homepage      = 'https://github.com/3ofcoins/cardea'
  spec.license       = 'MIT'

  spec.files         = %w(LICENSE.txt
                          README.md
                          cardea-gatekeeper.gemspec
                          config/gatekeeper.rb
                          lib/cardea/gatekeeper.rb)
  # lib/cardea/gatekeeper/**
  spec.executables   = spec.files.grep(/^bin\//) { |f| File.basename(f) }
  spec.test_files    = [] # Tests include integration, they're not in gem
  spec.require_paths = ['lib']

  spec.add_dependency 'cardea', Cardea::VERSION
  spec.add_dependency 'erubis',          '~> 2.7', '>= 2.7.0'
  spec.add_dependency 'sinatra',         '~> 1.4', '>= 1.4.5'
  spec.add_dependency 'sinatra-contrib', '~> 1.4', '>= 1.4.2'
  spec.add_dependency 'tinyconfig',      '~> 0.1', '>= 0.1.1'
  spec.add_dependency 'omniauth',        '~> 1.2', '>= 1.2.2'
  spec.add_dependency 'haml',            '~> 4.0', '>= 4.0.6'
end
