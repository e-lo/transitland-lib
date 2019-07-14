package gtdb

import (
	"strings"

	// Log
	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/gotransit"
	"github.com/jmoiron/sqlx"
	// Postgres
	// Drivers
)

// NewAdapter returns a Adapter for the given dburl.
func NewAdapter(dburl string) Adapter {
	if strings.HasPrefix(dburl, "postgres://") {
		return &PostgresAdapter{DBURL: dburl}
	} else if strings.HasPrefix(dburl, "sqlite3://") {
		return &SpatiaLiteAdapter{DBURL: dburl}
	}
	return nil
}

// Adapter implements details specific to each backend.
type Adapter interface {
	Open() error
	Close() error
	Create() error
	DB() *sqlx.DB
	Sqrl() sq.StatementBuilderType
	Insert(interface{}) (int, error)
	Find(interface{}) error
	Get(interface{}, string, ...interface{}) error
	Select(interface{}, string, ...interface{}) error
	BatchInsert([]gotransit.Entity) error
}
