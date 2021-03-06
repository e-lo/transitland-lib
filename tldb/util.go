package tldb

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx/reflectx"
	"github.com/rakyll/statik/fs"

	// Static assets
	_ "github.com/interline-io/transitland-lib/internal/schema"
)

var mapper = reflectx.NewMapperFunc("db", toSnakeCase)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

type hasTableName interface {
	TableName() string
}

type canSetID interface {
	SetID(int)
}

type canGetID interface {
	EntityID() string
}

type canUpdateTimestamps interface {
	UpdateTimestamps()
}

type canSetFeedVersion interface {
	SetFeedVersionID(int)
}

func getSchema(filename string) (string, error) {
	statikFS, err := fs.New()
	if err != nil {
		return "", err
	}
	f, err := statikFS.Open(filename)
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadAll(f)
	return string(data), err
}

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func getFieldNameIndexes(ent interface{}) ([]string, []string) {
	names := []string{}
	wraps := []string{}
	fields := mapper.TypeMap(reflect.TypeOf(ent))
	for _, fi := range fields.Index {
		if fi.Embedded == true || fi.Name == "id" || strings.Contains(fi.Path, ".") {
			continue
		}
		w := ""
		if wrap, ok := fi.Options["insert"]; ok {
			w = strings.Replace(wrap, "@", ",", -1)
		}
		names = append(names, fi.Path)
		wraps = append(wraps, w)
	}
	return names, wraps
}

// getInsert returns column names and a slice of placeholders or squirrel expressions.
func getInsert(ent interface{}) ([]string, []interface{}, error) {
	vals := make([]interface{}, 0)
	val := reflect.ValueOf(ent).Elem()
	fm := mapper.FieldMap(val)
	names, _ := getFieldNameIndexes(ent)
	for _, name := range names {
		v, ok := fm[name]
		if !ok {
			// This should not happen.
			return names[0:0], vals[0:0], fmt.Errorf("unknown field: %s", name)
		}
		vals = append(vals, v.Interface())
	}
	if len(names) == 0 || len(names) != len(vals) {
		return names[0:0], vals[0:0], errors.New("no columns or values")
	}
	return names, vals, nil
}

func getTableName(ent interface{}) string {
	if v, ok := ent.(hasTableName); ok {
		return v.TableName()
	}
	s := strings.Split(fmt.Sprintf("%T", ent), ".")
	return toSnakeCase(s[len(s)-1])
}

func getID(ent interface{}) (int, error) {
	if v, ok := ent.(canGetID); ok {
		return strconv.Atoi(v.EntityID())
	}
	return 0, errors.New("no ID")
}

func contains(a string, b []string) bool {
	for _, v := range b {
		if a == v {
			return true
		}
	}
	return false
}
