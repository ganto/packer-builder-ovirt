package ovirt

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

// The unique id for the builder
const BuilderId = "ganto.ovirt"

type Builder struct {
	config Config
	runner multistep.Runner
}

var pluginVersion = "0.0.1"

func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	c, warnings, errs := NewConfig(raws...)
	if errs != nil {
		return warnings, errs
	}
	b.config = *c

	return nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	var err error

	conn, err := ovirtsdk4.NewConnectionBuilder().
		URL(b.config.OvirtURL.String()).
		Username(b.config.Username).
		Password(b.config.Password).
		Insecure(b.config.SkipCertValidation).
		Compress(true).
		Timeout(time.Second * 10).
		Build()
	if err != nil {
		return nil, fmt.Errorf("oVirt: Connection failed, reason: %s", err.Error())
	}

	defer conn.Close()

	log.Printf("Successfully connected to %s\n", b.config.OvirtURL.String())

	// Set up the state
	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)
	state.Put("conn", conn)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Build the steps
	steps := []multistep.Step{
		&stepKeyPair{
			Debug:        b.config.PackerDebug,
			Comm:         &b.config.Comm,
			DebugKeyPath: fmt.Sprintf("ovirt_%s.pem", b.config.PackerBuildName),
		},
		&stepCreateInstance{
			Ctx:   b.config.ctx,
			Debug: b.config.PackerDebug,
		},
		&stepSetupInitialRun{
			Debug: b.config.PackerDebug,
			Comm:  &b.config.Comm,
		},
		&communicator.StepConnect{
			Config:    &b.config.Comm,
			Host:      commHost,
			SSHConfig: b.config.Comm.SSHConfigFunc(),
		},
		&common.StepProvision{},
		&common.StepCleanupTempKeys{
			Comm: &b.config.Comm,
		},
		&stepStopVM{},
	}

	// To use `Must` methods, you should recover it if panics
	defer func() {
		if err := recover(); err != nil {
			fmt.Errorf("oVirt: Panics occurs, try the non-Must methods to find the reason")
		}
	}()

	// Configure the runner and run the steps
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If there are no images, then just return
	if _, ok := state.GetOk("image"); !ok {
		return nil, nil
	}

	// Build the artifact and return it
	artifact := &Artifact{
		templateID: 42,
	}

	return artifact, nil
}
