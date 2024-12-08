package assets

import _ "embed"

//go:embed sql/schema-lookup.sql
var LocalDDL string

//go:embed sql/schema-group.sql
var RemoteDDL string
