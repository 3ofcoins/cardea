all: generate

generate: regexp.go

regexp.go: script/compose_regexp.go
	go run $< > $@
