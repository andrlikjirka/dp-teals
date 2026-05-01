package repository_test

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/andrlikjirka/dp-teals/services/teals/internal/infrastructure/repository"
	svcerrors "github.com/andrlikjirka/dp-teals/services/teals/internal/service/errors"
	svcmodel "github.com/andrlikjirka/dp-teals/services/teals/internal/service/model"
	"github.com/andrlikjirka/dp-teals/services/teals/internal/service/model/enum"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mmrLeafCounter generates unique leaf_index values across all insertMMRNode calls. leaf_index has a UNIQUE constraint, so each inserted node needs a distinct value.
var mmrLeafCounter atomic.Int64

// insertMMRNode inserts a minimal mmr_node row with a unique leaf_index and returns the generated node ID. Every log_entry requires a real mmr_node due to the FK constraint.
func insertMMRNode(t *testing.T) int64 {
	t.Helper()
	leafIndex := mmrLeafCounter.Add(1)
	var id int64
	err := testPool.QueryRow(context.Background(),
		`INSERT INTO teals.mmr_node (hash, level, leaf_index) VALUES ($1, 0, $2) RETURNING id`,
		[]byte("testhash"), leafIndex,
	).Scan(&id)
	require.NoError(t, err)
	return id
}

// logEntryFixture builds a minimal valid JSONB payload for a log entry.
type logEntryFixture struct {
	action       string
	actorType    string
	actorID      string
	subjectID    string
	resourceID   string
	resourceName string
	resultStatus string
	timestamp    time.Time
}

func defaultFixture() logEntryFixture {
	return logEntryFixture{
		action:       string(enum.ActionTypeAccess),
		actorType:    string(enum.ActorTypeUser),
		actorID:      "default-actor",
		subjectID:    "default-subject",
		resourceID:   "default-resource-id",
		resourceName: "default-resource-name",
		resultStatus: string(enum.ResultStatusSuccess),
		timestamp:    time.Now().UTC(),
	}
}

func (f logEntryFixture) toPayload() json.RawMessage {
	m := map[string]any{
		"action":    f.action,
		"actor":     map[string]any{"type": f.actorType, "id": f.actorID},
		"subject":   map[string]any{"id": f.subjectID},
		"resource":  map[string]any{"id": f.resourceID, "name": f.resourceName},
		"result":    map[string]any{"status": f.resultStatus},
		"timestamp": f.timestamp.Format("2006-01-02T15:04:05.000000Z"),
	}
	b, err := json.Marshal(m)
	if err != nil {
		panic("marshal log entry fixture: " + err.Error())
	}
	return b
}

// logEntrySetup encapsulates common setup for log entry tests, such as inserting a producer key and providing a repository instance.
type logEntrySetup struct {
	producerKeyID uuid.UUID
	repo          *repository.AuditLogRepository
}

func newLogEntrySetup(t *testing.T) logEntrySetup {
	t.Helper()
	producerID := uuid.New()
	insertProducer(t, producerID)
	key := newProducerKey(producerID)
	keyRepo := repository.NewProducerKeyRepository(testPool)
	require.NoError(t, keyRepo.AddPublicKey(context.Background(), key))
	return logEntrySetup{
		producerKeyID: key.ID,
		repo:          repository.NewAuditLogRepository(testPool),
	}
}

// storeEntry inserts a log entry using the given fixture and returns its event ID.
func (s logEntrySetup) storeEntry(t *testing.T, f logEntryFixture) uuid.UUID {
	t.Helper()
	eventID := uuid.New()
	nodeID := insertMMRNode(t)
	err := s.repo.StoreAuditLogEntry(
		context.Background(),
		eventID,
		f.toPayload(),
		"sig-token",
		s.producerKeyID,
		nodeID,
		[]byte("salt"),
	)
	require.NoError(t, err)
	return eventID
}

func TestAuditLogRepository_StoreAuditLogEntry(t *testing.T) {
	ctx := context.Background()

	t.Run("StoresEntry", func(t *testing.T) {
		truncateTables(t)
		s := newLogEntrySetup(t)

		err := s.repo.StoreAuditLogEntry(ctx, uuid.New(), defaultFixture().toPayload(),
			"sig-token", s.producerKeyID, insertMMRNode(t), []byte("salt"))

		require.NoError(t, err)
	})

	t.Run("DuplicateEventIDReturnsError", func(t *testing.T) {
		truncateTables(t)
		s := newLogEntrySetup(t)
		eventID := uuid.New()
		require.NoError(t, s.repo.StoreAuditLogEntry(ctx, eventID, defaultFixture().toPayload(),
			"sig-token", s.producerKeyID, insertMMRNode(t), []byte("salt")))

		err := s.repo.StoreAuditLogEntry(ctx, eventID, defaultFixture().toPayload(),
			"sig-token", s.producerKeyID, insertMMRNode(t), []byte("salt"))

		assert.ErrorIs(t, err, svcerrors.ErrDuplicateEventID)
	})
}

