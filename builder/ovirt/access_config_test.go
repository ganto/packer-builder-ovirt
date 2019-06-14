package ovirt

import (
	"testing"
)

func TestAccessConfig_Prepare(t *testing.T) {
	ac := testAccessConfig()
	errs := ac.Prepare(nil)
	if errs != nil {
		t.Fatal("should not fail to initialize minimal access config")
	}

	ac = testAccessConfig()
	ac.SkipCertValidation = true
	errs = ac.Prepare(nil)
	if errs != nil {
		t.Fatal("should allow to disable cert validation")
	}

	ac = testAccessConfig()
	ac.OvirtURLRaw = ""
	errs = ac.Prepare(nil)
	if errs == nil {
		t.Fatal("should not accept empty url")
	}

	ac = testAccessConfig()
	ac.OvirtURLRaw = ":foobar"
	errs = ac.Prepare(nil)
	if errs == nil {
		t.Fatal("should not accept invalid url")
	}

	ac = testAccessConfig()
	ac.Username = ""
	errs = ac.Prepare(nil)
	if errs == nil {
		t.Fatal("should not accept empty username")
	}

	ac = testAccessConfig()
	ac.Password = ""
	errs = ac.Prepare(nil)
	if errs == nil {
		t.Fatal("should not accept empty password")
	}
}

func testAccessConfig() AccessConfig {
	return AccessConfig {
		OvirtURLRaw: "https://ovirt.example.com/ovirt/api",
		Username:    "admin@internal",
		Password:    "password",
	}
}
