package burrito_common

import (
	"testing"
)

func TestMustGetStringFail(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			// The function is expected to panic()
			// so we're happy to have recovered
		}
	}()

	value := MustGetString("whatever")
	t.Errorf("suppsed to panic, got %s", value)

}

func TestMustGetStringSuccess(t *testing.T) {
	println("ciao")
	value := MustGetString("provider.outlook.oauth.clientId")

	if value != "clientid" {
		t.Errorf("expected 'clientId', got %s", value)
	}
}
