package model

import (
	"fmt"
	"math/rand"

	ingestionv1 "github.com/andrlikjirka/dp-teals/gen/audit/v1"
)

// Scenario defines a user activity Scenario with specific characteristics for generating audit events.
type Scenario struct {
	Name        string
	Actions     []ingestionv1.Action
	ActorTypes  []ingestionv1.Actor_Type
	Resources   []*Resource
	MetaTmpl    metadataTemplate // nil = no metadata
	FailureProb float64
	SelfSubject bool // subject.ID == actor.ID (auth scenarios)
}

var Scenarios = []Scenario{
	{
		Name:        "crm_auth",
		Actions:     []ingestionv1.Action{ingestionv1.Action_ACTION_LOGIN, ingestionv1.Action_ACTION_LOGOUT},
		ActorTypes:  []ingestionv1.Actor_Type{ingestionv1.Actor_TYPE_USER},
		Resources:   []*Resource{resCRMSession},
		MetaTmpl:    tmplAuthAttempt,
		FailureProb: 0.3,
		SelfSubject: true,
	},
	{
		Name:        "crm_contact_access",
		Actions:     []ingestionv1.Action{ingestionv1.Action_ACTION_ACCESS},
		ActorTypes:  []ingestionv1.Actor_Type{ingestionv1.Actor_TYPE_USER, ingestionv1.Actor_TYPE_SYSTEM},
		Resources:   []*Resource{resContact, resLead},
		MetaTmpl:    tmplIPUserAgent,
		FailureProb: 0.05,
	},
	{
		Name:        "crm_record_mutation",
		Actions:     []ingestionv1.Action{ingestionv1.Action_ACTION_CREATE, ingestionv1.Action_ACTION_UPDATE, ingestionv1.Action_ACTION_DELETE},
		ActorTypes:  []ingestionv1.Actor_Type{ingestionv1.Actor_TYPE_USER, ingestionv1.Actor_TYPE_SYSTEM},
		Resources:   []*Resource{resContact, resAccount, resOpportunity, resLead, resContract},
		MetaTmpl:    tmplChangedFields, // only applied when action == UPDATE
		FailureProb: 0.10,
	},
	{
		Name:        "crm_consent_change",
		Actions:     []ingestionv1.Action{ingestionv1.Action_ACTION_UPDATE},
		ActorTypes:  []ingestionv1.Actor_Type{ingestionv1.Actor_TYPE_USER},
		Resources:   []*Resource{resMarketingConsent},
		MetaTmpl:    nil,
		FailureProb: 0.05,
	},
	{
		Name:        "crm_data_export",
		Actions:     []ingestionv1.Action{ingestionv1.Action_ACTION_SHARE, ingestionv1.Action_ACTION_EXPORT},
		ActorTypes:  []ingestionv1.Actor_Type{ingestionv1.Actor_TYPE_USER},
		Resources:   []*Resource{resContact, resContract, resCustomerReport},
		MetaTmpl:    tmplExportInfo,
		FailureProb: 0.20,
	},
}

// metadataTemplate generates metadata for an event based on the Resource and Actor involved.
type metadataTemplate func(res *Resource, act *Actor) map[string]any

// tmplAuthAttempt simulates metadata for an authentication attempt, including the number of attempts and whether MFA was used.
func tmplAuthAttempt(res *Resource, act *Actor) map[string]any {
	return map[string]any{
		"attempt_count": rand.Intn(5) + 1,
		"mfa_used":      rand.Intn(2) == 1,
	}
}

// tmplIPUserAgent simulates metadata for an access event, including a random IP address and user agent string.
func tmplIPUserAgent(res *Resource, act *Actor) map[string]any {
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
	}
	return map[string]any{
		"ip_address": fmt.Sprintf("10.0.%d.%d", rand.Intn(255), rand.Intn(255)),
		"user_agent": userAgents[rand.Intn(len(userAgents))],
	}
}

// tmplChangedFields simulates metadata for an update event, randomly selecting a subset of the resource's fields as changed and providing dummy previous and new values for those fields.
func tmplChangedFields(res *Resource, act *Actor) map[string]any {
	if len(res.Fields) == 0 {
		return map[string]any{}
	}
	n := rand.Intn(min(2, len(res.Fields))) + 1
	perm := rand.Perm(len(res.Fields))
	changed := make([]any, n)
	prev := map[string]any{}
	next := map[string]any{}
	for i := range n {
		f := res.Fields[perm[i]]
		changed[i] = f
		prev[f], next[f] = pickOldNew(f)
	}
	return map[string]any{
		"changed_fields": changed,
		"previous_state": prev,
		"new_state":      next,
	}
}

// tmplExportInfo simulates metadata for a data export event, including the export format, destination, and number of records exported.
func tmplExportInfo(res *Resource, act *Actor) map[string]any {
	formats := []string{"pdf", "csv", "xlsx"}
	destinations := []string{"email", "s3", "sftp"}
	return map[string]any{
		"export_format": formats[rand.Intn(len(formats))],
		"destination":   destinations[rand.Intn(len(destinations))],
		"record_count":  rand.Intn(500) + 1,
	}
}

// pickOldNew returns two distinct values from the pool for a given field.
func pickOldNew(field string) (old, new any) {
	pool, ok := fieldValues[field]
	if !ok || len(pool) < 2 {
		pool = fallbackValues
	}
	// Pick old value
	oldIdx := rand.Intn(len(pool))
	old = pool[oldIdx]
	// Pick new value — must differ from old
	newIdx := oldIdx
	for newIdx == oldIdx {
		newIdx = rand.Intn(len(pool))
	}
	new = pool[newIdx]
	return old, new
}

// fallbackValues is used for any field not listed above.
var fallbackValues = []any{"value_a", "value_b", "value_c", "value_d"}
