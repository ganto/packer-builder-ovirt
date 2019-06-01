package ovirt

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

type stepShutdownInstance struct{}

func (s *stepShutdownInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	conn := state.Get("conn").(*ovirtsdk4.Connection)
	vmID := state.Get("vm_id").(string)

	ui.Say("Shutting down VM...")
	_, err := conn.SystemService().
		VmsService().
		VmService(vmID).
		Stop().
		Send()
	if err != nil {
		err = fmt.Errorf("Error stopping VM: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Waiting for VM to become down..."))
	stateChange := StateChangeConf{
		Pending:   []string{string(ovirtsdk4.VMSTATUS_UP)},
		Target:    []string{string(ovirtsdk4.VMSTATUS_DOWN)},
		Refresh:   VMStateRefreshFunc(conn, vmID),
		StepState: state,
	}
	_, err = WaitForState(&stateChange)
	if err != nil {
		err := fmt.Errorf("Failed waiting for VM (%s) to become down: %s", vmID, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Message("VM has been shutdown!")

	return multistep.ActionContinue
}

// Cleanup any resources that may have been created during the Run phase.
func (s *stepShutdownInstance) Cleanup(state multistep.StateBag) {
	// Nothing to cleanup for this step.
}
