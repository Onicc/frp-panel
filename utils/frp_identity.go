package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// FRPClientUser gives each physical Client its own FRPS proxy namespace.
// User names cannot contain dots, so the first segment remains the account
// name used by the FRPS authentication plugin.
func FRPClientUser(username, clientID string) string {
	digest := sha256.Sum256([]byte(clientID))
	return username + ".c" + hex.EncodeToString(digest[:12])
}

func FRPAccountName(frpUser string) string {
	account, _, _ := strings.Cut(frpUser, ".")
	return account
}
