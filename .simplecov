# -*- ruby -*-

SimpleCov.start do
  add_filter '/spec/' unless ENV['SPEC_COVERAGE']
  command_name ENV['COMMAND_NAME'] if ENV['COMMAND_NAME']
  coverage_dir 'target/report/simplecov'
  use_merging true
end
