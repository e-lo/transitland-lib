package copier

import (
	"testing"

	"github.com/interline-io/transitland-lib/tlcsv"
	"github.com/interline-io/transitland-lib/internal/testutil"
	"github.com/interline-io/transitland-lib/tl"
)

func Test_geomCache(t *testing.T) {
	r, err := tlcsv.NewReader(testutil.ExampleDir.URL)
	if err != nil {
		t.Error(err)
	}
	r.Open()
	defer r.Close()
	trips := map[string]tl.Trip{}
	count := 1
	for trip := range r.Trips() {
		trip.StopPatternID = count
		trips[trip.TripID] = trip
		count++
	}
	cache := newGeomCache()
	for e := range r.Shapes() {
		cache.AddShape(e.ShapeID, e)
	}
	for e := range r.Stops() {
		cache.AddStop(e.StopID, e)
	}
	for stoptimes := range r.StopTimesByTripID() {
		trip := trips[stoptimes[0].TripID]
		stoptimes2, err := cache.InterpolateStopTimes(trip, stoptimes)
		if err != nil {
			t.Error(err)
		}
		if len(stoptimes) != len(stoptimes2) {
			t.Error("unequal length")
		}
	}
	// check that we had cache hits
	if x := len(cache.positions); x < 9 {
		t.Errorf("expected at least %d cached trip journeys, got %d", 9, x)
	}
}
