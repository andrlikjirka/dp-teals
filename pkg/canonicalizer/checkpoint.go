package canonicalizer

// CheckpointPayload is the canonical representation of a ledger checkpoint used as the JWS signing payload.
type CheckpointPayload struct {
	RootHash   string `json:"root_hash"`
	Size       int64  `json:"size"`
	AnchoredAt string `json:"anchored_at"`
}
