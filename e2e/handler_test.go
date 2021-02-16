// Package e2e contains the handler to perform an end-to-end connectivity test
// on a given BMC module on the M-Lab infrastructure.
package e2e

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/m-lab/reboot-service/creds"
	"github.com/m-lab/reboot-service/creds/credstest"
)

func TestNewHandler(t *testing.T) {
	handler := NewHandler(806, credstest.NewProvider(), &mockConnector{})
	if handler == nil {
		t.Errorf("NewHandler() returned nil.")
	}
}

func TestHandler_ServeHTTP(t *testing.T) {
	expMetadata := `# HELP reboot_e2e_result E2E test result for this target
# TYPE reboot_e2e_result gauge
`
	tests := []struct {
		req               *http.Request
		status            int
		body              string
		connectorMustFail bool
	}{
		{
			req:    httptest.NewRequest("GET", "/v1/e2e?target=mlab1d-abc0t.mlab-sandbox.measurement-lab.org", nil),
			status: http.StatusOK,
			body: expMetadata + `reboot_e2e_result{reason="` + reasonSuccess + `",status="` + statusOK +
				`",target="mlab1d-abc0t.mlab-sandbox.measurement-lab.org"} 1
`,
		},
		{
			req:    httptest.NewRequest("GET", "/v1/e2e?target=mlab2d.abc0t.measurement-lab.org", nil),
			status: http.StatusOK,
			body: expMetadata + `reboot_e2e_result{reason="` + reasonCredsNotFound +
				`",status="` + statusOK + `",target="mlab2d.abc0t.measurement-lab.org"} 0
`,
		},
		{
			req:               httptest.NewRequest("GET", "/v1/e2e?target=mlab1d-abc0t.mlab-sandbox.measurement-lab.org", nil),
			status:            http.StatusOK,
			connectorMustFail: true,
			body: expMetadata + `reboot_e2e_result{reason="` + reasonConnectionFailed +
				`",status="` + statusOK + `",target="mlab1d-abc0t.mlab-sandbox.measurement-lab.org"} 0
`,
		},
		{
			req:    httptest.NewRequest("POST", "/v1/e2e?target=mlab1d.abc0t.measurement-lab.org", nil),
			status: http.StatusMethodNotAllowed,
		},
		{
			req:    httptest.NewRequest("GET", "/v1/e2e", nil),
			status: http.StatusBadRequest,
		},
		{
			req:    httptest.NewRequest("GET", "/v1/e2e?target=thisshouldfail", nil),
			status: http.StatusBadRequest,
		},
	}

	connector := &mockConnector{}

	// Create a FakeProvider and populate it with fake Credentials.
	provider := credstest.NewProvider()
	provider.AddCredentials(context.Background(),
		"mlab1d-abc0t.mlab-sandbox.measurement-lab.org", &creds.Credentials{
			Hostname: "mlab1d-abc0t.mlab-sandbox.measurement-lab.org",
			Username: "testuser",
			Password: "testpass",
			Model:    "drac",
			Address:  "testaddr",
		})

	h := &Handler{
		bmcPort:   806,
		connector: connector,
		provider:  provider,
	}

	for _, test := range tests {
		rr := httptest.NewRecorder()

		connector.mustFail = test.connectorMustFail

		h.ServeHTTP(rr, test.req)

		connector.mustFail = false

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
				t.Errorf("ServeHTTP() - expected response:\n%s\ngot:\n%s\n", string(body), test.body)
			}
		}
	}

}
