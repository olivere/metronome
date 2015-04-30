PACKAGE=github.com/olivere/metronome
.PHONY: build-server build-client

all: build-server build-client

build-server:
	go build $(PACKAGE)/cmd/metronomed

build-client:
	go build $(PACKAGE)/cmd/metronome

deps:
	go get github.com/gorilla/websocket
	go get github.com/rcrowley/go-metrics
	go get github.com/BurntSushi/toml
