package generator

import (
	"context"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"github.com/andrlikjirka/dp-teals/pkg/logger"
	model2 "github.com/andrlikjirka/dp-teals/tools/generator/internal/model"
	"github.com/google/uuid"
)

// Generator is responsible for creating random audit events based on predefined scenarios.
type Generator struct {
	signer signer
	sender sender
	log    *logger.Logger
}

// NewGenerator initializes a new Generator instance with the provided logger.
func NewGenerator(signer signer, sender sender, log *logger.Logger) *Generator {
	return &Generator{
		signer: signer,
		sender: sender,
		log:    log,
	}
}

// Run generates a specified number of audit events with a delay between each generation. It logs the outcome of each event generation and returns an error if any events failed to generate successfully.
func (g *Generator) Run(ctx context.Context, count int, delayMs int) error {
	m := metrics{total: count}
	runStart := time.Now()

	for i := range count {
		genStart := time.Now()
		event, err := buildAuditEvent()

		token := ""
		if g.signer != nil {
			token, err = g.signer.Sign(event)
			if err != nil {
				m.failed++
				g.log.Error("failed to sign event", "index", i, "error", err)
				continue
			}
		}
		m.genDur += time.Since(genStart)
		if err != nil {
			g.log.Error("failed to build event", "index", i, "error", err)
			m.failed++
			continue
		}

		sendStart := time.Now()
		res, err := g.sender.send(ctx, event, token)
		m.sendDur += time.Since(sendStart)

		if err != nil {
			m.failed++
			g.log.Error("failed to send event", "index", i, "error", err)
			continue
		}

		m.succeeded++
		g.log.Info("event sent", "progress", fmt.Sprintf("%d/%d", i+1, count),
			"event_id", event.ID,
			"ledger_size", res.LedgerSize,
			"appended_at", res.Timestamp.Format(time.RFC3339),
		)

		if delayMs > 0 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
	}

	m.totalDur = time.Since(runStart)
	g.logMetrics(m)

	if m.failed > 0 {
		return fmt.Errorf("%d/%d events failed", m.failed, count)
	}
	return nil
}

// buildAuditEvent constructs a random audit event by picking a random scenario and then randomly selecting appropriate values for each field based on that scenario's configuration.
func buildAuditEvent() (*model2.AuditEvent, error) {
	sc := model2.Scenarios[rand.Intn(len(model2.Scenarios))]
	a := pickAction(sc)
	r := pickResource(sc)
	act := pickActor(sc)
	s := pickSubject(sc)
	res := pickResult(sc)
	e := buildEnvironment()
	metadata, err := buildMetadata(sc, a, r, act)
	if err != nil {
		return nil, fmt.Errorf("build metadata: %w", err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generate event ID: %w", err)
	}
	return &model2.AuditEvent{
		ID:          id,
		Timestamp:   time.Now().UTC(),
		Environment: e,
		Actor:       *act,
		Subject:     *s,
		Action:      a,
		Resource:    *r,
		Result:      *res,
		Metadata:    metadata,
	}, nil
}

// buildMetadata generates metadata for an event based on the scenario's metadata template function, if defined. It returns nil if no metadata should be generated for this scenario/action combination.
func buildMetadata(sc model2.Scenario, action model2.ActionType, res *model2.Resource, act *model2.Actor) (map[string]any, error) {
	if sc.Name == "crm_record_mutation" && action != model2.ActionUpdate {
		return nil, nil // only generate metadata for UPDATE actions in this scenario
	}
	if sc.MetaTmpl == nil {
		return nil, nil
	}
	raw := sc.MetaTmpl(res, act)
	return raw, nil
}

// pickAction randomly selects an action from the scenario's list of possible actions.
func pickAction(sc model2.Scenario) model2.ActionType {
	return sc.Actions[rand.Intn(len(sc.Actions))]
}

// pickActor randomly selects an actor from the allowed actor types defined in the scenario. It first compiles a list of all actors that match the scenario's ActorTypes, then picks one at random.
func pickActor(sc model2.Scenario) *model2.Actor {
	var allowed []model2.Actor
	for _, t := range sc.ActorTypes {
		allowed = append(allowed, model2.ActorsOfType(t)...)
	}
	return &allowed[rand.Intn(len(allowed))]
}

// pickSubject randomly selects a subject for the event. For authentication scenarios (where SelfSubject is true), the subject is the same as the actor (i.e., the user is acting on themselves). For other scenarios, it picks a random subject from the global Subjects list.
func pickSubject(sc model2.Scenario) *model2.Subject {
	// for auth scenarios, subject == actor
	if sc.SelfSubject {
		actor := pickActor(sc)
		return &model2.Subject{ID: actor.ID}
	}
	return &model2.Subjects[rand.Intn(len(model2.Subjects))]
}

// pickResource randomly selects a resource from the scenario's list of possible resources.
func pickResource(sc model2.Scenario) *model2.Resource {
	return sc.Resources[rand.Intn(len(sc.Resources))]
}

// pickResult determines the result of the action based on the scenario's FailureProb. It randomly decides if the action was a success or failure, and if it's a failure, it picks a random reason from the predefined FailureReasons for that scenario.
func pickResult(sc model2.Scenario) *model2.Result {
	if rand.Float64() < sc.FailureProb {
		reasons := model2.FailureReasons[sc.Name]
		return &model2.Result{
			Status: model2.ResultStatusFailure,
			Reason: reasons[rand.Intn(len(reasons))],
		}
	}
	return &model2.Result{
		Status: model2.ResultStatusSuccess,
	}
}

// buildEnvironment constructs an environment object with a fixed service name and random trace and span IDs. The trace ID is 32 hex characters (16 bytes) and the span ID is 16 hex characters (8 bytes), following common tracing conventions.
func buildEnvironment() *model2.Environment {
	return &model2.Environment{
		Service: "crm_service",
		TraceID: randomHex(32),
		SpanID:  randomHex(16),
	}
}

// randomHex generates a random hexadecimal string of length 2n (representing n bytes). It uses crypto/rand for secure random byte generation and encodes the bytes as a lowercase hex string.
func randomHex(n int) string {
	b := make([]byte, n)
	_, _ = crand.Read(b)
	return hex.EncodeToString(b)
}

// logMetrics logs the performance metrics of the generation run, including total events, successes, failures, total time taken, time spent on generation and sending, and average send time per successful event.
func (g *Generator) logMetrics(m metrics) {
	var avgSend time.Duration
	if m.succeeded > 0 {
		avgSend = m.sendDur / time.Duration(m.succeeded)
	}
	g.log.Info("generator metrics",
		"total", m.total,
		"succeeded", m.succeeded,
		"failed", m.failed,
		"total_time", m.totalDur.Round(time.Millisecond),
		"gen_time", m.genDur.Round(time.Millisecond),
		"send_time", m.sendDur.Round(time.Millisecond),
		"avg_send", avgSend.Round(time.Microsecond),
	)
}
