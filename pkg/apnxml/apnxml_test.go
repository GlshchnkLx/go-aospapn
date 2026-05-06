package apnxml

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

//--------------------------------------------------------------------------------//
// Helper
//--------------------------------------------------------------------------------//

func intPtr(value int) *int {
	return &value
}

func stringPtr(value string) *string {
	return &value
}

//--------------------------------------------------------------------------------//
// Test
//--------------------------------------------------------------------------------//

func TestImportFromXMLGroupsEntriesByPLMNAndKeepsFirstTypeEntry(t *testing.T) {
	apns, err := ImportFromXMLByte([]byte(`<apns version="8">
		<apn carrier="Carrier Internet" mcc="250" mnc="01" apn="internet" type="default" protocol="IPV4V6" />
		<apn carrier="Carrier MMS" mcc="250" mnc="01" apn="mms" type="mms" mmsc="http://mmsc" />
		<apn carrier="Carrier Backup" mcc="250" mnc="01" apn="backup" type="default" profile_id="2" />
	</apns>`))
	if err != nil {
		t.Fatalf("ImportFromXMLByte returned error: %v", err)
	}

	if len(apns) != 1 {
		t.Fatalf("expected one grouped PLMN entry, got %d", len(apns))
	}

	if got := apns[0].GetPLMN(); got != "25001" {
		t.Fatalf("expected PLMN 25001, got %q", got)
	}

	if len(apns[0].GroupMapByType) != 2 {
		t.Fatalf("expected map keyed by two APN types, got %d", len(apns[0].GroupMapByType))
	}

	defaultRecord := apns[0].GroupMapByType[ObjectBaseTypeDefault]
	if defaultRecord == nil || defaultRecord.Base == nil || defaultRecord.Base.Apn == nil {
		t.Fatal("expected default APN record")
	}
	if *defaultRecord.Base.Apn != "internet" {
		t.Fatalf("expected first default APN to be kept, got %q", *defaultRecord.Base.Apn)
	}
}

func TestArrayCloneAndCountRecords(t *testing.T) {
	apns := Array{
		{
			ObjectRoot: &ObjectRoot{Carrier: "A", Mcc: intPtr(250), Mnc: intPtr(1)},
			GroupMapByType: map[ObjectBaseType]*Object{
				ObjectBaseTypeDefault: {Base: &ObjectBase{Apn: stringPtr("internet")}},
				ObjectBaseTypeMMS:     {Base: &ObjectBase{Apn: stringPtr("mms")}},
			},
		},
		{ObjectRoot: &ObjectRoot{Carrier: "B", Mcc: intPtr(251), Mnc: intPtr(2)}},
	}

	if got := apns.CountRecords(); got != 3 {
		t.Fatalf("expected three records, got %d", got)
	}

	clone := apns.Clone()
	*clone[0].GroupMapByType[ObjectBaseTypeDefault].Base.Apn = "changed"
	if *apns[0].GroupMapByType[ObjectBaseTypeDefault].Base.Apn != "internet" {
		t.Fatal("Clone must deep-copy grouped records")
	}
}

func TestObjectRecordsAreSortedByType(t *testing.T) {
	group := Object{
		GroupMapByType: map[ObjectBaseType]*Object{
			ObjectBaseTypeMMS:     {Base: &ObjectBase{Apn: stringPtr("mms")}},
			ObjectBaseTypeDefault: {Base: &ObjectBase{Apn: stringPtr("internet")}},
		},
	}

	records := group.Records()
	if len(records) != 2 {
		t.Fatalf("expected two records, got %d", len(records))
	}
	if *records[0].Base.Apn != "internet" {
		t.Fatalf("expected default record first, got %q", *records[0].Base.Apn)
	}
}

func TestGroupedMatchChecksRootBeforeGroupEntries(t *testing.T) {
	apns, err := ImportFromXMLByte([]byte(`<apns version="8">
		<apn carrier="Carrier A" mcc="250" mnc="01" apn="internet" type="default" />
	</apns>`))
	if err != nil {
		t.Fatalf("ImportFromXMLByte returned error: %v", err)
	}

	query := &Object{ObjectRoot: &ObjectRoot{Mcc: intPtr(251), Mnc: intPtr(1)}}
	if apns[0].Match(query) {
		t.Fatal("grouped APN must not match a query with different PLMN")
	}
}

func TestCloneAndMarshalAreNilSafe(t *testing.T) {
	var apn *Object
	if apn.Clone() != nil {
		t.Fatal("nil APN clone must be nil")
	}

	object := Object{Base: &ObjectBase{Apn: stringPtr("internet")}}
	if object.Clone() == nil {
		t.Fatal("partial APN clone must preserve object")
	}

	if _, err := xml.Marshal(object); err != nil {
		t.Fatalf("xml marshal of partial APN returned error: %v", err)
	}
}

