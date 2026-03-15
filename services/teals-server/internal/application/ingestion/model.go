package ingestion

type AppendEventInput struct {
	EventID string
}

type AppendEventOutput struct {
	Success bool
}
