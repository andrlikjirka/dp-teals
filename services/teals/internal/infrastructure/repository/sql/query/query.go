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

	//go:embed scripts/ledger/GetMmrSize.sql
	GetMmrSize string
	//go:embed scripts/ledger/InsertMmrNode.sql
	InsertMmrNode string
	//go:embed scripts/ledger/GetRightmostPeakAtLevel.sql
	GetRightmostPeakAtLevel string
	//go:embed scripts/ledger/SetMmrNodeParent.sql
	SetMmrNodeParent string
	//go:embed scripts/ledger/GetMmrPeaks.sql
	GetMmrPeaks string
	//go:embed scripts/audit_log/GetAuditLogEntryByEventID.sql
	GetAuditLogEntryByEventID string
	//go:embed scripts/ledger/GetLeafToPeakPath.sql
	GetLeafToPeakPath string
	//go:embed scripts/ledger/GetNodesByIDs.sql
	GetNodesByIDs string
	//go:embed scripts/ledger/GetAncestorAtLevel.sql
	GetAncestorAtLevel string
	//go:embed scripts/ledger/GetPathToClosestNewPeak.sql
	GetPathToClosestNewPeak string
)
