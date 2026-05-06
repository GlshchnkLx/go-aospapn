package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const apnctlFixtureXML = `<apns version="8">
	<apn carrier="Carrier A" carrier_id="10" mcc="250" mnc="01" apn="internet" type="default" protocol="IPV4V6" roaming_protocol="IPV4V6" bearer_bitmask="lte|nr" carrier_enabled="true" user_visible="true" user_editable="false" />
	<apn carrier="Carrier A MMS" carrier_id="10" mcc="250" mnc="01" apn="mms" type="mms" mmsc="http://mms.example" />
	<apn carrier="Carrier B" carrier_id="20" mcc="251" mnc="02" apn="ims" type="ims" protocol="IPV6" />
	<apn carrier="Broken" mcc="999" apn="broken" type="default" />
</apns>`

func TestAPNCtlCapabilities(t *testing.T) {
	fixture := newAPNCtlFixture(t)

	tests := []struct {
		name        string
		args        func(t *testing.T, fixture apnctlFixture) []string
		wantOut     []string
		wantErr     string
		validateOut func(t *testing.T, out string)
	}{
		{
			name: "convert xml to json with pagination",
			args: func(t *testing.T, fixture apnctlFixture) []string {
				return []string{
					"convert",
					"--in", fixture.inputXML,
					"--flat",
					"--limit", "1",
					"--output-format", "json",
					"--out", fixture.out(t),
				}
			},
			wantOut: []string{`"carrierName": "Carrier A"`, `"apn": "internet"`},
		},
		{
			name: "find renders filtered table",
			args: func(t *testing.T, fixture apnctlFixture) []string {
				return []string{
					"find",
					"--in", fixture.inputXML,
					"--type", "mms",
					"--output-format", "table",
					"--out", fixture.out(t),
				}
			},
			wantOut: []string{"PLMN\tCarrier\tCarrierID\tType\tAPN", "25001\tCarrier A\t10\tmms\tmms"},
		},
		{
			name: "list returns distinct APNs as sorted text",
			args: func(t *testing.T, fixture apnctlFixture) []string {
				return []string{
					"list",
					"--in", fixture.inputXML,
					"--kind", "apn",
					"--out", fixture.out(t),
				}
			},
			wantOut: []string{"ims\n", "internet\n", "mms\n"},
		},
		{
			name: "stats writes summary counters",
			args: func(t *testing.T, fixture apnctlFixture) []string {
				return []string{
					"stats",
					"--in", fixture.inputXML,
					"--out", fixture.out(t),
				}
			},
			wantOut: []string{"groups: 2", "records: 3", "invalid: 0", "25001: 2", "25102: 1"},
		},
		{
			name: "inspect expands grouped records",
			args: func(t *testing.T, fixture apnctlFixture) []string {
				return []string{
					"inspect",
					"--in", fixture.inputXML,
					"--plmn", "25001",
					"--out", fixture.out(t),
				}
			},
			wantOut: []string{"PLMN: 25001", "records: 2", "type=default apn=internet", "type=mms apn=mms"},
		},
		{
			name: "patch updates records selected by filter",
			args: func(t *testing.T, fixture apnctlFixture) []string {
				return []string{
					"patch",
					"--in", fixture.inputXML,
					"--plmn", "25001",
					"--type", "default",
					"--set", "base.profileID=42",
					"--set", "other.carrierEnabled=false",
					"--output-format", "json",
					"--out", fixture.out(t),
				}
			},
			wantOut: []string{`"profileID": 42`, `"IsEnabled": false`},
		},
		{
			name: "build creates one normalized record",
			args: func(t *testing.T, fixture apnctlFixture) []string {
				return []string{
					"build",
					"--carrier", "Built Carrier",
					"--mcc", "310",
					"--mnc", "260",
					"--apn", "built",
					"--type", "default,supl",
					"--protocol", "ipv4v6",
					"--network", "lte,nr",
					"--enabled", "true",
					"--output-format", "json",
					"--out", fixture.out(t),
				}
			},
			validateOut: func(t *testing.T, out string) {
				t.Helper()
				var records []map[string]any
				if err := json.Unmarshal([]byte(out), &records); err != nil {
					t.Fatalf("build output is not valid JSON: %v\n%s", err, out)
				}
				if len(records) != 1 {
					t.Fatalf("expected one built record, got %d", len(records))
				}
				if records[0]["carrierName"] != "Built Carrier" || records[0]["mcc"].(float64) != 310 || records[0]["mnc"].(float64) != 260 {
					t.Fatalf("unexpected built root: %#v", records[0])
				}
				base := records[0]["base"].(map[string]any)
				if base["apn"] != "built" {
					t.Fatalf("unexpected built base: %#v", base)
				}
			},
		},
		{
			name: "fetch decodes base64 URL body",
			args: func(t *testing.T, fixture apnctlFixture) []string {
				server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
					_, _ = response.Write([]byte(base64.StdEncoding.EncodeToString([]byte(apnctlFixtureXML))))
				}))
				t.Cleanup(server.Close)
				return []string{
					"fetch",
					"--url", server.URL,
					"--base64",
					"--output-format", "json",
					"--out", fixture.out(t),
				}
			},
			wantOut: []string{`"carrierName": "Carrier A"`, `"carrierName": "Carrier B"`},
		},
		{
			name: "validate strict reports invalid records",
			args: func(t *testing.T, fixture apnctlFixture) []string {
				return []string{
					"validate",
					"--in", fixture.invalidJSON,
					"--input-format", "json",
					"--strict",
					"--out", fixture.out(t),
				}
			},
			wantErr: "invalid APN records: 1",
			wantOut: []string{"invalid: 1"},
		},
		{
			name: "unknown command returns error",
			args: func(t *testing.T, fixture apnctlFixture) []string {
				return []string{"unknown"}
			},
			wantErr: `unknown command "unknown"`,
		},
		{
			name: "mutually exclusive filters return error",
			args: func(t *testing.T, fixture apnctlFixture) []string {
				return []string{
					"find",
					"--in", fixture.inputXML,
					"--valid-only",
					"--invalid-only",
					"--out", fixture.out(t),
				}
			},
			wantErr: "--valid-only and --invalid-only are mutually exclusive",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			args := test.args(t, fixture)
			err := run(args)
			if test.wantErr == "" && err != nil {
				t.Fatalf("run(%q) returned error: %v", args, err)
			}
			if test.wantErr != "" {
				if err == nil {
					t.Fatalf("run(%q) returned nil error, want %q", args, test.wantErr)
				}
				if !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("run(%q) error = %q, want containing %q", args, err.Error(), test.wantErr)
				}
			}

			out := readAPNCtlOutput(t, args)
			for _, want := range test.wantOut {
				if !strings.Contains(out, want) {
					t.Fatalf("output does not contain %q:\n%s", want, out)
				}
			}
			if test.validateOut != nil {
				test.validateOut(t, out)
			}
		})
	}
}

