BUILD_ARG= --build-arg http_proxy=http://proxy.daumkakao.io:3128 --build-arg https_proxy=http://proxy.daumkakao.io:3128 
NBGREPD_REPO="idock.daumkakao.io/kage/nbgrepd"
RELEASE=$(shell date +%Y%m%d_%H)

all: nbgrepd

clean:
	rm nbgrepd

nbgrepd:
	go generate ./...
	go fmt ./...
	go build -o $@ ./cmd/*.go

build:
	-mkdir tmp
	GOOS=linux GOARCH=amd64 go build -o ./tmp/nbgrepd ./cmd/*.go
	bash -c 'docker build $(BUILD_ARG) -t $(NBGREPD_REPO):$(RELEASE) -f ./cmd/Dockerfile ./tmp;'
	bash -c 'docker push $(NBGREPD_REPO):$(RELEASE);'

govendor:
	-go mod init github.daumkakao.io/tscoke/nbgrepd
	-export GOPRIVATE=*.daumkakao.com;go mod vendor
	-export http_proxy=http://proxy.daumkakao.io:3128;export https_proxy=http://proxy.daumkakao.io:3128;export GOPRIVATE=*.daumkakao.com;go mod vendor
	-mkdir -p vendor/github.com/gogo/;cd vendor/github.com/gogo/;rm -rf protobuf; git clone https://github.com/gogo/protobuf.git;export http_proxy=http://proxy.daumkakao.io:3128;export https_proxy=http://proxy.daumkakao.io:3128;git clone https://github.com/gogo/protobuf.git

.PHONY: nbgrepd
