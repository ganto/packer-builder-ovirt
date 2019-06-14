package ovirt

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/hashicorp/packer/template/interpolate"
)

// AccessConfig contains the oVirt API access and authentication configuration
type AccessConfig struct {
	OvirtURLRaw        string `mapstructure:"ovirt_url"`
	OvirtURL           *url.URL
	SkipCertValidation bool   `mapstructure:"insecure_skip_tls_verify"`
	Username           string `mapstructure:"username"`
	Password           string `mapstructure:"password"`
}

// Prepare performs basic validation on the AccessConfig
func (c *AccessConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error

	if c.OvirtURLRaw == "" {
		c.OvirtURLRaw = os.Getenv("OVIRT_URL")
	}
	if c.Username == "" {
		c.Username = os.Getenv("OVIRT_USERNAME")
	}
	if c.Password == "" {
		c.Password = os.Getenv("OVIRT_PASSWORD")
	}

	// Required configurations that will display errors if not set
	if c.Username == "" {
		errs = append(errs, errors.New("username must be specified"))
	}
	if c.Password == "" {
		errs = append(errs, errors.New("password must be specified"))
	}
	if c.OvirtURLRaw == "" {
		errs = append(errs, errors.New("ovirt_url must be specified"))
	}

	var err error
	if c.OvirtURL, err = url.Parse(c.OvirtURLRaw); err != nil {
		errs = append(errs, fmt.Errorf("Could not parse ovirt_url: %s", err))
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
