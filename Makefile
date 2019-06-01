.PHONY: default
default: packer-builder-ovirt

.PHONY: clean
clean:
	rm -f packer-builder-ovirt

.PHONY: deps
deps:
	go get -v

packer-builder-ovirt: *.go ovirt/*.go
	gofmt -w *.go ovirt/*.go
	GOOS=linux GOARCH=amd64 go build -v -o packer-builder-ovirt

test: packer-builder-ovirt
	packer validate template.json
	packer build template.json
