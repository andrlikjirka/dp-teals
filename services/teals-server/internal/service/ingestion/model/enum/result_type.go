package enum

type ResultStatusType string

const (
	ResultTypeSuccess ResultStatusType = "SUCCESS"
	ResultTypeFailure ResultStatusType = "FAILURE"
)

// IsValid checks if the ResultStatusType is one of the defined constants.
func (a ResultStatusType) IsValid() bool {
	switch a {
	case ResultTypeSuccess, ResultTypeFailure:
		return true
	default:
		return false
	}
}
