package whatsapp

import (
	"go.mau.fi/whatsmeow/types"
)

// PhoneToJID converte numero de telefone para JID do WhatsApp
func PhoneToJID(phone string) types.JID {
	user := phone
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
