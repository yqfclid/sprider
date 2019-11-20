GOPATH := $(shell pwd)

all: dingdian

dingdian:
	GOPATH=$(GOPATH) go get -d
	GOPATH=$(GOPATH) go build -o $@

clean:
	GOPATH=$(GOPATH) go clean
	${RM} -r pkg/ src/

.PHONY: dingdian
