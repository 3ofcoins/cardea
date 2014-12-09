# -*- ruby -*-

SimpleCov.start do
  add_filter '/test/' unless ENV['TEST_COVERAGE']
  command_name ENV['COMMAND_NAME'] if ENV['COMMAND_NAME']
  coverage_dir 'target/report/simplecov'
  use_merging true
end
