package utils

const (
	RolePatient   = 47293
	RoleClinician = 82651
	RoleCaretaker = 61847
)

func RoleToString(role int) string {
	switch role {
	case RolePatient:
		return "patient"
	case RoleClinician:
		return "clinician"
	case RoleCaretaker:
		return "caretaker"
	default:
		return "unknown"
	}
}

func StringToRole(roleStr string) int {
	switch roleStr {
	case "patient":
		return RolePatient
	case "clinician":
		return RoleClinician
	case "caretaker":
		return RoleCaretaker
	default:
		return 0
	}
}

func IsValidRole(role int) bool {
	return role == RolePatient || role == RoleClinician || role == RoleCaretaker
}
