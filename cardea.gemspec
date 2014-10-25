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

  spec.files         = `git ls-files -z`.split("\x0")
  spec.executables   = spec.files.grep(/^bin\//) { |f| File.basename(f) }
  spec.test_files    = spec.files.grep(/^(test|spec|features)\//)
  spec.require_paths = ['lib']

  spec.add_development_dependency 'bundler', '~> 1.7'
  spec.add_development_dependency 'rake', '~> 10.0'
  spec.add_development_dependency 'rubocop', '~> 0.26'
  spec.add_development_dependency 'minitest', '~> 5.4'
  spec.add_development_dependency 'simplecov', '~> 0.9.1'
  spec.add_development_dependency 'wrong', '>= 0.7.1'
end
