package ovirt

import (
	"context"
	"fmt"
	"log"
	//	"math"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

// stepCreateVM represents a Packer build step that creates ovirtsdk4 instances.
type stepCreateVM struct {
	Debug bool
	Ctx   interpolate.Context
}

func (s *stepCreateVM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	c := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	conn := state.Get("conn").(*ovirtsdk4.Connection)

	ui.Say("Creating VM...")

	clustersResponse, err := conn.SystemService().ClustersService().List().Send()
	if err != nil {
		err := fmt.Errorf("Error getting cluster list: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	var clusterID string
	if clusters, ok := clustersResponse.Clusters(); ok {
		for _, cluster := range clusters.Slice() {
			if clusterName, ok := cluster.Name(); ok {
				log.Printf("Found cluster name: %v\n", clusterName)
				if clusterName == c.Cluster {
					clusterID = cluster.MustId()
					log.Printf("Using cluster: %s", clusterID)
					break
				}
			}
		}
	}
	if clusterID == "" {
		err = fmt.Errorf("Could not find cluster '%s'", c.Cluster)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

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
		err = fmt.Errorf("Could not find template '%s' with version '%d'", c.SourceTemplate, c.SourceTemplateVersion)
		ui.Error(err.Error())
		state.Put("error", err)
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
		Name(c.VMName)

	cluster, err := ovirtsdk4.NewClusterBuilder().
		Id(clusterID).
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
		err := fmt.Errorf("Error creating VM: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	newVM, ok := vmAddResp.Vm()
	if !ok {
		state.Put("vm_id", "")
		return multistep.ActionHalt
	}

	vmID := newVM.MustId()
	ui.Message(fmt.Sprintf("Virtual machine '%s' has been defined", vmID))
	log.Printf("virtual machine id: %s", vmID)

	ui.Message(fmt.Sprintf("Waiting for VM to become ready (status down) ..."))
	stateChange := StateChangeConf{
		Pending:   []string{"image_locked"},
		Target:    []string{string(ovirtsdk4.VMSTATUS_DOWN)},
		Refresh:   VMStateRefreshFunc(conn, vmID),
		StepState: state,
	}
	latestVM, err := WaitForState(&stateChange)
	if err != nil {
		err := fmt.Errorf("Failed waiting for VM (%s) to become down: %s", vmID, err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("vm_id", latestVM.(*ovirtsdk4.Vm).MustId())

	return multistep.ActionContinue
}

// Cleanup any resources that may have been created during the Run phase.
func (s *stepCreateVM) Cleanup(state multistep.StateBag) {
	if _, ok := state.GetOk("vm_id"); !ok {
		return
	}

	ui := state.Get("ui").(packer.Ui)
	conn := state.Get("conn").(*ovirtsdk4.Connection)
	vmID := state.Get("vm_id").(string)

	ui.Say(fmt.Sprintf("Removing VM: %s", vmID))
	_, err := conn.SystemService().
		VmsService().
		VmService(vmID).
		Remove().
		Send()
	if err != nil {
		ui.Error(fmt.Sprintf("Error removing VM, may still be around: %s", err))
		return
	}

	stateChange := StateChangeConf{
		Pending:   []string{string(ovirtsdk4.VMSTATUS_UP), string(ovirtsdk4.VMSTATUS_DOWN)},
		Target:    []string{string(ovirtsdk4.VMSTATUS_DOWN)},
		Refresh:   VMStateRefreshFunc(conn, vmID),
		StepState: state,
	}
	WaitForState(&stateChange)
}
