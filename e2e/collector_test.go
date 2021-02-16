package e2e

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/m-lab/reboot-service/connector"
	"github.com/m-lab/reboot-service/creds/credstest"

	"github.com/m-lab/reboot-service/creds"
	"github.com/prometheus/client_golang/prometheus"
)

// Mock structs for Connector and Connection interfaces.
type mockConnector struct {
	mustFail bool
}

type mockConnection struct {
	mustFail bool
}

func (connector *mockConnector) NewConnection(*connector.ConnectionConfig) (connector.Connection, error) {
	if connector.mustFail {
		return nil, errors.New("method NewConnection() failed")
	}
	return &mockConnection{}, nil
}

func (connection *mockConnection) ExecDRACShell(string) (string, error) {
	return "Not implemented", nil
}

func (connection *mockConnection) Reboot() (string, error) {
	return "Not implemented", nil
}

func (connection *mockConnection) Close() error {
	return nil
}

func Test_newE2ETestCollector(t *testing.T) {
	config := &collectorConfig{
		bmcPort:   806,
		connector: &mockConnector{},
		provider:  credstest.NewProvider(),
	}

	collector := newE2ETestCollector("mlab1.abc0t.measurement-lab.org", config)
	if collector == nil {
		t.Errorf("newE2ETestCollector() returned nil.")
	}
}

func Test_e2eTestCollector_Collect(t *testing.T) {
	provider := credstest.NewProvider()
	connector := &mockConnector{}
	provider.AddCredentials(context.Background(),
		"mlab1d.abc0t.measurement-lab.org", &creds.Credentials{
			Hostname: "mlab1d.abc0t.measurement-lab.org",
			Username: "admin",
			Password: "dummy",
		})
	config := &collectorConfig{
		bmcPort:   806,
		connector: connector,
		provider:  provider,
	}
	collector := newE2ETestCollector("mlab1d.abc0t.measurement-lab.org", config)

	// Compare actual vs expected output in the "ok" case.
	expMetadata := `# HELP reboot_e2e_success E2E test result for this target
# TYPE reboot_e2e_success gauge

`
	expMetric := `
reboot_e2e_success{reason="` + reasonSuccess + `",target="mlab1d.abc0t.measurement-lab.org"} 1
`
	err := testutil.CollectAndCompare(collector, strings.NewReader(
		expMetadata+expMetric))
	if err != nil {
		t.Errorf("CollectAndCompare() returned err: %v", err)
	}

	// Compare actual vs expected output in the "connection_failed" case.
	expMetric = `
reboot_e2e_success{reason="` + reasonConnectionFailed + `",target="mlab1d.abc0t.measurement-lab.org"} 0
`
	connector.mustFail = true
	collector = newE2ETestCollector("mlab1d.abc0t.measurement-lab.org", config)
	err = testutil.CollectAndCompare(collector, strings.NewReader(
		expMetadata+expMetric))
	if err != nil {
		t.Errorf("CollectAndCompare() returned err: %v", err)
	}

	// Compare actual vs expected output in the "credentials_not_found" case.
	expMetric = `
reboot_e2e_success{reason="` + reasonCredsNotFound + `",target="mlab2d.abc0t.measurement-lab.org"} 0
`
	collector = newE2ETestCollector("mlab2d.abc0t.measurement-lab.org", config)
	err = testutil.CollectAndCompare(collector, strings.NewReader(
		expMetadata+expMetric))
	if err != nil {
		t.Errorf("CollectAndCompare() returned err: %v", err)
	}

}

func Test_e2eTestCollector_getCredentials(t *testing.T) {
	type fields struct {
		target       string
		config       *collectorConfig
		resultMetric *prometheus.Desc
	}
	type args struct {
		hostname string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *creds.Credentials
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &e2eTestCollector{
				target:       tt.fields.target,
				config:       tt.fields.config,
				resultMetric: tt.fields.resultMetric,
			}
			got, err := c.getCredentials(tt.args.hostname)
			if (err != nil) != tt.wantErr {
				t.Errorf("e2eTestCollector.getCredentials() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("e2eTestCollector.getCredentials() = %v, want %v", got, tt.want)
			}
		})
	}
}
