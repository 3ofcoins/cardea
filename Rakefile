require 'bundler/setup'

# require 'bundler/gem_tasks'
require 'rake/clean'
require 'rake/testtask'
require 'rubocop/rake_task'

# rubocop:disable Style/HashSyntax, Metrics/LineLength

GO_DIRS = `git ls-files -z`
          .split("\0")
          .grep(/\.go$/)
          .map { |f| File.dirname(File.join('.', f)) }
          .uniq

GO_TOOLS = %w(cover vet)
           .map { |cmd| "code.google.com/p/go.tools/cmd/#{cmd}" }

def cov?
  ENV['COVERAGE']
end

def verbose?
  ENV['VERBOSE']
end

task :clean do
  sh 'git clean -fdx target/'
end

if ENV['CI']
  ENV['COVERAGE'] = '1'
  Rake::Task[:clean].invoke
end

desc 'Get prerequisite libraries and such'
task :prereqs do
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

### Code Generation
###################

GENERATED = FileList.new

def generate_file(*args, &block)
  GENERATED << file(*args, &block)
end

generate_file 'regexp.go' => 'script/compose_regexp.go' do
  sh 'go run script/compose_regexp.go > regexp.go'
end

desc 'Generate secondary files'
task :generate => :prereqs
task :generate => GENERATED do
  sh "git diff --name-status --exit-code #{GENERATED}"
end

### Build Targets
#################

TARGETS = FileList.new

def build_file(*args, &block)
  TARGETS << file(*args, &block)
end

build_file 'target/nginx-auth-cardea' => FileList['*.go', 'nginx-auth-cardea/*.go'] do |t|
  sh "go build -o #{t} ./nginx-auth-cardea"
end

desc 'Build targets'
task :build => [:prereqs, :generate]
task :build => TARGETS

namespace :go do
  desc 'Run GoConvey tests'
  task :convey => :prereqs do
    test_cmd = ['go test']
    test_cmd << '-v' if verbose?
    test_cmd << '-coverprofile=target/report/gocov.txt' if cov?
    sh test_cmd.join(' ')

    sh 'go tool cover -func=target/report/gocov.txt' if cov?
  end

  desc 'Vet the Go files'
  task :vet => :prereqs do
    vet_cmd = ['go tool vet']
    vet_cmd << '-v' if verbose?
    vet_cmd.concat(GO_DIRS)
    sh vet_cmd.join(' ')
  end

  desc 'Run Go tests'
  task :test => [:convey, :vet]

  desc 'Generate Go reports'
  task :report => [:test] do
    sh 'go tool cover -html=target/report/gocov.txt -o target/report/gocov.html' if cov?
  end
end

desc 'Run whole workflow for Go'
task :go => ['go:test', 'go:report']

namespace :ruby do
  RuboCop::RakeTask.new

  Rake::TestTask.new :spec do |t|
    t.pattern = 'test/spec/**_spec.rb'
    t.options = '--command-name=spec'
    t.libs = %w(lib test/lib)
    t.verbose = verbose?
  end

  Rake::TestTask.new :integration do |t|
    t.pattern = 'test/integration/**_spec.rb'
    t.options = '--command-name=integration'
    t.libs = %w(lib test/lib)
    t.verbose = verbose?
  end

  desc 'Run Ruby tests'
  task :test => [:rubocop, :spec, :integration]

  desc 'Generate Ruby reports'
  task :report => [:test]
end

desc 'Run full Ruby workflow'
task :ruby => ['ruby:test', 'ruby:report']

desc 'Run all tests'
task :test => ['go:test', 'ruby:test']

desc 'Generate all reports'
task :report => ['go:report', 'ruby:report']

task :default => [:build, :test, :report]

task 'ruby:integration' => :build
