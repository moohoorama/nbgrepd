
nbgrepd:
	go generate ./...
	go build -o $@ ./cmd/*.go
