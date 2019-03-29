package reboot

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/m-lab/reboot-service/connector"
	"github.com/m-lab/reboot-service/creds"
)

// Mock structs for Connector and Connection interfaces.
type mockConnector struct{}

type mockConnection struct{}

func (connector *mockConnector) NewConnection(*connector.ConnectionConfig) (connector.Connection, error) {
	return &mockConnection{}, nil
}

func (connection *mockConnection) Reboot() (string, error) {
	return "Server power operation successful", nil
}

func (connection *mockConnection) Close() error {
	return nil
}

// Mock struct for credentials Provider
type mockProvider struct{}

func (p *mockProvider) FindCredentials(context.Context, string) (*creds.Credentials, error) {
	return &creds.Credentials{
		Hostname: "testhost",
		Username: "testuser",
		Password: "testpass",
		Model:    "drac",
		Address:  "testaddr",
	}, nil
}

func TestHandler_ServeHTTP(t *testing.T) {
	h := &Handler{
		config: &Config{
			ProjectID:      "test",
			PrivateKeyPath: "",
			DRACPort:       806,
			SSHPort:        22,
			Namespace:      "test",
		},
		connector:     &mockConnector{},
		credsProvider: &mockProvider{},
	}

	rr := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "/v1/reboot?host=test", nil)
	if err != nil {
		t.Errorf("Can't create test HTTP request: %v", err)
	}

	h.ServeHTTP(rr, req)
	resp := rr.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("ServeHTTP() - wrong status code (expected %v, got %v)",
			http.StatusOK, resp.StatusCode)
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("ServeHTTP() - cannot read response: %v", err)
	}
	if string(content) != "Server power operation successful" {
		t.Errorf("ServeHTTP() unexpected content: %s", string(content))
	}
}
