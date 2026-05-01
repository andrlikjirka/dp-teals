package canonical

import (
	"bytes"
	"testing"
)

func TestCanonicalizeCheckpoint(t *testing.T) {
	tests := []struct {
		name     string
		input    *CheckpointPayload
		expected string
	}{
		{
			name: "golden value",
			input: &CheckpointPayload{
				RootHash:   "abc123",
				Size:       42,
				AnchoredAt: "2024-01-01T00:00:00Z",
			},
			// JCS sorts keys lexicographically: anchored_at, root_hash, size
			expected: `{"anchored_at":"2024-01-01T00:00:00Z","root_hash":"abc123","size":42}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CanonicalizeCheckpoint(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.expected {
				t.Errorf("got  %s\nwant %s", got, tt.expected)
			}
		})
	}
}

func TestCanonicalizeAuditEvent(t *testing.T) {
	tests := []struct {
		name     string
		input    *AuditEventPayload
		expected string
	}{
		{
			name: "all required fields, no environment or metadata",
			input: &AuditEventPayload{
				ID:        "evt-1",
				Timestamp: "2024-01-01T00:00:00Z",
				Actor:     ActorPayload{Type: "user", ID: "user-1"},
				Subject:   SubjectPayload{ID: "subject-1"},
				Action:    "create",
				Resource:  ResourcePayload{ID: "res-1", Name: "test-resource"},
				Result:    ResultPayload{Status: "success"},
			},
			// Top-level keys sorted: action, actor, id, resource, result, subject, timestamp
			// actor sorted: id, type
			// resource sorted: id, name  (fields omitted — empty slice)
			// result: only status (reason omitted)
			expected: `{"action":"create","actor":{"id":"user-1","type":"user"},"id":"evt-1","resource":{"id":"res-1","name":"test-resource"},"result":{"status":"success"},"subject":{"id":"subject-1"},"timestamp":"2024-01-01T00:00:00Z"}`,
		},
		{
			name: "with environment",
			input: &AuditEventPayload{
				ID:          "evt-2",
				Timestamp:   "2024-01-01T00:00:00Z",
				Environment: &EnvironmentPayload{Service: "test-svc", TraceID: "trace-1", SpanID: "span-1"},
				Actor:       ActorPayload{Type: "service", ID: "svc-1"},
				Subject:     SubjectPayload{ID: "subject-2"},
				Action:      "read",
				Resource:    ResourcePayload{ID: "res-2", Name: "document"},
				Result:      ResultPayload{Status: "success"},
			},
			// environment added before id; environment keys sorted: service, span_id, trace_id
			expected: `{"action":"read","actor":{"id":"svc-1","type":"service"},"environment":{"service":"test-svc","span_id":"span-1","trace_id":"trace-1"},"id":"evt-2","resource":{"id":"res-2","name":"document"},"result":{"status":"success"},"subject":{"id":"subject-2"},"timestamp":"2024-01-01T00:00:00Z"}`,
		},
		{
			name: "with result reason",
			input: &AuditEventPayload{
				ID:        "evt-3",
				Timestamp: "2024-01-01T00:00:00Z",
				Actor:     ActorPayload{Type: "user", ID: "user-3"},
				Subject:   SubjectPayload{ID: "subject-3"},
				Action:    "delete",
				Resource:  ResourcePayload{ID: "res-3", Name: "record"},
				Result:    ResultPayload{Status: "failure", Reason: "not found"},
			},
			// result keys sorted: reason, status
			expected: `{"action":"delete","actor":{"id":"user-3","type":"user"},"id":"evt-3","resource":{"id":"res-3","name":"record"},"result":{"reason":"not found","status":"failure"},"subject":{"id":"subject-3"},"timestamp":"2024-01-01T00:00:00Z"}`,
		},
		{
			name: "with resource fields",
			input: &AuditEventPayload{
				ID:        "evt-4",
				Timestamp: "2024-01-01T00:00:00Z",
				Actor:     ActorPayload{Type: "user", ID: "user-4"},
				Subject:   SubjectPayload{ID: "subject-4"},
				Action:    "update",
				Resource:  ResourcePayload{ID: "res-4", Name: "profile", Fields: []string{"email", "name"}},
				Result:    ResultPayload{Status: "success"},
			},
			// resource keys sorted: fields, id, name
			expected: `{"action":"update","actor":{"id":"user-4","type":"user"},"id":"evt-4","resource":{"fields":["email","name"],"id":"res-4","name":"profile"},"result":{"status":"success"},"subject":{"id":"subject-4"},"timestamp":"2024-01-01T00:00:00Z"}`,
		},
		{
			name: "with metadata",
			input: &AuditEventPayload{
				ID:        "evt-5",
				Timestamp: "2024-01-01T00:00:00Z",
				Actor:     ActorPayload{Type: "user", ID: "user-5"},
				Subject:   SubjectPayload{ID: "subject-5"},
				Action:    "create",
				Resource:  ResourcePayload{ID: "res-5", Name: "item"},
				Result:    ResultPayload{Status: "success"},
				Metadata:  map[string]any{"z_key": "last", "a_key": "first"},
			},
			// metadata keys sorted: a_key, z_key; metadata appears before resource
			expected: `{"action":"create","actor":{"id":"user-5","type":"user"},"id":"evt-5","metadata":{"a_key":"first","z_key":"last"},"resource":{"id":"res-5","name":"item"},"result":{"status":"success"},"subject":{"id":"subject-5"},"timestamp":"2024-01-01T00:00:00Z"}`,
		},
		{
			name: "nil environment omitted",
			input: &AuditEventPayload{
				ID:          "evt-6",
				Timestamp:   "2024-01-01T00:00:00Z",
				Environment: nil,
				Actor:       ActorPayload{Type: "user", ID: "user-6"},
				Subject:     SubjectPayload{ID: "subject-6"},
				Action:      "create",
				Resource:    ResourcePayload{ID: "res-6", Name: "item"},
				Result:      ResultPayload{Status: "success"},
			},
			// environment key must not appear in output
			expected: `{"action":"create","actor":{"id":"user-6","type":"user"},"id":"evt-6","resource":{"id":"res-6","name":"item"},"result":{"status":"success"},"subject":{"id":"subject-6"},"timestamp":"2024-01-01T00:00:00Z"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CanonicalizeAuditEvent(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.expected {
				t.Errorf("got  %s\nwant %s", got, tt.expected)
			}
		})
	}
}

func TestCanonicalizeProtectedAuditEvent(t *testing.T) {
	tests := []struct {
		name     string
		input    *ProtectedAuditEventPayload
		expected string
	}{
		{
			name: "with protected metadata",
			input: &ProtectedAuditEventPayload{
				ID:        "evt-1",
				Timestamp: "2024-01-01T00:00:00Z",
				Actor:     ActorPayload{Type: "user", ID: "user-1"},
				Subject:   SubjectPayload{ID: "subject-1"},
				Action:    "create",
				Resource:  ResourcePayload{ID: "res-1", Name: "record"},
				Result:    ResultPayload{Status: "success"},
				ProtectedMetadata: &ProtectedMetadataPayload{
					Ciphertext: "enc-data",
					WrappedDEK: "wrapped-key",
					Commitment: "commit-hash",
				},
			},
			// protected_metadata keys sorted: ciphertext, commitment, wrapped_dek
			// top-level: action, actor, id, protected_metadata, resource, result, subject, timestamp
			expected: `{"action":"create","actor":{"id":"user-1","type":"user"},"id":"evt-1","protected_metadata":{"ciphertext":"enc-data","commitment":"commit-hash","wrapped_dek":"wrapped-key"},"resource":{"id":"res-1","name":"record"},"result":{"status":"success"},"subject":{"id":"subject-1"},"timestamp":"2024-01-01T00:00:00Z"}`,
		},
		{
			name: "nil protected metadata omitted",
			input: &ProtectedAuditEventPayload{
				ID:                "evt-2",
				Timestamp:         "2024-01-01T00:00:00Z",
				Actor:             ActorPayload{Type: "user", ID: "user-2"},
				Subject:           SubjectPayload{ID: "subject-2"},
				Action:            "read",
				Resource:          ResourcePayload{ID: "res-2", Name: "document"},
				Result:            ResultPayload{Status: "success"},
				ProtectedMetadata: nil,
			},
			// protected_metadata key must not appear
			expected: `{"action":"read","actor":{"id":"user-2","type":"user"},"id":"evt-2","resource":{"id":"res-2","name":"document"},"result":{"status":"success"},"subject":{"id":"subject-2"},"timestamp":"2024-01-01T00:00:00Z"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CanonicalizeProtectedAuditEvent(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.expected {
				t.Errorf("got  %s\nwant %s", got, tt.expected)
			}
		})
	}
}

func TestCanonicalizeMetadata(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected string
	}{
		{
			name:     "keys sorted lexicographically",
			input:    map[string]any{"z_key": "last", "a_key": "first", "m_key": "middle"},
			expected: `{"a_key":"first","m_key":"middle","z_key":"last"}`,
		},
		{
			name:     "mixed value types",
			input:    map[string]any{"flag": true, "count": 3, "label": "ok"},
			expected: `{"count":3,"flag":true,"label":"ok"}`,
		},
		{
			name:     "empty map",
			input:    map[string]any{},
			expected: `{}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CanonicalizeMetadata(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.expected {
				t.Errorf("got  %s\nwant %s", got, tt.expected)
			}
		})
	}
}

func TestCanonicalize_Determinism(t *testing.T) {
	payload := &AuditEventPayload{
		ID:          "evt-det",
		Timestamp:   "2024-06-15T12:00:00Z",
		Environment: &EnvironmentPayload{Service: "svc", TraceID: "t1", SpanID: "s1"},
		Actor:       ActorPayload{Type: "user", ID: "u1"},
		Subject:     SubjectPayload{ID: "sub1"},
		Action:      "update",
		Resource:    ResourcePayload{ID: "r1", Name: "doc", Fields: []string{"title", "body"}},
		Result:      ResultPayload{Status: "success"},
		Metadata:    map[string]any{"key": "value"},
	}

	first, err := CanonicalizeAuditEvent(payload)
	if err != nil {
		t.Fatalf("first call: unexpected error: %v", err)
	}

	for i := range 10 {
		got, err := CanonicalizeAuditEvent(payload)
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i+1, err)
		}
		if !bytes.Equal(first, got) {
			t.Errorf("call %d produced different output:\nfirst %s\ngot   %s", i+1, first, got)
		}
	}
}
