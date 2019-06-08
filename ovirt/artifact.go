package ovirt

import (
	"fmt"
	"log"
)

// Artifact is an artifact implementation that contains built disk.
type Artifact struct {
	diskId string
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (*Artifact) Files() []string {
	return nil
}

func (a *Artifact) Id() string {
	return a.diskId
}

func (a *Artifact) String() string {
	return fmt.Sprintf("A disk was created: %s", a.diskId)
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	log.Printf("Destroying disk: %d", a.diskId)
	//TODO
	return nil
}
