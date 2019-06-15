package ovirt

import (
	"fmt"
	"log"
)

// Artifact is an artifact implementation that contains built disk.
type Artifact struct {
	diskID string
}

// BuilderId uniquely identifies the builder.
func (*Artifact) BuilderId() string {
	return BuilderId
}

// Files returns the files represented by the artifact. Not used for oVirt.
func (*Artifact) Files() []string {
	return nil
}

// Id returns the disk identifier of the artifact.
func (a *Artifact) Id() string {
	return a.diskID
}

func (a *Artifact) String() string {
	return fmt.Sprintf("A disk was created: %s", a.diskID)
}

// State returns specific details from the artifact. Not used for oVirt.
func (a *Artifact) State(name string) interface{} {
	return nil
}

// Destroy deletes the custom image associated with the artifact.
func (a *Artifact) Destroy() error {
	log.Printf("Destroying disk: %s", a.diskID)
	//TODO
	return nil
}
