package schema

import (
	"embed"
)

//go:embed *.sql
var Migrations embed.FS

//go:embed reporting_tables.json
var ReportingTables []byte
