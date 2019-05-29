package ovirt

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	//	"time"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/bootcommand"
	"github.com/hashicorp/packer/common/uuid"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	//"github.com/mitchellh/mapstructure"
)

type Config struct {
	common.PackerConfig    `mapstructure:",squash"`
	common.HTTPConfig      `mapstructure:",squash"`
	bootcommand.BootConfig `mapstructure:",squash"`
	//	RawBootKeyInterval     string              `mapstructure:"boot_key_interval"`
	//	BootKeyInterval        time.Duration       ``
	Comm communicator.Config `mapstructure:",squash"`

	OvirtURLRaw        string `mapstructure:"ovirt_url"`
	OvirtURL           *url.URL
	SkipCertValidation bool   `mapstructure:"insecure_skip_tls_verify"`
	Username           string `mapstructure:"username"`
	Password           string `mapstructure:"password"`

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
	//	ISOFile string       `mapstructure:"iso_file"`
	//	Agent   bool         `mapstructure:"qemu_agent"`

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

	//	var md mapstructure.Metadata
	err := config.Decode(c, &config.DecodeOpts{
		//		Metadata:           &md,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		//		InterpolateFilter: &interpolate.RenderFilter{
		//			Exclude: []string{
		//				"boot_command",
		//			},
		//		},
	}, raws...)
	if err != nil {
		//        return nil, fmt.Errorf("Failed to mapstructure Config: %+v", err), err
		return nil, nil, err
	}

	var errs *packer.MultiError

	// Defaults
	if c.OvirtURLRaw == "" {
		c.OvirtURLRaw = os.Getenv("OVIRT_URL")
	}
	if c.Username == "" {
		c.Username = os.Getenv("OVIRT_USERNAME")
	}
	if c.Password == "" {
		c.Password = os.Getenv("OVIRT_PASSWORD")
	}
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

	errs = packer.MultiErrorAppend(errs, c.Comm.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.BootConfig.Prepare(&c.ctx)...)
	errs = packer.MultiErrorAppend(errs, c.HTTPConfig.Prepare(&c.ctx)...)

	// Required configurations that will display errors if not set
	if c.Username == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("username must be specified"))
	}
	if c.Password == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("password must be specified"))
	}
	if c.OvirtURLRaw == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("ovirt_url must be specified"))
	}
	if c.OvirtURL, err = url.Parse(c.OvirtURLRaw); err != nil {
		errs = packer.MultiErrorAppend(errs, errors.New(fmt.Sprintf("Could not parse ovirt_url: %s", err)))
	}
	if c.SourceTemplate == "" {
		errs = packer.MultiErrorAppend(errs, errors.New(fmt.Sprintf("source_template must be specified")))
	}
	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	packer.LogSecretFilter.Set(c.Password)
	return c, nil, nil
}
