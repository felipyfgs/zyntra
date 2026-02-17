package whatsapp

import (
	"strings"

	"go.mau.fi/whatsmeow/types"
)

// PhoneToJID converts phone number OR JID/LID string to WhatsApp JID.
// Handles:
//   - JID format: "559981769536@s.whatsapp.net" -> parsed directly
//   - LID format: "131357900538070@lid" -> parsed directly
//   - Phone format: "+559981769536" or "559981769536" -> converted to JID
func PhoneToJID(input string) types.JID {
	// If contains @, it's already a JID or LID - parse directly
	if strings.Contains(input, "@") {
		jid, err := types.ParseJID(input)
		if err == nil {
			return jid
		}
		// Fallback: extract user part before @
		parts := strings.Split(input, "@")
		return types.NewJID(parts[0], types.DefaultUserServer)
	}

	// Phone number format - convert to JID
	user := input
	if len(user) > 0 && user[0] == '+' {
		user = user[1:]
	}
	return types.NewJID(user, types.DefaultUserServer)
}

// JIDToPhone converte JID para numero de telefone
func JIDToPhone(jid types.JID) string {
	if jid.User == "" {
		return ""
	}
	return "+" + jid.User
}

// ParseJID faz parse de string JID
func ParseJID(jidStr string) (types.JID, error) {
	return types.ParseJID(jidStr)
}

// IsValidPhone verifica se e um numero valido
func IsValidPhone(phone string) bool {
	if len(phone) < 10 {
		return false
	}
	start := 0
	if phone[0] == '+' {
		start = 1
	}
	for i := start; i < len(phone); i++ {
		if phone[i] < '0' || phone[i] > '9' {
			return false
		}
	}
	return true
}
