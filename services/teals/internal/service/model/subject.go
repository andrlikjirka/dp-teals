package model

import (
	"time"
)

// ForgetSubjectResult represents the result of a subject forgetting operation, containing the subject ID and the timestamp when the subject was forgotten. This struct can be used to return relevant information about the forgetting operation, such as confirming which subject was forgotten and when the operation took place.
type ForgetSubjectResult struct {
	SubjectID   string
	ForgottenAt time.Time
}
