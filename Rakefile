require 'bundler/setup'

# require 'bundler/gem_tasks'
require 'rake/testtask'
require 'rubocop/rake_task'

# rubocop:disable Style/HashSyntax

GO_DIRS = `git ls-files -z`
  .split("\0")
  .grep(/\.go$/)
  .map { |f| File.dirname(File.join('.', f)) }
  .uniq

GO_TOOLS = %w(cover vet)
  .map { |cmd| "code.google.com/p/go.tools/cmd/#{cmd}" }

if ENV['CI']
  ENV['COVERAGE'] = '1'
  rm_rf 'report'
end

def cov?
  ENV['COVERAGE']
end

mkdir 'report'

namespace :go do
  desc 'Get Go dependencies'
  task :get do
    get_cmd = ['go get -t -v']
    get_cmd << '-u' if ENV['UPDATE']
    get_cmd.concat(GO_DIRS)
    get_cmd.concat(GO_TOOLS)
    sh get_cmd.join(' ')

    gobin = `go env GOPATH`
      .strip
      .split(':')
      .map { |dir| File.join(dir, 'bin') }
      .select { |dir| File.directory?(dir) }
    ENV['PATH'] = "#{gobin.join(':')}:#{ENV['PATH']}"
  end

  desc 'Run GoConvey tests'
  task :convey => :get do
    test_cmd = ['go test']
    test_cmd << '-v' if ENV['VERBOSE']
    test_cmd << '-coverprofile=report/gocov.txt' if cov?
    sh test_cmd.join(' ')

    sh 'go tool cover -func=report/gocov.txt' if cov?
  end

  desc 'Vet the Go files'
  task :vet => :get do
    vet_cmd = ['go tool vet']
    vet_cmd << '-v' if ENV['VERBOSE']
    vet_cmd.concat(GO_DIRS)
    sh vet_cmd.join(' ')
  end

  desc 'Run Go tests'
  task :test => [:convey, :vet]

  desc 'Generate Go reports'
  task :report => [:test] do
    sh 'go tool cover -html=report/gocov.txt -o report/gocov.html' if cov?
  end
end

namespace :ruby do
  RuboCop::RakeTask.new

  Rake::TestTask.new :unit do |t|
    t.pattern = 'spec/unit/**_spec.rb'
    t.verbose = ENV['VERBOSE']
    t.options = '--command-name=spec'
  end

  Rake::TestTask.new :integration do |t|
    t.pattern = 'spec/integration/**_spec.rb'
    t.verbose = ENV['VERBOSE']
    t.options = '--command-name=integration'
  end

  desc 'Run Ruby tests'
  task :test => [:rubocop, :unit, :integration]

  desc 'Generate Ruby reports'
  task :report => [:test]
end

file 'report/index.html' do |t|
  File.write t.to_s, <<EOF
<!DOCTYPE html>
<meta charset="utf-8">
<title>Cardea Coverage</title>
<h1>Cardea Coverage Reports</h1>
<ul>
  <li><a href="gocov.html">GoCov</a> (Go)</li>
  <li><a href="simplecov/index.html">SimpleCov</a> (Ruby)</li>
</ul>
EOF
end

desc 'Run all tests'
task :test => ['go:test', 'ruby:test']

desc 'Generate all reports'
task :report => ['go:report', 'ruby:report', 'report/index.html']

task :default => :report