package group

import _ "embed"

//go:embed sql/schema.sql
var GroupDDL string
