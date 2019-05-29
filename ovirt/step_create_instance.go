package ovirt

import (
	"context"
	"fmt"
	"log"
	"math"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

// stepCreateInstance represents a Packer build step that creates ovirtsdk4 instances.
type stepCreateInstance struct {
	Debug bool
	Ctx   interpolate.Context
}

// Run executes the Packer build step that creates a ovirtsdk4 instance.
func (s *stepCreateInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	conn := state.Get("conn").(*ovirtsdk4.Connection)

	ui.Say("Creating instance...")

	// Get the reference to the service that manages the storage domains
	sdsService := conn.SystemService().StorageDomainsService()

	// Find the storage domain we want to be used for virtual machine disks
	log.Printf("Searching for storage domain '%s'\n", c.StorageDomain)
	sdsResp, err := sdsService.List().Search(fmt.Sprintf("name=%s", c.StorageDomain)).Send()
	if err != nil {
		ui.Error(fmt.Sprintf("ovirtsdk4: Error searching for storage domains\n%v", err))
		return multistep.ActionHalt
	}
	sdSlice, ok := sdsResp.StorageDomains()
	if !ok {
		ui.Error(fmt.Sprintf("ovirtsdk4: Error getting storage domain"))
		return multistep.ActionHalt
	}
	sd := sdSlice.Slice()[0]
	sdID := sd.MustId()
	log.Printf("Using storage domain: %s", sdID)

	log.Printf("Searching for template '%s'\n", c.SourceTemplate)
	templatesService := conn.SystemService().TemplatesService()
	tpsResp, err := templatesService.List().Search(fmt.Sprintf("name=%s", c.SourceTemplate)).Send()
	if err != nil {
		ui.Error(fmt.Sprintf("ovirtsdk4: Error searching for templates\n%v", err))
		return multistep.ActionHalt
	}
	tpSlice, _ := tpsResp.Templates()

	var templateID string
	for _, tp := range tpSlice.Slice() {
		if tp.MustVersion().MustVersionNumber() == int64(c.SourceTemplateVersion) {
			templateID = tp.MustId()
			break
		}
	}
	if templateID == "" {
		ui.Error(fmt.Sprintf("Could not find template '%s' with version '%d'", c.SourceTemplate, c.SourceTemplateVersion))
		return multistep.ActionHalt
	}
	log.Printf("Using template: %s", templateID)

	// Find the template disk we want be created on specific storage domain
	// for our virtual machine
	// tpService := templatesService.TemplateService(templateID)
	// tpGetResp, _ := tpService.Get().Send()
	// tp, _ := tpGetResp.Template()
	// tpDisks, _ := conn.FollowLink(tp.MustDiskAttachments())
	// disks, ok := tpDisks.(*ovirtsdk4.DiskAttachmentSlice)
	// if !ok {
	// 	ui.Error(fmt.Sprintf("ovirtsdk4: Failed to get template disk attachment, reason: %v", err))
	// 	return multistep.ActionHalt
	// }
	// disk := disks.Slice()[0].MustDisk()

	vmBuilder := ovirtsdk4.NewVmBuilder().
		Name(c.VMName).
		// Memory is specified in MB
		Memory(int64(c.Memory) * int64(math.Pow(2, 20)))

	cluster, err := ovirtsdk4.NewClusterBuilder().
		Id(state.Get("cluster_id").(string)).
		Build()
	if err != nil {
		ui.Error(fmt.Sprintf("ovirtsdk4: Error creating cluster reference\n%v", err))
		return multistep.ActionHalt
	}
	vmBuilder.Cluster(cluster)

	t, err := ovirtsdk4.NewTemplateBuilder().
		Id(templateID).
		Build()
	if err != nil {
		ui.Error(fmt.Sprintf("ovirtsdk4: Error creating template reference\n%v", err))
		return multistep.ActionHalt
	}
	vmBuilder.Template(t)

	cpuTopo := ovirtsdk4.NewCpuTopologyBuilder().
		Sockets(int64(c.Sockets)).
		Cores(int64(c.Cores)).
		Threads(1).
		MustBuild()
	cpu, err := ovirtsdk4.NewCpuBuilder().
		Topology(cpuTopo).
		Build()
	if err != nil {
		ui.Error(fmt.Sprintf("ovirtsdk4: Error creating cpu reference\n%v", err))
		return multistep.ActionHalt
	}
	vmBuilder.Cpu(cpu)

	sdBuilder, err := ovirtsdk4.NewStorageDomainBuilder().
		Id(sdID).
		Build()
	if err != nil {
		ui.Error(fmt.Sprintf("ovirtsdk4: Error creating storage domain reference\n%v", err))
		return multistep.ActionHalt
	}
	disk, err := ovirtsdk4.NewDiskBuilder().
		Format(ovirtsdk4.DISKFORMAT_COW).
		StorageDomainsOfAny(sdBuilder).
		Build()
	if err != nil {
		ui.Error(fmt.Sprintf("ovirtsdk4: Error creating disk reference\n%v", err))
		return multistep.ActionHalt
	}
	daBuilder, err := ovirtsdk4.NewDiskAttachmentBuilder().
		Disk(disk).
		Build()
	if err != nil {
		ui.Error(fmt.Sprintf("ovirtsdk4: Error creating disk attachment reference\n%v", err))
		return multistep.ActionHalt
	}
	vmBuilder.DiskAttachmentsOfAny(daBuilder)

	vm, err := vmBuilder.Build()
	if err != nil {
		ui.Error(fmt.Sprintf("ovirtsdk4: Error creating VM reference\n%v", err))
		return multistep.ActionHalt
	}

	vmAddResp, err := conn.SystemService().
		VmsService().
		Add().
		Vm(vm).
		Send()
	if err != nil {
		ui.Error(fmt.Sprintf("ovirtsdk4: Error creating VM\n%v", err))
		return multistep.ActionHalt
	}

	newVM, ok := vmAddResp.Vm()
	if !ok {
		state.Put("vm_id", "")
		return multistep.ActionHalt
	}

	state.Put("vm_id", newVM.MustId())

	ui.Message(fmt.Sprintf("Instance '%s' has been defined. Waiting for status ready...", state.Get("vm_id")))

	return multistep.ActionContinue
}

// Cleanup any resources that may have been created during the Run phase.
func (s *stepCreateInstance) Cleanup(state multistep.StateBag) {
	// Nothing to cleanup for this step.
}
