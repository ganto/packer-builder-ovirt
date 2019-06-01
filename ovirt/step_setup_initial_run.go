package ovirt

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

type stepSetupInitialRun struct {
	Debug bool
	Comm  *communicator.Config
}

// Run executes the Packer build step that configures the initial run setup
func (s *stepSetupInitialRun) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	conn := state.Get("conn").(*ovirtsdk4.Connection)

	ui.Say("Setting up initial run...")

	vmID := state.Get("vm_id").(string)
	vmService := conn.SystemService().
		VmsService().
		VmService(vmID)

	initializationBuilder := ovirtsdk4.NewInitializationBuilder()
	if s.Comm.SSHUsername != "" {
		log.Printf("Set SSH user name: %s", s.Comm.SSHUsername)
		initializationBuilder.UserName(s.Comm.SSHUsername)
	}
	if string(s.Comm.SSHPublicKey) != "" {
		publicKey := s.Comm.SSHPublicKey
		//		publicKeyDer := x509.MarshalPKIXPublicKey(&publicKey)
		//		publicKeyBlk := pem.Block{Type: "PUBLIC KEY", Headers: nil, Bytes: privDer}
		log.Printf("Set authorized SSH key: %s", string(publicKey))
		initializationBuilder.AuthorizedSshKeys(string(publicKey))
	}
	initialization, err := initializationBuilder.Build()
	if err != nil {
		err = fmt.Errorf("Error setting up initial run: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	vmBuilder := ovirtsdk4.NewVmBuilder()
	vm, err := vmBuilder.Initialization(initialization).Build()
	if err != nil {
		err = fmt.Errorf("Error defining VM initialization: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Say("Starting virtual machine...")

	_, err = vmService.Start().
		UseCloudInit(true).
		Vm(vm).
		Send()
	if err != nil {
		err = fmt.Errorf("Error starting VM: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Waiting for VM to become ready (status up) ..."))
	stateChange := StateChangeConf{
		Pending:   []string{"wait_for_launch", "powering_up"},
		Target:    []string{string(ovirtsdk4.VMSTATUS_UP)},
		Refresh:   VMStateRefreshFunc(conn, vmID),
		StepState: state,
	}
	_, err = WaitForState(&stateChange)
	if err != nil {
		err := fmt.Errorf("Failed waiting for VM (%s) to become up: %s", vmID, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	ui.Message("VM successfully started!")

	return multistep.ActionContinue
}

// Cleanup any resources that may have been created during the Run phase.
func (s *stepSetupInitialRun) Cleanup(state multistep.StateBag) {
	// Nothing to cleanup for this step.
}
