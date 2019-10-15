package dmfr

import (
	"errors"
	"time"

	"github.com/interline-io/gotransit"
	"github.com/interline-io/gotransit/gtcsv"
	"github.com/interline-io/gotransit/gtdb"
)

// MainFetchFeed .
// Fetch errors are logged to Feed LastFetchError and saved.
// An error return from this function is a serious failure.
func MainFetchFeed(tx gtdb.Adapter, feedid int) (int, error) {
	// Get the parent feed
	tlfeed := Feed{}
	tlfeed.ID = feedid
	fvid := 0
	if err := tx.Find(&tlfeed); err != nil {
		return fvid, err
	}
	fetchtime := time.Now().UTC()
	tlfeed.LastFetchedAt = fetchtime
	tlfeed.LastFetchError = ""
	// Immediately save LastFetchedAt to prevent possible re-enqueing
	if err := tx.Update(&tlfeed, "last_fetched_at", "last_fetch_error"); err != nil {
		return fvid, err
	}
	// Start fetching
	if fvid2, err := FetchAndCreateFeedVersion(tx, feedid, tlfeed.URL, fetchtime); err != nil {
		tlfeed.LastFetchError = err.Error()
	} else {
		tlfeed.LastFetchError = ""
		tlfeed.LastSuccessfulFetchAt = fetchtime
		fvid = fvid2
	}
	// Save updated timestamps
	if err := tx.Update(&tlfeed, "last_fetched_at", "last_fetch_error", "last_successful_fetch_at"); err != nil {
		return fvid, err
	}
	return fvid, nil
}

// FetchAndCreateFeedVersion from a URL.
// Returns an error if the source cannot be loaded or is invalid GTFS.
// Returns no error if the SHA1 is already present, or a FeedVersion is created.
func FetchAndCreateFeedVersion(tx gtdb.Adapter, feedid int, url string, fetchtime time.Time) (int, error) {
	fvid := 0
	fv, err := NewFeedVersionFromURL(url)
	if err != nil {
		return fvid, err
	}
	fv.FeedID = feedid
	fv.FetchedAt = fetchtime
	// Is this SHA1 already present?
	checkfvid := gotransit.FeedVersion{}
	if err := tx.Where("sha1 = ?", fv.SHA1).FirstOrInit(&checkfvid).Error; err != nil {
		return fvid, err
	} else if checkfvid.ID != 0 {
		// fmt.Printf("feed_version with SHA1 '%s' already exists: %d", fv.SHA1, checkfvid.ID)
		return checkfvid.ID, nil
	}
	// Create FeedVersion
	if err := tx.Create(&fv).Error; err != nil {
		return fvid, err
	}
	return fv.ID, nil
}

// NewFeedVersionFromURL returns a new FeedVersion initialized from the given URL.
func NewFeedVersionFromURL(url string) (gotransit.FeedVersion, error) {
	// Init FV
	fv := gotransit.FeedVersion{}
	fv.URL = url
	fv.FeedType = "gtfs"
	// Download feed
	reader, err := gtcsv.NewReader(url)
	if err != nil {
		return fv, err
	}
	if err := reader.Open(); err != nil {
		return fv, err
	}
	defer reader.Close()
	// fv.FetchedAt = &fetchtime
	// Are we a zip archive? Can we read the SHA1?
	if h, err := getSHA1(reader); err == nil {
		fv.SHA1 = h
	} else {
		return fv, err
	}
	// Perform basic GTFS validity checks
	if errs := reader.ValidateStructure(); len(errs) > 0 {
		return fv, errs[0]
	}
	// Get service dates
	start, end, err := servicePeriod(reader)
	if err != nil {
		return fv, err
	}
	fv.EarliestCalendarDate = start
	fv.LatestCalendarDate = end
	return fv, nil
}

type canSHA1 interface {
	SHA1() (string, error)
}

func getSHA1(reader gotransit.Reader) (string, error) {
	ad, ok := reader.(canSHA1)
	if !ok {
		return "", errors.New("not a zip source")
	}
	h, err := ad.SHA1()
	if err != nil {
		return "", err
	}
	return h, nil
}

func servicePeriod(reader gotransit.Reader) (time.Time, time.Time, error) {
	var start time.Time
	var end time.Time
	for c := range reader.Calendars() {
		if start.IsZero() || c.StartDate.Before(start) {
			start = c.StartDate
		}
		if end.IsZero() || c.EndDate.After(end) {
			end = c.EndDate
		}
	}
	for cd := range reader.CalendarDates() {
		if cd.ExceptionType != 1 {
			continue
		}
		if start.IsZero() || cd.Date.Before(start) {
			start = cd.Date
		}
		if end.IsZero() || cd.Date.After(end) {
			end = cd.Date
		}
	}
	if start.IsZero() || end.IsZero() {
		return start, end, errors.New("start or end dates were empty")
	}
	if end.Before(start) {
		return start, end, errors.New("end before start")
	}
	return start, end, nil
}
