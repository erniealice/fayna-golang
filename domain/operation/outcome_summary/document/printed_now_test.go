package document

import (
	"context"
	"testing"
	"time"

	pyezatypes "github.com/erniealice/pyeza-golang/types"
)

func TestPrintedNowUsesRequestLocation(t *testing.T) {
	manila, err := time.LoadLocation("Asia/Manila")
	if err != nil {
		t.Fatalf("load zone: %v", err)
	}
	got := printedNow(pyezatypes.WithLocation(context.Background(), manila))
	if got.Location().String() != "Asia/Manila" {
		t.Fatalf("printedNow zone = %q, want Asia/Manila", got.Location())
	}
	if _, off := got.Zone(); off != 8*3600 {
		t.Fatalf("printedNow offset = %d, want +08:00", off)
	}
}
