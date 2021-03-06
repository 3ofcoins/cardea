# coding: utf-8
lib = File.expand_path('../lib', __FILE__)
$LOAD_PATH.unshift(lib) unless $LOAD_PATH.include?(lib)
require 'cardea/version'

Gem::Specification.new do |spec|
  spec.name          = 'cardea'
  spec.version       = Cardea::VERSION
  spec.authors       = ['Maciej Pasternacki']
  spec.email         = ['maciej@3ofcoins.net']
  spec.summary       = 'A cookie-based single sign-on system'
  # spec.description   = %q{TODO: Write a longer description. Optional.}
  spec.homepage      = 'https://github.com/3ofcoins/cardea'
  spec.license       = 'MIT'

  spec.files         = %w(LICENSE.txt
                          README.md
                          cardea.gemspec
                          lib/cardea.rb
                          lib/cardea/helpers.rb
                          lib/cardea/token.rb
                          lib/cardea/version.rb)
  spec.executables   = spec.files.grep(/^bin\//) { |f| File.basename(f) }
  spec.test_files    = [] # Tests include integration, they're not in gem
  spec.require_paths = ['lib']

  spec.add_dependency 'hashie', '~> 3.3'
end
