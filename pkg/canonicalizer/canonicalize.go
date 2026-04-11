package canonicalizer

import (
	"encoding/json"

	"github.com/gowebpki/jcs"
)

// CanonicalizeAuditEvent serializes an audit event payload to canonical JSON (JCS / RFC 8785).
func CanonicalizeAuditEvent(p *AuditEventPayload) ([]byte, error) {
	return canonicalize(p)
}

// CanonicalizeCheckpoint serializes a checkpoint payload to canonical JSON (JCS / RFC 8785).
func CanonicalizeCheckpoint(p *CheckpointPayload) ([]byte, error) {
	return canonicalize(p)
}

// canonicalize serializes the payload to canonical JSON using JCS (RFC 8785).
// The resulting bytes are deterministic and suitable for signing and storage.
func canonicalize(payload any) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return jcs.Transform(data)
}
