package ovirt

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

type stepCreateVMFromTemplate struct {
	Debug bool
	Ctx   interpolate.Context
}

func (s *stepCreateVMFromTemplate) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)
	conn := state.Get("conn").(*ovirtsdk4.Connection)

	ui.Say("Creating virtual machine...")

	cResp, err := conn.SystemService().
		ClustersService().
		List().
		Send()
	if err != nil {
		err := fmt.Errorf("Error getting cluster list: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	var clusterID string
	if clusters, ok := cResp.Clusters(); ok {
		for _, cluster := range clusters.Slice() {
			if clusterName, ok := cluster.Name(); ok {
				if clusterName == config.Cluster {
					clusterID = cluster.MustId()
					log.Printf("Using cluster id: %s", clusterID)
					break
				}
			}
		}
	}
	if clusterID == "" {
		err = fmt.Errorf("Could not find cluster '%s'", config.Cluster)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	var templateID string
	if config.SourceTemplateID != "" {
		templateID = config.SourceTemplateID
	} else {
		templatesService := conn.SystemService().TemplatesService()
		log.Printf("Searching for template '%s'", config.SourceTemplateName)
		tpsResp, err := templatesService.List().
			Search(fmt.Sprintf("name=%s", config.SourceTemplateName)).
			Send()
		if err != nil {
			err = fmt.Errorf("Error searching templates: %s", err)
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}
		tpSlice, _ := tpsResp.Templates()

		for _, tp := range tpSlice.Slice() {
			if tp.MustVersion().MustVersionNumber() == int64(config.SourceTemplateVersion) {
				templateID = tp.MustId()
				break
			}
		}
		if templateID == "" {
			err = fmt.Errorf("Could not find template '%s' with version '%d'", config.SourceTemplateName, config.SourceTemplateVersion)
			ui.Error(err.Error())
			state.Put("error", err)
			return multistep.ActionHalt
		}
	}
	log.Printf("Using template id: %s", templateID)

	vmBuilder := ovirtsdk4.NewVmBuilder().
		Name(config.VMName)

	cluster, err := ovirtsdk4.NewClusterBuilder().
		Id(clusterID).
		Build()
	if err != nil {
		err = fmt.Errorf("Error creating cluster object: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	vmBuilder.Cluster(cluster)

	t, err := ovirtsdk4.NewTemplateBuilder().
		Id(templateID).
		Build()
	if err != nil {
		err = fmt.Errorf("Error creating template object: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}
	vmBuilder.Template(t)

	vm, err := vmBuilder.Build()
	if err != nil {
		err = fmt.Errorf("Error creating VM object: %s", err)
		ui.Error(err.Error())
		state.Put("error", err)
		return multistep.ActionHalt
	}

	vmAddResp, err := conn.SystemService().
		VmsService().
		Add().
		Vm(vm).
		Send()
	if err != nil {
		err := fmt.Errorf("Error creating virtual machine: %s", err)
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
	log.Printf("Virtual machine id: %s", vmID)

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

func (s *stepCreateVMFromTemplate) Cleanup(state multistep.StateBag) {
	if _, ok := state.GetOk("vm_id"); !ok {
		return
	}

	ui := state.Get("ui").(packer.Ui)
	conn := state.Get("conn").(*ovirtsdk4.Connection)
	vmID := state.Get("vm_id").(string)

	ui.Say(fmt.Sprintf("Deleting virtual machine: %s ...", vmID))

	if _, err := conn.SystemService().VmsService().VmService(vmID).Remove().Send(); err != nil {
		ui.Error(fmt.Sprintf("Error deleting VM '%s', may still be around: %s", vmID, err))
	}
}
