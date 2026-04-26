package canonical

import (
	"encoding/json"

	"github.com/gowebpki/jcs"
)

// CanonicalizeAuditEvent serializes an audit event payload to canonical JSON (JCS / RFC 8785).
func CanonicalizeAuditEvent(p *AuditEventPayload) (json.RawMessage, error) {
	return canonicalize(p)
}

// CanonicalizeCheckpoint serializes a checkpoint payload to canonical JSON (JCS / RFC 8785).
func CanonicalizeCheckpoint(p *CheckpointPayload) (json.RawMessage, error) {
	return canonicalize(p)
}

// CanonicalizeProtectedAuditEvent serializes a protected audit event payload to canonical JSON (JCS / RFC 8785).
func CanonicalizeProtectedAuditEvent(p *ProtectedAuditEventPayload) (json.RawMessage, error) {
	return canonicalize(p)
}

// CanonicalizeMetadata serializes metadata to canonical JSON (JCS / RFC 8785).
func CanonicalizeMetadata(metadata map[string]any) (json.RawMessage, error) {
	return canonicalize(metadata)
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
