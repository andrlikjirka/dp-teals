package enum

type ActorType string

const (
	ActorTypeUser   ActorType = "USER"
	ActorTypeSystem ActorType = "SYSTEM"
)

// IsValid checks if the ActorType is one of the defined valid values.
func (a ActorType) IsValid() bool {
	switch a {
	case ActorTypeUser, ActorTypeSystem:
		return true
	default:
		return false
	}
}
