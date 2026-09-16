package utils

import (
	"strings"
	"testing"
)

func TestArgon2PasswordHash(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Fatalf("unexpected password format: %s", hash)
	}
	if !CheckPasswordHash("correct horse battery staple", hash) {
		t.Fatal("correct password was rejected")
	}
	if CheckPasswordHash("wrong", hash) {
		t.Fatal("wrong password was accepted")
	}
}

func TestMachineCredentialHashAndDerivation(t *testing.T) {
	credential := DeriveCredential("master-secret", "agent", "enrollment-token")
	if credential == "" || credential == DeriveCredential("other", "agent", "enrollment-token") {
		t.Fatal("credential derivation is not keyed")
	}
	encoded := HashCredential(credential)
	if strings.Contains(encoded, credential) || !CheckCredential(credential, encoded) {
		t.Fatal("credential hash did not verify safely")
	}
	if CheckCredential("wrong", encoded) || CheckCredential(credential, credential) {
		t.Fatal("invalid credential was accepted")
	}
}
