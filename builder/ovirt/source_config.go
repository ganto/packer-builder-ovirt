package ovirt

import (
	"errors"
	"fmt"
	"log"
	"sort"

	"github.com/google/uuid"
	"github.com/hashicorp/packer/template/interpolate"
)

// SourceConfig contains the various source properties for an oVirt image
type SourceConfig struct {
	Cluster string `mapstructure:"cluster"`

	SourceType string `mapstructure:"source_type"`

	SourceTemplateName    string `mapstructure:"source_template_name"`
	SourceTemplateVersion int    `mapstructure:"source_template_version"`
	SourceTemplateID      string `mapstructure:"source_template_id"`
}

// Prepare performs basic validation on the SourceConfig
func (c *SourceConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error
	// Supported source types must be added in alphabetical order
	validSourceTypes := []string{"template"}

	if c.Cluster == "" {
		c.Cluster = "Default"
	}

	if c.SourceType == "" {
		c.SourceType = "template"
		log.Printf("Using default source_type: %s", c.SourceType)
	}
	i := sort.SearchStrings(validSourceTypes, c.SourceType)
	if validSourceTypes[i] != c.SourceType {
		errs = append(errs, fmt.Errorf("Invalid source_type: %s", c.SourceType))
	}

	if (c.SourceType == "template") {
		if (c.SourceTemplateName != "") && (c.SourceTemplateVersion < 1) {
			c.SourceTemplateVersion = 1
			log.Printf("Using default source_template_version: %d", c.SourceTemplateVersion)
		}
		if (c.SourceTemplateID != "") {
			if _, err := uuid.Parse(c.SourceTemplateID); err != nil {
				errs = append(errs, fmt.Errorf("Invalid source_template_id: %s", c.SourceTemplateID))
			}
		}
		if (c.SourceTemplateName != "") && (c.SourceTemplateID != "") {
			errs = append(errs, errors.New("Conflict: Set either source_template_name or source_template_id"))
		}
	}

	// Required configurations that will display errors if not set
	if (c.SourceType == "template") && (c.SourceTemplateName == "") && (c.SourceTemplateID == "") {
		errs = append(errs, errors.New("source_template_name or source_template_id must be specified"))
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
