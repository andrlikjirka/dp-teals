package query

import (
	_ "embed"
)

var (
	//go:embed scripts/audit_log/Insert.sql
	InsertAuditEvent string
)
