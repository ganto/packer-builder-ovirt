package ovirt

import (
	"errors"
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

	Comm communicator.Config `mapstructure:",squash"`

	Cluster       string `mapstructure:"cluster"`
	StorageDomain string `mapstructure:"storage_domain"`

	SourceTemplate        string `mapstructure:"source_template"`
	SourceTemplateVersion int    `mapstructure:"source_template_version"`

	VMName      string       `mapstructure:"vm_name"`
	Description string       `mapstructure:"description"`
	Memory      int          `mapstructure:"memory"`
	Cores       int          `mapstructure:"cores"`
	Sockets     int          `mapstructure:"sockets"`
	OS          string       `mapstructure:"os"`
	NICs        []nicConfig  `mapstructure:"network_adapters"`
	Disks       []diskConfig `mapstructure:"disks"`
	IPAddress   string       `mapstructure:"address"`
	Netmask     string       `mapstructure:"netmask"`
	Gateway     string       `mapstructure:"gateway"`

	TemplateName        string `mapstructure:"template_name"`
	TemplateDescription string `mapstructure:"template_description"`

	ctx interpolate.Context
}

type nicConfig struct {
	Model   string `mapstructure:"model"`
	Profile string `mapstructure:"profile"`
}
type diskConfig struct {
	Type            string `mapstructure:"type"`
	StoragePool     string `mapstructure:"storage_pool"`
	StoragePoolType string `mapstructure:"storage_pool_type"`
	Size            string `mapstructure:"disk_size"`
	CacheMode       string `mapstructure:"cache_mode"`
	DiskFormat      string `mapstructure:"format"`
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

	// Defaults
	if c.Cluster == "" {
		c.Cluster = "Default"
	}
	if c.StorageDomain == "" {
		c.StorageDomain = "data"
	}
	if c.SourceTemplateVersion < 1 {
		c.SourceTemplateVersion = 1
	}
	if c.VMName == "" {
		// Default to packer-[time-ordered-uuid]
		c.VMName = fmt.Sprintf("packer-%s", uuid.TimeOrderedUUID())
	}
	if c.Memory < 16 {
		log.Printf("Memory %d is too small, using default: 512", c.Memory)
		c.Memory = 512
	}
	if c.Cores < 1 {
		log.Printf("Number of cores %d is too small, using default: 1", c.Cores)
		c.Cores = 1
	}
	if c.Sockets < 1 {
		log.Printf("Number of sockets %d is too small, using default: 1", c.Sockets)
		c.Sockets = 1
	}
	if c.OS == "" {
		log.Printf("OS not set, using default 'other'")
		c.OS = "other"
	}
	for idx := range c.NICs {
		if c.NICs[idx].Model == "" {
			log.Printf("NIC %d model not set, using default 'virtio'", idx)
			c.NICs[idx].Model = "virtio"
		}
	}
	for idx := range c.Disks {
		if c.Disks[idx].Type == "" {
			log.Printf("Disk %d type not set, using default 'virtio'", idx)
			c.Disks[idx].Type = "scsi"
		}
		if c.Disks[idx].Size == "" {
			log.Printf("Disk %d size not set, using default '20G'", idx)
			c.Disks[idx].Size = "20G"
		}
		if c.Disks[idx].CacheMode == "" {
			log.Printf("Disk %d cache mode not set, using default 'none'", idx)
			c.Disks[idx].CacheMode = "none"
		}
	}
	if c.Netmask == "" {
		c.Netmask = "255.255.255.0"
		log.Printf("Set default netmask to %s", c.Netmask)
	}

	errs = packer.MultiErrorAppend(errs, c.Comm.Prepare(&c.ctx)...)

	// Required configurations that will display errors if not set
	if c.SourceTemplate == "" {
		errs = packer.MultiErrorAppend(errs, errors.New(fmt.Sprintf("source_template must be specified")))
	}
	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	packer.LogSecretFilter.Set(c.Password)
	return c, nil, nil
}
