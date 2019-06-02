package ovirt

import (
	"github.com/hashicorp/packer/helper/multistep"
)

func commHost(state multistep.StateBag) (string, error) {
	c := state.Get("config").(*Config)
	return c.IPAddress, nil
}
