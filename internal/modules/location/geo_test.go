package location

import "testing"

func TestDistanceKm(t *testing.T) {
	// Approximate distance Delhi to Mumbai (~1150 km).
	got := DistanceKm(28.6139, 77.2090, 19.0760, 72.8777)
	if got < 1100 || got > 1200 {
		t.Fatalf("DistanceKm() = %f, want roughly 1150 km", got)
	}
}
