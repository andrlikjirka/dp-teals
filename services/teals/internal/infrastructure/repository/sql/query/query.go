package query

import (
	_ "embed"
)

var (
	//go:embed scripts/audit_log/Insert.sql
	InsertAuditEvent string

	//go:embed scripts/producer/AddPublicKey.sql
	AddProducerPublicKey string
	//go:embed scripts/producer/SelectPublicKey.sql
	SelectProducerPublicKey string
	//go:embed scripts/producer/RevokePublicKey.sql
	RevokeProducerPublicKey string
)
