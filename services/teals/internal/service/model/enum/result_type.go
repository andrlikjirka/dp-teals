package enum

type ResultStatusType string

const (
	ResultStatusSuccess ResultStatusType = "SUCCESS"
	ResultStatusFailure ResultStatusType = "FAILURE"
)

// IsValid checks if the ResultStatusType is one of the defined constants.
func (a ResultStatusType) IsValid() bool {
	switch a {
	case ResultStatusSuccess, ResultStatusFailure:
		return true
	default:
		return false
	}
}