func TestValidateDoesNotNormalizeBase(t *testing.T) {
	base := &ObjectBase{Apn: stringPtr("internet")}

	if !base.Validate() {
		t.Fatal("base with APN must be valid")
	}
	if base.Type != nil {
		t.Fatal("Validate must not mutate Base.Type")
	}

	base.Normalize()
	if base.Type == nil || *base.Type != ObjectBaseTypeDefault {
		t.Fatal("Normalize must assign default APN type when APN is set")
	}
}

func TestObjectFieldCloneDoesNotAliasPointers(t *testing.T) {
	baseType := ObjectBaseTypeDefault
	base := &ObjectBase{
		Apn:  stringPtr("internet"),
		Type: &baseType,
	}

	clone := base.Clone()
	*clone.Apn = "changed"
	*clone.Type = ObjectBaseTypeMMS

	if *base.Apn != "internet" {
		t.Fatal("Clone must not alias string pointers")
	}
	if *base.Type != ObjectBaseTypeDefault {
		t.Fatal("Clone must not alias enum pointers")
	}
}

func TestObjectUpdateModes(t *testing.T) {
	target := &ObjectBase{
		Apn: stringPtr("internet"),
	}
	source := &ObjectBase{
		Apn:       stringPtr("mms"),
		ProfileID: intPtr(7),
	}

	if !target.Merge(source) {
		t.Fatal("Merge returned false")
	}
	if *target.Apn != "internet" {
		t.Fatal("Merge must not overwrite existing fields")
	}
	if target.ProfileID == nil || *target.ProfileID != 7 {
		t.Fatal("Merge must fill empty fields")
	}

	if !target.Patch(source) {
		t.Fatal("Patch returned false")
	}
	if *target.Apn != "mms" {
		t.Fatal("Patch must overwrite fields present in source")
	}

	if !target.Apply(&ObjectBase{}) {
		t.Fatal("Apply returned false")
	}
	if target.Apn != nil || target.ProfileID != nil {
		t.Fatal("Apply must copy zero values")
	}
}

func TestImportExportReaderWriter(t *testing.T) {
	input := strings.NewReader(`<apns version="8"><apn carrier="A" mcc="250" mnc="01" apn="internet" type="default" /></apns>`)

	apns, err := ImportFromReader(input, FormatXML)
	if err != nil {
		t.Fatalf("ImportFromReader returned error: %v", err)
	}

	var builder strings.Builder
	if err := ExportToWriter(apns, &builder, FormatJSON); err != nil {
		t.Fatalf("ExportToWriter returned error: %v", err)
	}

	var decoded Array
	if err := json.Unmarshal([]byte(builder.String()), &decoded); err != nil {
		t.Fatalf("exported JSON is invalid: %v", err)
	}
}

func TestImportFromURLUsesContextAndClient(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		_, _ = response.Write([]byte(`<apns version="8"><apn carrier="A" mcc="250" mnc="01" /></apns>`))
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	apns, err := ImportFromURL(ctx, server.Client(), server.URL, FormatXML, false)
	if err != nil {
		t.Fatalf("ImportFromURL returned error: %v", err)
	}
	if len(apns) != 1 {
		t.Fatalf("expected one APN entry, got %d", len(apns))
	}
}

func TestImportFromSimpleURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		_, _ = response.Write([]byte(`<apns version="8"><apn carrier="A" mcc="250" mnc="01" /></apns>`))
	}))
	defer server.Close()

	apns, err := ImportFromSimpleURL(server.URL, false)
	if err != nil {
		t.Fatalf("ImportFromSimpleURL returned error: %v", err)
	}
	if len(apns) != 1 {
		t.Fatalf("expected one APN entry, got %d", len(apns))
	}
}

func TestEnumUnmarshalRejectsEmptyJSON(t *testing.T) {
	var baseType ObjectBaseType
	if err := baseType.UnmarshalJSON(nil); err == nil {
		t.Fatal("expected empty JSON to be rejected")
	}
}

func TestParseHelpers(t *testing.T) {
	format, err := ParseFormat("XML")
	if err != nil || format != FormatXML {
		t.Fatalf("unexpected format parse result: %q %v", format, err)
	}
	mode, err := ParseObjectUpdateMode("merge")
	if err != nil || mode != ObjectUpdateMerge {
		t.Fatalf("unexpected update mode parse result: %v %v", mode, err)
	}
	baseType, err := ParseObjectBaseType("default,mms")
	if err != nil {
		t.Fatalf("ParseObjectBaseType returned error: %v", err)
	}
	if baseType&ObjectBaseTypeDefault == 0 || baseType&ObjectBaseTypeMMS == 0 {
		t.Fatalf("unexpected base type mask: %s", baseType.String())
	}
	protocol, err := ParseObjectBearerProtocol("ipv4v6")
	if err != nil || protocol != ObjectBearerProtocolIPv4v6 {
		t.Fatalf("unexpected protocol parse result: %s %v", protocol.String(), err)
	}
	network, err := ParseObjectNetworkType("lte,nr")
	if err != nil || network&ObjectNetworkTypeLTE == 0 || network&ObjectNetworkTypeNR == 0 {
		t.Fatalf("unexpected network parse result: %s %v", network.String(), err)
	}
}

//--------------------------------------------------------------------------------//
