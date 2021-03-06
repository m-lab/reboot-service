package credstest

import (
	"context"
	"fmt"
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

func TestFakeProvider_ListCredentials(t *testing.T) {
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
	creds, err := provider.ListCredentials(context.Background())
	if err != nil {
		t.Errorf("ListCredentials() returned an error")
	}
	fmt.Println(creds[0])
	if len(creds) != 1 || *creds[0] != *fakeDrac {
		t.Errorf("ListCredentials() didn't return the expected Credentials.")
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

func TestFakeProvider_DeleteCredentials(t *testing.T) {
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

	err := provider.DeleteCredentials(context.Background(), "test")
	if err != nil {
		t.Errorf("DeleteCredentials() returned an error: %v", err)
	}

	// This should fail the second time as the entity has been removed.
	err = provider.DeleteCredentials(context.Background(), "test")
	if err == nil {
		t.Errorf("DeleteCredentials() - expected err, got nil.")
	}

}

func TestFakeProvider_Close(t *testing.T) {
	provider := &FakeProvider{}
	provider.Close()
}
