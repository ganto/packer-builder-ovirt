package main

import (
	"github.com/ganto/packer-builder-ovirt/builder/ovirt"
	"github.com/hashicorp/packer/packer/plugin"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterBuilder(new(ovirt.Builder))
	server.Serve()
}
