package credstest

import (
	"context"
	"testing"

	"github.com/m-lab/reboot-service/creds"
)

// Test the FakeProvider implementation.
func TestFakeProvider_AddCredentials(t *testing.T) {
	// Create a FakeProvider and add a Credentials to the map.
	provider := &FakeProvider{
		creds: map[string]*creds.Credentials{},
	}

	fakeDrac := &creds.Credentials{
		Hostname: "host",
		Username: "user",
		Password: "pass",
		Model:    "model",
		Address:  "address",
	}

	provider.AddCredentials(context.Background(), "test", fakeDrac)
	if creds, ok := provider.creds["test"]; !ok || creds != fakeDrac {
		t.Errorf("AddCredentials() didn't add the expected Credentials.")
	}
}

func TestFakeProvider_FindCredentials(t *testing.T) {
	fakeDrac := &creds.Credentials{
		Hostname: "host",
		Username: "user",
		Password: "pass",
		Model:    "model",
		Address:  "address",
	}

	provider := &FakeProvider{
		creds: map[string]*creds.Credentials{
			"test": fakeDrac,
		},
	}

	// Retrieve previously added Credentials from the FakeProvider's map.
	creds, err := provider.FindCredentials(context.Background(), "test")
	if err != nil || creds != fakeDrac {
		t.Errorf("FindCredentials() returned an error or wrong Credentials.")
	}

	// Attempt to retrieve Credentials for an unknown hostname.
	creds, err = provider.FindCredentials(context.Background(), "fail")
	if err == nil || creds != nil {
		t.Errorf("FindCredentials() didn't return an error.")
	}
}

func TestNewProvider(t *testing.T) {
	prov := NewProvider()
	if prov == nil {
		t.Errorf("NewProvider() returned nil.")
	}
}
