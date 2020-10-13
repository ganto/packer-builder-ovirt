//go:generate mapstructure-to-hcl2 -type Config
package ovirt

import (
	"fmt"
	"log"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	AccessConfig `mapstructure:",squash"`
	SourceConfig `mapstructure:",squash"`

	Comm communicator.Config `mapstructure:",squash"`

	VMName      string       `mapstructure:"vm_name"`
	IPAddress   string       `mapstructure:"address"`
	Netmask     string       `mapstructure:"netmask"`
	Gateway     string       `mapstructure:"gateway"`

	DiskName        string `mapstructure:"disk_name"`
	DiskDescription string `mapstructure:"disk_description"`

	ctx interpolate.Context
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := new(Config)

	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
	}, raws...)
	if err != nil {
		return nil, nil, err
	}

	// Accumulate any errors
	var errs *packer.MultiError
	errs = packer.MultiErrorAppend(errs, c.AccessConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.SourceConfig.Prepare(&c.ctx)...)

	if c.VMName == "" {
		// Default to packer-[time-ordered-uuid]
		c.VMName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}
	if c.DiskName == "" {
		c.DiskName = c.VMName
	}
	if c.Netmask == "" {
		c.Netmask = "255.255.255.0"
		log.Printf("Set default netmask to %s", c.Netmask)
	}

	errs = packer.MultiErrorAppend(errs, c.Comm.Prepare(&c.ctx)...)

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	packer.LogSecretFilter.Set(c.Password)
	return c, nil, nil
}
