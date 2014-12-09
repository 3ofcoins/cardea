# -*- mode: ruby; coding: utf-8 -*-
# rubocop:disable GlobalVars

require 'rubygems'
require 'bundler/setup'

ENV['COMMAND_NAME'] = 'spec'
ARGV.reject! do |arg|
  arg =~ /^--command-name=/ &&
    ENV['COMMAND_NAME'] = Regexp.last_match.post_match
end

require 'simplecov' if ENV['COVERAGE']

require 'minitest/autorun'
require 'minitest/spec'
require 'minitest/reporters'
require 'minitest/pride' if $stdout.tty?
require 'wrong'
require 'childprocess_helper'

Minitest::Reporters.use! Minitest::Reporters::SpecReporter.new
Wrong.config.alias_assert :expect, override: true

require 'cardea'

class Minitest::Spec
  include ::Wrong::Assert
  include ::Wrong::Helpers

  def increment_assertion_count
    self.assertions += 1
  end

  def failure_class
    Minitest::Assertion
  end

  include Cardea::Spec

  let(:token) { Cardea::Token.new('ladmin', g: %w(foo bar)) }
  let(:authreq) { Endpoint[:authreq] }
  let(:nginx) { Endpoint[:nginx] }
end
