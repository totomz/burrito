package common

import (
	"testing"
)

func TestParseBaoURN(t *testing.T) {
	if IsCICD() {
		t.Skip("integration test that requires OpenBAO")
	}

	secret, err := GetSecretBao(t.Context(), "bao://totomz/massive#key")
	if err != nil {
		t.Fatal(err)
	}
	if secret == "" {
		t.Fatal("secret not found")
	}

	secret, err = GetSecretBao(t.Context(), "bao://totomz/massive")
	if err == nil {
		t.Fatal("no fragment: expected an error")
	}

	secret, err = GetSecretBao(t.Context(), "bao://totomz/pippo#cane")
	if err == nil {
		t.Fatal("nx secret: expected an error")
	}
}
