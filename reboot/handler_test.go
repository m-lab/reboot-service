package reboot

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/m-lab/reboot-service/connector"
	"github.com/m-lab/reboot-service/creds"
)

// Mock structs for Connector and Connection interfaces.
type mockConnector struct {
	mustFail     bool
	connMustFail bool
}

type mockConnection struct {
	mustFail bool
}

func (connector *mockConnector) NewConnection(*connector.ConnectionConfig) (connector.Connection, error) {
	if connector.mustFail {
		return nil, errors.New("method NewConnection() failed")
	}
	return &mockConnection{
		mustFail: connector.connMustFail,
	}, nil
}

func (connection *mockConnection) Reboot() (string, error) {
	if connection.mustFail {
		return "", errors.New("method Reboot() failed")
	}
	return "Server power operation successful", nil
}

func (connection *mockConnection) Close() error {
	return nil
}

// Mock struct for credentials Provider
type mockProvider struct {
	mustFail bool
}

func (p *mockProvider) FindCredentials(context.Context, string) (*creds.Credentials, error) {
	if p.mustFail {
		return nil, errors.New("method FindCredentials() failed")
	}
	return &creds.Credentials{
		Hostname: "testhost",
		Username: "testuser",
		Password: "testpass",
		Model:    "drac",
		Address:  "testaddr",
	}, nil
}

func TestServeHTTP(t *testing.T) {
	type fields struct {
		status int
		body   string
	}

	tests := []struct {
		req                *http.Request
		status             int
		body               string
		credsMustFail      bool
		connectorMustFail  bool
		connectionMustFail bool
	}{
		{
			req:    httptest.NewRequest("POST", "/v1/reboot?host=mlab1.lga0t", nil),
			status: http.StatusOK,
			body:   "Server power operation successful",
		},
		{
			req:    httptest.NewRequest("GET", "/v1/reboot?host=mlab1.lga0t", nil),
			status: http.StatusMethodNotAllowed,
			body:   "",
		},
		{
			req:    httptest.NewRequest("POST", "/v1/reboot", nil),
			status: http.StatusBadRequest,
			body:   "",
		},
		{
			req:           httptest.NewRequest("POST", "/v1/reboot?host=mlab1.lga0t", nil),
			credsMustFail: true,
			status:        http.StatusInternalServerError,
			body:          "",
		},
		{
			req:               httptest.NewRequest("POST", "/v1/reboot?host=mlab1.lga0t", nil),
			connectorMustFail: true,
			status:            http.StatusInternalServerError,
			body:              "",
		},
		{
			req:                httptest.NewRequest("POST", "/v1/reboot?host=mlab1.lga0t", nil),
			connectionMustFail: true,
			status:             http.StatusInternalServerError,
			body:               "",
		},
		{
			req:    httptest.NewRequest("POST", "/v1/reboot?host=thisshouldfail", nil),
			status: http.StatusBadRequest,
			body:   "The specified hostname is not a valid M-Lab node: thisshouldfail",
		},
	}

	connector := &mockConnector{}
	creds := &mockProvider{}
	h := &Handler{
		config: &Config{
			ProjectID:      "test",
			PrivateKeyPath: "",
			DRACPort:       806,
			SSHPort:        22,
			Namespace:      "test",
		},
		connector:     connector,
		credsProvider: creds,
	}

	for _, test := range tests {
		rr := httptest.NewRecorder()

		creds.mustFail = test.credsMustFail
		connector.mustFail = test.connectorMustFail
		connector.connMustFail = test.connectionMustFail

		h.ServeHTTP(rr, test.req)

		creds.mustFail = false
		connector.mustFail = false
		connector.connMustFail = false

		resp := rr.Result()

		// Test StatusCode and Body against the expected values.
		if resp.StatusCode != test.status {
			t.Errorf("ServeHTTP - expected %d, got %d", test.status,
				resp.StatusCode)
		}

		if test.body != "" {
			body, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				t.Errorf("ServeHTTP() - cannot read response: %v", err)
			}
			if string(body) != test.body {
				t.Errorf("ServeHTTP() - unexpected response: %s", string(body))
			}
		}
	}

}

func TestNewHandler(t *testing.T) {
	NewHandler(&Config{}, &mockProvider{}, &mockConnector{})
}
