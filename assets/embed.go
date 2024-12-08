package assets

import _ "embed"

//go:embed sql/snitch-local.sql
var LocalDDL string

//go:embed sql/snitch-server.sql
var RemoteDDL string
