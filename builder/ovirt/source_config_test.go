package ovirt

import (
	"testing"
)

func TestSourceConfig_Prepare_name(t *testing.T) {
	sc := testSourceConfig()
	errs := sc.Prepare(nil)
	if errs != nil {
		t.Fatal("should not fail to initialize minimal source config")
	}

	sc = testSourceConfig()
	sc.SourceTemplateName = ""
	errs = sc.Prepare(nil)
	if errs == nil {
		t.Fatal("should not accept empty name")
	}
}

func TestSourceConfig_Prepare_type(t *testing.T) {
	sc := testSourceConfig()
	sc.SourceType = "foo"
	errs := sc.Prepare(nil)
	if errs == nil {
		t.Fatal("should not accept invalid type")
	}
}

func TestSourceConfig_Prepare_template(t *testing.T) {
	sc := testTemplateSourceConfig()
	errs := sc.Prepare(nil)
	if errs != nil {
		t.Fatal("should not fail to initialize template source config")
	}

	sc = testTemplateSourceConfig()
	sc.SourceTemplateVersion = 42
	errs = sc.Prepare(nil)
	if errs != nil {
		t.Fatal("should not fail to accept template version")
	}

	sc = testTemplateSourceConfig()
	sc.SourceTemplateName = ""
	errs = sc.Prepare(nil)
	if errs == nil {
		t.Fatal("should not accept empty template name")
	}

	sc = testTemplateSourceConfig()
	sc.SourceTemplateName = ""
	sc.SourceTemplateID = "foo"
	errs = sc.Prepare(nil)
	if errs == nil {
		t.Fatal("should not accept invalid template uuid")
	}

	sc = testTemplateSourceConfig()
	sc.SourceTemplateName = ""
	sc.SourceTemplateID = "c2867299-28ea-48a2-922a-805b999fcb2d"
	errs = sc.Prepare(nil)
	if errs != nil {
		t.Fatal("should not fail when template id is given")
	}

	sc = testTemplateSourceConfig()
	sc.SourceTemplateID = "c2867299-28ea-48a2-922a-805b999fcb2d"
	errs = sc.Prepare(nil)
	if errs == nil {
		t.Fatal("should fail when both template name and id are given")
	}
}

func testSourceConfig() SourceConfig {
	return SourceConfig {
		SourceTemplateName: "foo",
	}
}

func testTemplateSourceConfig() SourceConfig {
	return SourceConfig {
		SourceType: "template",
		SourceTemplateName: "foo",
	}
}
