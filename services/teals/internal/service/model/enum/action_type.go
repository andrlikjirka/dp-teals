package enum

type ActionType string

const (
	ActionTypeAccess    ActionType = "ACCESS"
	ActionTypeCreate    ActionType = "CREATE"
	ActionTypeUpdate    ActionType = "UPDATE"
	ActionTypeDelete    ActionType = "DELETE"
	ActionTypeShare     ActionType = "SHARE"
	ActionTypeExport    ActionType = "EXPORT"
	ActionTypeLogin     ActionType = "LOGIN"
	ActionTypeLogout    ActionType = "LOGOUT"
	ActionTypeAuthorize ActionType = "AUTHORIZE"
)

// IsValid checks if the ActionType is one of the defined constants.
func (a ActionType) IsValid() bool {
	switch a {
	case ActionTypeAccess, ActionTypeCreate, ActionTypeUpdate, ActionTypeDelete, ActionTypeLogin, ActionTypeLogout, ActionTypeShare, ActionTypeExport, ActionTypeAuthorize:
		return true
	default:
		return false
	}
}
