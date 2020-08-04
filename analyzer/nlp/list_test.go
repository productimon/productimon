package nlp

import "testing"

func TestProductimonLabels(t *testing.T) {
	if listLabel("productimon") != LabelProductimon {
		t.Fatal("production should be of label Productimon")
	}
	if listLabel("PrOdUcTimOn") != LabelProductimon {
		t.Fatal("production should be of label Productimon")
	}
	if listLabel("PrOdUcTimOn-reporter") != LabelProductimon {
		t.Fatal("production should be of label Productimon")
	}
}

func TestEducationDomainLabels(t *testing.T) {
	if listLabel("xvm.mit.edu") != LabelEducation {
		t.Fatal("Education domain test failed")
	}
	if listLabel("xVm.mIt.eDu") != LabelEducation {
		t.Fatal("Education domain test failed")
	}
	if listLabel("cse.unsw.edu.au") != LabelEducation {
		t.Fatal("Education domain test failed")
	}
	if listLabel("cSe.UNsW.EdU.Au") != LabelEducation {
		t.Fatal("Education domain test failed")
	}
}

func TestGovernmentDomainLabels(t *testing.T) {
	if listLabel("wh.gov") != LabelGovernment {
		t.Fatal("Government domain test failed")
	}
	if listLabel("my.gov.au") != LabelGovernment {
		t.Fatal("Government domain test failed")
	}
	if listLabel("wH.gOv") != LabelGovernment {
		t.Fatal("Government domain test failed")
	}
	if listLabel("mY.GoV.aU") != LabelGovernment {
		t.Fatal("Government domain test failed")
	}
}
