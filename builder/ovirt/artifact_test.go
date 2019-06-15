package ovirt

import (
    "testing"

    "github.com/hashicorp/packer/packer"
)

func TestArtifact_Impl(t *testing.T) {
    var _ packer.Artifact = new(Artifact)
}

func TestArtifactId(t *testing.T) {
    expected := `c2867299-28ea-48a2-922a-805b999fcb2d`

    a := &Artifact{
        diskID: "c2867299-28ea-48a2-922a-805b999fcb2d",
    }

    result := a.Id()
    if result != expected {
        t.Fatalf("wrong artifact id returned: %s", result)
    }
}

func TestArtifactString(t *testing.T) {
    expected := "A disk was created: c2867299-28ea-48a2-922a-805b999fcb2d"

    a := &Artifact{
        diskID: "c2867299-28ea-48a2-922a-805b999fcb2d",
    }
    result := a.String()
    if result != expected {
        t.Fatalf("bad message returned: %s", result)
    }
}