type apnctlFixture struct {
	dir         string
	inputXML    string
	invalidJSON string
}

func newAPNCtlFixture(t *testing.T) apnctlFixture {
	t.Helper()
	dir := t.TempDir()
	inputXML := filepath.Join(dir, "apns.xml")
	if err := os.WriteFile(inputXML, []byte(apnctlFixtureXML), 0o600); err != nil {
		t.Fatalf("write XML fixture: %v", err)
	}
	invalidJSON := filepath.Join(dir, "invalid.json")
	if err := os.WriteFile(invalidJSON, []byte(`[{"carrierName":"Broken","mcc":999}]`), 0o600); err != nil {
		t.Fatalf("write invalid JSON fixture: %v", err)
	}
	return apnctlFixture{dir: dir, inputXML: inputXML, invalidJSON: invalidJSON}
}

func (fixture apnctlFixture) out(t *testing.T) string {
	t.Helper()
	return filepath.Join(fixture.dir, strings.ReplaceAll(t.Name(), "/", "_")+".out")
}

func readAPNCtlOutput(t *testing.T, args []string) string {
	t.Helper()
	for index, arg := range args {
		if arg == "--out" && index+1 < len(args) {
			data, err := os.ReadFile(args[index+1])
			if err != nil {
				if os.IsNotExist(err) {
					return ""
				}
				t.Fatalf("read output %q: %v", args[index+1], err)
			}
			return string(data)
		}
	}
	return ""
}
