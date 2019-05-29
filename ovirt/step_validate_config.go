package ovirt

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

type stepValidateConfig struct{}

func (s *stepValidateConfig) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	conn := state.Get("conn").(*ovirtsdk4.Connection)
	config := state.Get("config").(*Config)
	ui := state.Get("ui").(packer.Ui)

	var err error
	//	var errs *packer.MultiError

	ui.Say("Query oVirt clusters...")
	clustersService := conn.SystemService().ClustersService()

	// Use the "list" method of the "clusters" service to list all the clusters of the system
	clustersResponse, err := clustersService.List().Send()
	if err != nil {
		ui.Error(fmt.Sprintf("oVirt: Failed to get cluster list, reason: %v", err))
		return multistep.ActionHalt
	}

	if clusters, ok := clustersResponse.Clusters(); ok {
		for _, cluster := range clusters.Slice() {
			if clusterName, ok := cluster.Name(); ok {
				log.Printf("Found cluster name: %v\n", clusterName)
				if clusterName == config.Cluster {
					cId := cluster.MustId()
					log.Printf("Using cluster: %s", cId)
					state.Put("cluster_id", cId)
					break
				}
			}
		}
	}

	ui.Message("Configuration validated!")
	return multistep.ActionContinue
}

// Cleanup any resources that may have been created during the Run phase.
func (s *stepValidateConfig) Cleanup(state multistep.StateBag) {
	// Nothing to cleanup for this step.
}
