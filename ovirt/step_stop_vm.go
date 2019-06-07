package ovirt

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

type stepStopVM struct{}

func (s *stepStopVM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	conn := state.Get("conn").(*ovirtsdk4.Connection)
	vmID := state.Get("vm_id").(string)

	ui.Say(fmt.Sprintf("Stopping VM: %s ...", vmID))
	_, err := conn.SystemService().
		VmsService().
		VmService(vmID).
		Stop().
		Send()
	if err != nil {
		err = fmt.Errorf("Error stopping VM: %s", err)
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Waiting for VM to stop: %s ...", vmID))
	stateChange := StateChangeConf{
		Pending:   []string{string(ovirtsdk4.VMSTATUS_UP)},
		Target:    []string{string(ovirtsdk4.VMSTATUS_DOWN)},
		Refresh:   VMStateRefreshFunc(conn, vmID),
		StepState: state,
	}
	if _, err := WaitForState(&stateChange); err != nil {
		err := fmt.Errorf("Error waiting for VM (%s) to stop: %s", vmID, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepStopVM) Cleanup(state multistep.StateBag) {}
