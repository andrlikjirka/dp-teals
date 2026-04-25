package model

// Actors (CRM users + background services)
var Actors = []Actor{
	{ActorTypeUser, "7b46e8c2-4540-4239-be95-20673c333c22"},
	{ActorTypeUser, "5d695536-c35c-4e34-a4c5-ea6cbf0ec90c"},
	{ActorTypeUser, "19945b15-2d93-4e01-a943-959f6bf2bff9"},
	{ActorTypeUser, "13236a7a-66fc-400b-a01d-98a82e0cecf5"},
	{ActorTypeUser, "fb5cafa6-e9ad-4b61-990f-d81e8469a0df"},
	{ActorTypeUser, "56f29607-98f9-4754-b8f0-b247e21609cb"},
	{ActorTypeSystem, "crm-automation-service"},
	{ActorTypeSystem, "crm-integration-service"},
}

// ActorsByType pre-group actors by type for more efficient lookup when picking actors for a Scenario
var actorsByType map[ActorType][]Actor

func init() {
	actorsByType = make(map[ActorType][]Actor)
	for _, a := range Actors {
		actorsByType[a.Type] = append(actorsByType[a.Type], a)
	}
}

// ActorsOfType returns all actors of the given type.
func ActorsOfType(t ActorType) []Actor {
	return actorsByType[t]
}

// Subjects (personal data owners = CRM customers)
var Subjects = []Subject{
	{ID: "f47ac10b-58cc-4372-a567-0e02b2c3d479"},
	{ID: "550e8400-e29b-41d4-a716-446655440000"},
	{ID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8"},
	{ID: "6ba7b811-9dad-11d1-80b4-00c04fd430c8"},
	{ID: "1e783bd9-fc98-4e6a-b3d3-870657913d27"},
	{ID: "7c9e6679-7425-40de-944b-e07fc1f90ae7"},
	{ID: "a97b2b4a-76db-4c88-8d32-e4c1e2cc6a22"},
	{ID: "b8f3a2d1-4e5f-6789-abcd-ef0123456789"},
	{ID: "c1d2e3f4-a5b6-7890-cdef-012345678901"},
	{ID: "d4e5f6a7-b8c9-0123-def0-123456789abc"},
}

// ---- Resources (CRM data entities) ----
var (
	resContact = &Resource{
		ID:     "aaaaaaaa-0001-4000-8000-000000000001",
		Name:   "Contact", // qualified individual person you are actively working with
		Fields: []string{"first_name", "last_name", "email", "phone", "address"},
	}
	resAccount = &Resource{
		ID:     "aaaaaaaa-0001-4000-8000-000000000002",
		Name:   "Account", // qualified company you have a business relationship
		Fields: []string{"company_name", "industry", "annual_revenue", "employee_count"},
	}
	resOpportunity = &Resource{
		ID:     "aaaaaaaa-0001-4000-8000-000000000003",
		Name:   "Opportunity", // qualified sales deal in progress
		Fields: []string{"title", "stage", "amount", "close_date", "probability"},
	}
	resLead = &Resource{
		ID:     "aaaaaaaa-0001-4000-8000-000000000004",
		Name:   "Lead", // unqualified prospect — raw incoming interest (web form, trade show, cold call)
		Fields: []string{"first_name", "last_name", "email", "source", "status"},
	}
	resContract = &Resource{
		ID:     "aaaaaaaa-0001-4000-8000-000000000005",
		Name:   "Contract",
		Fields: []string{"start_date", "end_date", "value", "terms"},
	}
	resMarketingConsent = &Resource{
		ID:     "aaaaaaaa-0001-4000-8000-000000000006",
		Name:   "MarketingConsent",
		Fields: []string{"email_consent", "sms_consent", "third_party_sharing"},
	}
	resCustomerReport = &Resource{
		ID:   "aaaaaaaa-0001-4000-8000-000000000007",
		Name: "CustomerReport",
	}
	resCRMSession = &Resource{
		ID:   "aaaaaaaa-0001-4000-8000-000000000008",
		Name: "CRMSession",
	}
)

// FailureReasons (per Scenario)
var FailureReasons = map[string][]string{
	"crm_auth": {
		"invalid credentials",
		"account locked after too many attempts",
		"MFA verification failed",
		"session token expired",
	},
	"crm_contact_access": {
		"contact not found",
		"insufficient permissions",
		"record access restricted by data policy",
	},
	"crm_record_mutation": {
		"insufficient permissions",
		"invalid field value",
		"record locked by another user",
		"invalid contract state transition",
	},
	"crm_consent_change": {
		"consent record not found",
		"consent already in requested state",
	},
	"crm_data_export": {
		"export limit exceeded",
		"insufficient permissions",
		"destination unreachable",
	},
}

// fieldValues maps each known resource field to a pool of realistic fake values.
var fieldValues = map[string][]any{
	// Contact / Lead
	"first_name": {"Alice", "Bob", "Carol", "David", "Eva", "Frank"},
	"last_name":  {"Smith", "Johnson", "Williams", "Brown", "Jones"},
	"email":      {"alice@example.com", "bob@example.com", "carol@example.com", "david@example.com"},
	"phone":      {"+1-555-0101", "+1-555-0202", "+1-555-0303", "+1-555-0404"},
	"address":    {"123 Main St", "456 Oak Ave", "789 Pine Rd", "321 Elm Blvd"},
	"source":     {"web", "referral", "cold_call", "trade_show", "social_media"},
	"status":     {"new", "contacted", "qualified", "unqualified", "converted"},
	// Account
	"company_name":   {"Acme Corp", "Globex Inc", "Initech", "Umbrella Ltd", "Stark Industries"},
	"industry":       {"technology", "finance", "healthcare", "retail", "manufacturing"},
	"annual_revenue": {50000, 120000, 500000, 1200000, 5000000},
	"employee_count": {5, 25, 100, 500, 2000},
	// Opportunity
	"title":       {"Enterprise Deal", "Renewal Q1", "Upsell - Premium", "New Business"},
	"stage":       {"prospecting", "qualification", "proposal", "negotiation", "closed_won", "closed_lost"},
	"amount":      {1500, 8000, 25000, 75000, 200000},
	"close_date":  {"2026-06-30", "2026-09-30", "2026-12-31", "2027-03-31"},
	"probability": {10, 25, 50, 75, 90},
	// Contract
	"start_date": {"2026-01-01", "2026-04-01", "2026-07-01", "2026-10-01"},
	"end_date":   {"2027-01-01", "2027-04-01", "2027-07-01", "2027-10-01"},
	"value":      {5000, 15000, 50000, 120000},
	"terms":      {"standard", "custom", "enterprise", "trial"},
	// MarketingConsent
	"email_consent":       {true, false},
	"sms_consent":         {true, false},
	"third_party_sharing": {true, false},
}
