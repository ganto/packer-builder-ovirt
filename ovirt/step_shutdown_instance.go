package ovirt

import (
	"context"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	//ovirtsdk4 "github.com/ovirt/go-ovirt"
)

type stepShutdownInstance struct{}

func (s *stepShutdownInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	//	conn := state.Get("conn").(*ovirtsdk4.Connection)
	//	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Shutting down instance...")
	ui.Message("Instance has been shutdown!")
	return multistep.ActionContinue
}

// Cleanup any resources that may have been created during the Run phase.
func (s *stepShutdownInstance) Cleanup(state multistep.StateBag) {
	// Nothing to cleanup for this step.
}
