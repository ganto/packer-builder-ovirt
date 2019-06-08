package ovirt

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

type stepUpdateDisk struct{}

func (s *stepUpdateDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	conn := state.Get("conn").(*ovirtsdk4.Connection)
	vmID := state.Get("vm_id").(string)

	ui.Say("Updating disk properties ...")

	resp, err := conn.SystemService().
		VmsService().
		VmService(vmID).
		DiskAttachmentsService().
		List().
		Send()
	if err != nil {
		err = fmt.Errorf("Error listing disks of VM: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	das := resp.MustAttachments()

	d, _ := conn.FollowLink(das.Slice()[0].MustDisk())
	disk, ok := d.(*ovirtsdk4.Disk)
	if !ok {
		err = fmt.Errorf("Error getting disk of VM: '%s': %s", vmID, err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	diskID := disk.MustId()
	log.Printf("Disk identifier: %s", diskID)

	diskAttachmentService := conn.SystemService().
		VmsService().
		VmService(vmID).
		DiskAttachmentsService().
		AttachmentService(diskID)

	_, err = diskAttachmentService.Get().Send()
	if err != nil {
		err = fmt.Errorf("Error getting disk attachment '%s': %s", diskID, err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	diskBuilder := ovirtsdk4.NewDiskBuilder().
		Name(config.DiskName).
		Description(config.DiskDescription)

	log.Printf(fmt.Sprintf("Disk name: %s", config.DiskName))
	log.Printf(fmt.Sprintf("Disk description: %s", config.DiskDescription))

	_, err = diskAttachmentService.Update().DiskAttachment(
		ovirtsdk4.NewDiskAttachmentBuilder().
			Disk(diskBuilder.MustBuild()).
			MustBuild()).
		Send()
	if err != nil {
		err = fmt.Errorf("Failed to update disk properties: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	ui.Message(fmt.Sprintf("Waiting for disk '%s' reaching status OK...", diskID))
	stateChange := StateChangeConf{
		Pending:   []string{string(ovirtsdk4.DISKSTATUS_LOCKED)},
		Target:    []string{string(ovirtsdk4.DISKSTATUS_OK)},
		Refresh:   DiskStateRefreshFunc(conn, diskID),
		StepState: state,
	}
	_, err = WaitForState(&stateChange)
	if err != nil {
		err := fmt.Errorf("Failed waiting for disk attachment (%s) to become inactive: %s", diskID, err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepUpdateDisk) Cleanup(state multistep.StateBag) {}
