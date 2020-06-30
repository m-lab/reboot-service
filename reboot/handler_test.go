package reboot

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/m-lab/reboot-service/creds"
	"github.com/m-lab/reboot-service/creds/credstest"

	"github.com/m-lab/reboot-service/connector"
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

func (connection *mockConnection) ExecDRACShell(string) (string, error) {
	return "Not implemented", nil
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

func TestServeHTTP(t *testing.T) {
	type fields struct {
		status int
		body   string
	}

	tests := []struct {
		req                *http.Request
		status             int
		body               string
		connectorMustFail  bool
		connectionMustFail bool
	}{
		{
			req: httptest.NewRequest("POST",
				"/v1/reboot?host=mlab1d-abc0t.mlab-sandbox.measurement-lab.org", nil),
			status: http.StatusOK,
			body:   "Server power operation successful",
		},
		{
			req:    httptest.NewRequest("POST", "/v1/reboot?host=mlab2d.abc0t.measurement-lab.org&method=host", nil),
			status: http.StatusOK,
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
			req: httptest.NewRequest("POST",
				"/v1/reboot?host=mlab1d-abc1t.mlab-sandbox.measurement-lab.org", nil),
			status: http.StatusInternalServerError,
			body:   "",
		},
		{
			req: httptest.NewRequest("POST",
				"/v1/reboot?host=mlab1d-abc0t.mlab-sandbox.measurement-lab.org", nil),
			connectorMustFail: true,
			status:            http.StatusInternalServerError,
			body:              "",
		},
		{
			req: httptest.NewRequest("POST",
				"/v1/reboot?host=mlab1d-abc0t.mlab-sandbox.measurement-lab.org", nil),
			connectionMustFail: true,
			status:             http.StatusInternalServerError,
			body:               "",
		},
		{
			req: httptest.NewRequest("POST",
				"/v1/reboot?host=mlab1d-abc0t.mlab-sandbox.measurement-lab.org&method=host", nil),
			connectorMustFail: true,
			status:            http.StatusInternalServerError,
			body:              "",
		},
		{
			req: httptest.NewRequest("POST",
				"/v1/reboot?host=mlab1d-abc0t.mlab-sandbox.measurement-lab.org&method=host", nil),
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

	// Create a FakeProvider and populate it with fake Credentials.
	provider := credstest.NewProvider()
	provider.AddCredentials(context.Background(),
		"mlab1d-abc0t.mlab-sandbox.measurement-lab.org", &creds.Credentials{
			Hostname: "mlab1d.abc0t",
			Username: "testuser",
			Password: "testpass",
			Model:    "drac",
			Address:  "testaddr",
		})

	h := &Handler{
		config: &Config{
			ProjectID:      "test",
			PrivateKeyPath: "",
			BMCPort:        806,
			SSHPort:        22,
			Namespace:      "test",
		},
		connector:     connector,
		credsProvider: provider,
	}

	for _, test := range tests {
		rr := httptest.NewRecorder()

		connector.mustFail = test.connectorMustFail
		connector.connMustFail = test.connectionMustFail

		h.ServeHTTP(rr, test.req)

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
	NewHandler(&Config{}, credstest.NewProvider(), &mockConnector{})
}
