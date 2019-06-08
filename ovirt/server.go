package ovirt

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

// StateRefreshFunc is a function type used for StateChangeConf that is
// responsible for refreshing the item being watched for a state change.
//
// It returns three results:
// `result` is any object that will be returned as the final object after
// waiting for state change. This allows you to return the final updated
// object.
// `state` is the latest state of that object.
// `err` is any error that may have happened while refreshing the state.
type StateRefreshFunc func() (result interface{}, state string, err error)

// StateChangeConf is the configuration struct used for `WaitForState`.
type StateChangeConf struct {
	Pending   []string
	Refresh   StateRefreshFunc
	StepState multistep.StateBag
	Target    []string
}

// VMStateRefreshFunc returns a StateRefreshFunc that is used to watch
// a oVirt virtual machine.
func VMStateRefreshFunc(
	conn *ovirtsdk4.Connection, vmID string) StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.SystemService().
			VmsService().
			VmService(vmID).
			Get().
			Send()
		if err != nil {
			if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
				// Sometimes oVirt has consistency issues and doesn't see
				// newly created VM instance. Return empty state.
				return nil, "", nil
			}
			return nil, "", err
		}

		return resp.MustVm(), string(resp.MustVm().MustStatus()), nil
	}
}

// DiskAttachmentStateRefreshFunc returns a StateRefreshFunc that is used to
// watch a oVirt disk attachment.
func DiskAttachmentStateRefreshFunc(
	conn *ovirtsdk4.Connection, vmID string, diskID string) StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.SystemService().
			VmsService().
			VmService(vmID).
			DiskAttachmentsService().
			AttachmentService(diskID).
			Get().
			Send()
		if err != nil {
			if _, ok := err.(*ovirtsdk4.NotFoundError); ok {
				// Sometimes oVirt has consistency issues and doesn't see
				// newly created Disk instance. Return empty state.
				return nil, "", nil
			}
			return nil, "", nil
		}

		attachmentState := "inactive"
		if resp.MustAttachment().MustActive() {
			attachmentState = "active"
		}

		return resp.MustAttachment(), attachmentState, nil
	}
}

// WaitForState watches an object and waits for it to achieve a certain
// state.
func WaitForState(conf *StateChangeConf) (i interface{}, err error) {
	log.Printf("Waiting for state to become: %s", conf.Target)

	for {
		var currentState string
		i, currentState, err := conf.Refresh()
		if err != nil {
			return i, err
		}

		for _, t := range conf.Target {
			if currentState == t {
				return i, err
			}
		}

		if conf.StepState != nil {
			if _, ok := conf.StepState.GetOk(multistep.StateCancelled); ok {
				return nil, errors.New("interrupted")
			}
		}

		found := false
		for _, allowed := range conf.Pending {
			if currentState == allowed {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("unexpected state '%s', wanted target '%s'", currentState, conf.Target)
		}

		log.Printf("Waiting for state to become %s, currently %s", conf.Target, currentState)
		time.Sleep(2 * time.Second)
	}
}
