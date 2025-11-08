package state

import "testing"

func TestCollectSnapshotEmpty(t *testing.T) {
	snap := CollectSnapshot("")
	if snap == nil {
		t.Fatal("snapshot should not be nil")
	}
	// We can't guarantee brew/mas presence; just ensure maps exist
	if snap.BrewFormulae == nil || snap.BrewCasks == nil || snap.MASApps == nil {
		t.Error("expected non-nil maps in snapshot")
	}
}
