package fayna

import (
	"net/http"
	"testing"
)

func TestHTMXSuccess(t *testing.T) {
	t.Parallel()

	result := HTMXSuccess("inventory-table")

	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}

	trigger, ok := result.Headers["HX-Trigger"]
	if !ok {
		t.Fatal("missing HX-Trigger header")
	}

	wantTrigger := `{"formSuccess":true,"refreshTable":"inventory-table"}`
	if trigger != wantTrigger {
		t.Errorf("HX-Trigger = %q, want %q", trigger, wantTrigger)
	}
}

func TestHTMXError(t *testing.T) {
	t.Parallel()

	result := HTMXError("Something went wrong")

	if result.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusUnprocessableEntity)
	}

	msg, ok := result.Headers["HX-Error-Message"]
	if !ok {
		t.Fatal("missing HX-Error-Message header")
	}
	if msg != "Something went wrong" {
		t.Errorf("HX-Error-Message = %q, want %q", msg, "Something went wrong")
	}
}
