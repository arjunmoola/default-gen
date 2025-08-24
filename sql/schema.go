package sql

import (
	_ "embed"
	"strings"
)

//go:embed schema.sql
var schema string

func Get() string {
	return schema
}

func GetSchemas() []string {
	return parseSchema(schema)
}

func parseSchema(s string) []string {
	var schemas []string

	idx := strings.Index(s, ";")

	if idx < 0 {
		return []string{s}
	}

	schemas = append(schemas, s[:idx+1])
	s = s[idx+2:]
	schemas = append(schemas, s)

	return schemas
}
