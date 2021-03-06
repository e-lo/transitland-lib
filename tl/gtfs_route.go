package tl

import (
	"fmt"
	"strconv"

	"github.com/interline-io/transitland-lib/tl/causes"
	"github.com/interline-io/transitland-lib/tl/enum"
)

// Route routes.txt
type Route struct {
	RouteID        string `csv:"route_id" required:"true"`
	AgencyID       string `csv:"agency_id"`
	RouteShortName string `csv:"route_short_name"`
	RouteLongName  string `csv:"route_long_name"`
	RouteDesc      string `csv:"route_desc"`
	RouteType      int    `csv:"route_type" required:"true"`
	RouteURL       string `csv:"route_url"`
	RouteColor     string `csv:"route_color"`
	RouteTextColor string `csv:"route_text_color"`
	RouteSortOrder int    `csv:"route_sort_order"`
	BaseEntity
}

// EntityID returns ID or RouteID.
func (ent *Route) EntityID() string {
	return entID(ent.ID, ent.RouteID)
}

// EntityKey returns the GTFS identifier.
func (ent *Route) EntityKey() string {
	return ent.RouteID
}

// Warnings for this Entity.
func (ent *Route) Warnings() (errs []error) {
	errs = append(errs, ent.loadWarnings...)
	if len(ent.AgencyID) == 0 {
		errs = append(errs, causes.NewValidationWarning("agency_id", "optional agency_id value recommended"))
	}
	if len(ent.RouteDesc) > 0 && (ent.RouteDesc == ent.RouteLongName || ent.RouteDesc == ent.RouteShortName) {
		errs = append(errs, causes.NewValidationWarning("route_desc", "route_desc should not duplicate route_short_name or route_long_name"))
	}
	if _, ok := enum.GetRouteType(ent.RouteType); ok && ent.RouteType > 7 {
		errs = append(errs, causes.NewValidationWarning("route_type", "extended route_types not universally supported"))
	}
	if len(ent.RouteColor) > 0 && ent.RouteColor == ent.RouteTextColor {
		errs = append(errs, causes.NewValidationWarning("route_text_color", "route_text_color should provide contrast with route_color"))
	}
	return errs
}

// Errors for this Entity.
func (ent *Route) Errors() (errs []error) {
	errs = append(errs, ent.BaseEntity.Errors()...)
	errs = append(errs, enum.CheckPresent("route_id", ent.RouteID)...)
	errs = append(errs, enum.CheckURL("route_url", ent.RouteURL)...)
	errs = append(errs, enum.CheckColor("route_color", ent.RouteColor)...)
	errs = append(errs, enum.CheckColor("route_text_color", ent.RouteTextColor)...)
	errs = append(errs, enum.CheckPositiveInt("route_sort_order", ent.RouteSortOrder)...)
	if len(ent.RouteShortName) == 0 && len(ent.RouteLongName) == 0 {
		errs = append(errs, causes.NewConditionallyRequiredFieldError("route_short_name"))
	}
	if _, ok := enum.GetRouteType(ent.RouteType); !ok {
		errs = append(errs, causes.NewInvalidFieldError("route_type", strconv.Itoa(ent.RouteType), fmt.Errorf("invalid route_type %d", ent.RouteType)))
	}
	return errs
}

// Filename routes.txt
func (ent *Route) Filename() string {
	return "routes.txt"
}

// TableName gtfs_routes
func (ent *Route) TableName() string {
	return "gtfs_routes"
}

// UpdateKeys updates Entity references.
func (ent *Route) UpdateKeys(emap *EntityMap) error {
	if agencyID, ok := emap.GetEntity(&Agency{AgencyID: ent.AgencyID}); ok {
		ent.AgencyID = agencyID
	} else {
		return causes.NewInvalidReferenceError("agency_id", ent.AgencyID)
	}
	return nil
}
