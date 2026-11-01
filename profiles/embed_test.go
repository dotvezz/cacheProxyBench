package profiles

import (
	"testing"
)

func TestGetIncluded(t *testing.T) {
	included := GetIncluded()

	if len(included) < 1 {
		t.Errorf("GetIncluded() = %v, want at least 1", len(included))
	}

	if included[0].Name != "Simple Test - Just Two Requests" {
		t.Errorf("GetIncluded()[0].Name = %v, want Simple Test - Just Two Requests", included[0].Name)
	}
}