func TestAuditLogRepository_GetAuditLogEntryByEventID(t *testing.T) {
	ctx := context.Background()

	t.Run("ReturnsEntry", func(t *testing.T) {
		truncateTables(t)
		s := newLogEntrySetup(t)
		payload := defaultFixture().toPayload()
		salt := []byte("salt")
		eventID := uuid.New()
		nodeID := insertMMRNode(t)
		require.NoError(t, s.repo.StoreAuditLogEntry(ctx, eventID, payload,
			"sig-token", s.producerKeyID, nodeID, salt))

		got, err := s.repo.GetAuditLogEntryByEventID(ctx, eventID)

		require.NoError(t, err)
		assert.Equal(t, eventID, got.EventID)
		assert.Equal(t, s.producerKeyID, got.ProducerKeyID)
		assert.Equal(t, "sig-token", got.SignatureToken)
		assert.JSONEq(t, string(payload), string(got.Payload))
	})

	t.Run("NotFound", func(t *testing.T) {
		truncateTables(t)
		repo := repository.NewAuditLogRepository(testPool)

		_, err := repo.GetAuditLogEntryByEventID(ctx, uuid.New())

		assert.ErrorIs(t, err, svcerrors.ErrAuditLogEntryNotFound)
	})
}

func TestAuditLogRepository_ListAuditLogEntries(t *testing.T) {
	ctx := context.Background()
	emptyFilter := &svcmodel.AuditEventFilter{}

	t.Run("ReturnsAllEntries", func(t *testing.T) {
		truncateTables(t)
		s := newLogEntrySetup(t)
		s.storeEntry(t, defaultFixture())
		s.storeEntry(t, defaultFixture())

		results, err := s.repo.ListAuditLogEntries(ctx, emptyFilter, nil, 10)

		require.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("ReturnsEmptySliceWhenNoEntries", func(t *testing.T) {
		truncateTables(t)
		repo := repository.NewAuditLogRepository(testPool)

		results, err := repo.ListAuditLogEntries(ctx, emptyFilter, nil, 10)

		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("FilterByAction", func(t *testing.T) {
		truncateTables(t)
		s := newLogEntrySetup(t)
		f := defaultFixture()
		f.action = string(enum.ActionTypeCreate)
		matchID := s.storeEntry(t, f)
		f.action = string(enum.ActionTypeDelete)
		s.storeEntry(t, f)

		results, err := s.repo.ListAuditLogEntries(ctx,
			&svcmodel.AuditEventFilter{Actions: []enum.ActionType{enum.ActionTypeCreate}},
			nil, 10)

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, matchID, results[0].EventID)
	})

	t.Run("FilterByActorType", func(t *testing.T) {
		truncateTables(t)
		s := newLogEntrySetup(t)
		f := defaultFixture()
		f.actorType = string(enum.ActorTypeUser)
		matchID := s.storeEntry(t, f)
		f.actorType = string(enum.ActorTypeSystem)
		s.storeEntry(t, f)

		results, err := s.repo.ListAuditLogEntries(ctx,
			&svcmodel.AuditEventFilter{ActorTypes: []enum.ActorType{enum.ActorTypeUser}},
			nil, 10)

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, matchID, results[0].EventID)
	})

	t.Run("FilterByActorID", func(t *testing.T) {
		truncateTables(t)
		s := newLogEntrySetup(t)
		f := defaultFixture()
		f.actorID = "actor-match"
		matchID := s.storeEntry(t, f)
		f.actorID = "actor-other"
		s.storeEntry(t, f)

		results, err := s.repo.ListAuditLogEntries(ctx,
			&svcmodel.AuditEventFilter{ActorID: "actor-match"},
			nil, 10)

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, matchID, results[0].EventID)
	})

	t.Run("FilterBySubjectID", func(t *testing.T) {
		truncateTables(t)
		s := newLogEntrySetup(t)
		f := defaultFixture()
		f.subjectID = "subject-match"
		matchID := s.storeEntry(t, f)
		f.subjectID = "subject-other"
		s.storeEntry(t, f)

		results, err := s.repo.ListAuditLogEntries(ctx,
			&svcmodel.AuditEventFilter{SubjectID: "subject-match"},
			nil, 10)

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, matchID, results[0].EventID)
	})

	t.Run("FilterByResourceID", func(t *testing.T) {
		truncateTables(t)
		s := newLogEntrySetup(t)
		f := defaultFixture()
		f.resourceID = "resource-match"
		matchID := s.storeEntry(t, f)
		f.resourceID = "resource-other"
		s.storeEntry(t, f)

		results, err := s.repo.ListAuditLogEntries(ctx,
			&svcmodel.AuditEventFilter{ResourceID: "resource-match"},
			nil, 10)

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, matchID, results[0].EventID)
	})

	t.Run("FilterByResourceName", func(t *testing.T) {
		truncateTables(t)
		s := newLogEntrySetup(t)
		f := defaultFixture()
		f.resourceName = "name-match"
		matchID := s.storeEntry(t, f)
		f.resourceName = "name-other"
		s.storeEntry(t, f)

		results, err := s.repo.ListAuditLogEntries(ctx,
			&svcmodel.AuditEventFilter{ResourceName: "name-match"},
			nil, 10)

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, matchID, results[0].EventID)
	})

	t.Run("FilterByResultStatus", func(t *testing.T) {
		truncateTables(t)
		s := newLogEntrySetup(t)
		f := defaultFixture()
		f.resultStatus = string(enum.ResultStatusSuccess)
		matchID := s.storeEntry(t, f)
		f.resultStatus = string(enum.ResultStatusFailure)
		s.storeEntry(t, f)

		results, err := s.repo.ListAuditLogEntries(ctx,
			&svcmodel.AuditEventFilter{ResultStatuses: []enum.ResultStatusType{enum.ResultStatusSuccess}},
			nil, 10)

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, matchID, results[0].EventID)
	})

	t.Run("FilterByTimestampFrom", func(t *testing.T) {
		truncateTables(t)
		s := newLogEntrySetup(t)
		base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
		f := defaultFixture()
		f.timestamp = base.Add(1 * time.Hour) // after threshold — matches
		matchID := s.storeEntry(t, f)
		f.timestamp = base.Add(-1 * time.Hour) // before threshold — excluded
		s.storeEntry(t, f)

		results, err := s.repo.ListAuditLogEntries(ctx,
			&svcmodel.AuditEventFilter{TimestampFrom: &base},
			nil, 10)

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, matchID, results[0].EventID)
	})

	t.Run("FilterByTimestampTo", func(t *testing.T) {
		truncateTables(t)
		s := newLogEntrySetup(t)
		base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
		f := defaultFixture()
		f.timestamp = base.Add(-1 * time.Hour) // before threshold — matches
		matchID := s.storeEntry(t, f)
		f.timestamp = base.Add(1 * time.Hour) // after threshold — excluded
		s.storeEntry(t, f)

		results, err := s.repo.ListAuditLogEntries(ctx,
			&svcmodel.AuditEventFilter{TimestampTo: &base},
			nil, 10)

		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.Equal(t, matchID, results[0].EventID)
	})

	t.Run("CursorSkipsEntriesUpToAndIncludingCursor", func(t *testing.T) {
		truncateTables(t)
		s := newLogEntrySetup(t)
		s.storeEntry(t, defaultFixture())
		s.storeEntry(t, defaultFixture())
		s.storeEntry(t, defaultFixture())

		// fetch all to get IDs in insertion order
		all, err := s.repo.ListAuditLogEntries(ctx, emptyFilter, nil, 10)
		require.NoError(t, err)
		require.Len(t, all, 3)

		cursor := all[0].ID

		results, err := s.repo.ListAuditLogEntries(ctx, emptyFilter, cursor, 10)

		require.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Equal(t, all[1].EventID, results[0].EventID)
		assert.Equal(t, all[2].EventID, results[1].EventID)
	})

	t.Run("SizeLimitsNumberOfReturnedEntries", func(t *testing.T) {
		truncateTables(t)
		s := newLogEntrySetup(t)
		s.storeEntry(t, defaultFixture())
		s.storeEntry(t, defaultFixture())
		s.storeEntry(t, defaultFixture())

		results, err := s.repo.ListAuditLogEntries(ctx, emptyFilter, nil, 2)

		require.NoError(t, err)
		assert.Len(t, results, 2)
	})
}
