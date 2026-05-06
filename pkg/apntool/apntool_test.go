package apntool

import (
	"strings"
	"testing"

	"github.com/GlshchnkLx/go-aospapn/pkg/apnxml"
)

func intPtr(value int) *int {
	return &value
}

func stringPtr(value string) *string {
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}

func baseTypePtr(value apnxml.ObjectBaseType) *apnxml.ObjectBaseType {
	return &value
}

func protocolPtr(value apnxml.ObjectBearerProtocol) *apnxml.ObjectBearerProtocol {
	return &value
}

func testData() apnxml.Array {
	return apnxml.Array{
		{
			ObjectRoot: &apnxml.ObjectRoot{Carrier: "Carrier A", Mcc: intPtr(250), Mnc: intPtr(1)},
			GroupMapByType: map[apnxml.ObjectBaseType]*apnxml.Object{
				apnxml.ObjectBaseTypeDefault: {
					Base: &apnxml.ObjectBase{
						Apn:  stringPtr("internet"),
						Type: baseTypePtr(apnxml.ObjectBaseTypeDefault),
					},
				},
				apnxml.ObjectBaseTypeMMS: {
					Base: &apnxml.ObjectBase{
						Apn:  stringPtr("mms"),
						Type: baseTypePtr(apnxml.ObjectBaseTypeMMS),
					},
				},
			},
		},
		{
			ObjectRoot: &apnxml.ObjectRoot{Carrier: "Carrier B", Mcc: intPtr(251), Mnc: intPtr(2)},
			GroupMapByType: map[apnxml.ObjectBaseType]*apnxml.Object{
				apnxml.ObjectBaseTypeIMS: {
					Base: &apnxml.ObjectBase{
						Apn:  stringPtr("ims"),
						Type: baseTypePtr(apnxml.ObjectBaseTypeIMS),
					},
				},
			},
		},
	}
}

func TestArrayFilterFiltersGroupedEntriesWithoutMutatingInput(t *testing.T) {
	data := testData()

	result := From(data).Filter(And(ByPLMN(250, 1), ByType(apnxml.ObjectBaseTypeDefault)))
	if result.Len() != 1 {
		t.Fatalf("expected one group, got %d", result.Len())
	}
	if got := result.CountRecords(); got != 1 {
		t.Fatalf("expected one record, got %d", got)
	}
	if got := data.CountRecords(); got != 3 {
		t.Fatalf("input must not be mutated, got %d records", got)
	}
}

func TestArrayFlattenAndGroupByPLMN(t *testing.T) {
	data := testData()

	flat := From(data).Flatten()
	if flat.Len() != 3 {
		t.Fatalf("expected three flat records, got %d", flat.Len())
	}
	for _, record := range flat.Data() {
		if record.ObjectRoot == nil {
			t.Fatal("flat record must include root fields")
		}
	}

	grouped := flat.GroupByPLMN()
	if grouped.Len() != 2 {
		t.Fatalf("expected two groups, got %d", grouped.Len())
	}
	if got := grouped.CountRecords(); got != 3 {
		t.Fatalf("expected three grouped records, got %d", got)
	}
}

func TestArrayGroupByPLMNAndIdentityUseDifferentKeys(t *testing.T) {
	apnType := apnxml.ObjectBaseTypeDefault
	flat := apnxml.Array{
		{
			ObjectRoot: &apnxml.ObjectRoot{
				Carrier:   "Carrier A",
				CarrierID: intPtr(10),
				Mcc:       intPtr(250),
				Mnc:       intPtr(1),
			},
			Base: &apnxml.ObjectBase{Apn: stringPtr("a"), Type: &apnType},
		},
		{
			ObjectRoot: &apnxml.ObjectRoot{
				Carrier:   "Carrier B",
				CarrierID: intPtr(20),
				Mcc:       intPtr(250),
				Mnc:       intPtr(1),
			},
			Base: &apnxml.ObjectBase{Apn: stringPtr("b"), Type: &apnType},
		},
	}

	byPLMN := From(flat).GroupByPLMN()
	if byPLMN.Len() != 1 {
		t.Fatalf("expected one PLMN group, got %d", byPLMN.Len())
	}

	byIdentity := From(flat).GroupByIdentity()
	if byIdentity.Len() != 2 {
		t.Fatalf("expected two identity groups, got %d", byIdentity.Len())
	}
}

func TestSetPolicy(t *testing.T) {
	var text *string
	if !Set(&text, "value", SetIfEmpty) || text == nil || *text != "value" {
		t.Fatal("SetIfEmpty must write nil string")
	}
	if Set(&text, "changed", SetIfEmpty) {
		t.Fatal("SetIfEmpty must not overwrite non-empty string")
	}
	if !Set(&text, "changed", SetIfExists) || *text != "changed" {
		t.Fatal("SetIfExists must overwrite existing string")
	}

	var enabled *bool
	if !Set(&enabled, false, SetAlways) || enabled == nil || *enabled != false {
		t.Fatal("SetAlways must write bool")
	}

	var protocol *apnxml.ObjectBearerProtocol
	if !Set(&protocol, apnxml.ObjectBearerProtocolIPv4v6, SetAlways) {
		t.Fatal("Set must write protocol")
	}
	if *protocol != apnxml.ObjectBearerProtocolIPv4v6 {
		t.Fatal("unexpected protocol value")
	}
}

func TestArrayApplyTransformsCloneAndPropagatesErrors(t *testing.T) {
	data := testData()
	protocol := apnxml.ObjectBearerProtocolIPv4v6

	result, err := From(data).
		Filter(ByPLMN(250, 1)).
		Apply(func(record *apnxml.Object) error {
			bearer := EnsureBearer(record)
			Set(&bearer.Type, protocol, SetAlways)
			return nil
		})
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	if result.CountRecords() != 2 {
		t.Fatalf("expected two transformed records, got %d", result.CountRecords())
	}
	for _, record := range result.Data()[0].Records() {
		if record.Bearer == nil || record.Bearer.Type == nil || *record.Bearer.Type != protocol {
			t.Fatal("record was not transformed")
		}
	}

	if data[0].GroupMapByType[apnxml.ObjectBaseTypeDefault].Bearer != nil {
		t.Fatal("Apply must not mutate input")
	}
}

func TestMaterializeRecordPreservesExistingRoot(t *testing.T) {
	record := &apnxml.Object{
		ObjectRoot: &apnxml.ObjectRoot{Carrier: "Record", Mcc: intPtr(1), Mnc: intPtr(1)},
	}
	group := &apnxml.Object{
		ObjectRoot: &apnxml.ObjectRoot{Carrier: "Group", Mcc: intPtr(2), Mnc: intPtr(2)},
	}

	materialized := MaterializeRecord(group, record)
	if materialized.Carrier != "Record" {
		t.Fatalf("expected record root to win, got %q", materialized.Carrier)
	}
}

func TestEnsureHelpersHandleNilRecord(t *testing.T) {
	if EnsureRoot(nil) != nil ||
		EnsureBase(nil) != nil ||
		EnsureAuth(nil) != nil ||
		EnsureBearer(nil) != nil ||
		EnsureProxy(nil) != nil ||
		EnsureMMS(nil) != nil ||
		EnsureMVNO(nil) != nil ||
		EnsureLimit(nil) != nil ||
		EnsureOther(nil) != nil {
		t.Fatal("ensure helpers must be nil-safe")
	}

	var record apnxml.Object
	if EnsureRoot(&record) == nil ||
		EnsureBase(&record) == nil ||
		EnsureAuth(&record) == nil ||
		EnsureBearer(&record) == nil ||
		EnsureProxy(&record) == nil ||
		EnsureMMS(&record) == nil ||
		EnsureMVNO(&record) == nil ||
		EnsureLimit(&record) == nil ||
		EnsureOther(&record) == nil {
		t.Fatal("ensure helpers must allocate missing sections")
	}

	_ = boolPtr
	_ = protocolPtr
}

func TestArrayAPIIsCloneSafeAtBoundaries(t *testing.T) {
	data := testData()
	tool := From(data)

	exported := tool.Data()
	exported[0].GroupMapByType[apnxml.ObjectBaseTypeDefault].Base.Apn = stringPtr("changed")

	record, ok := tool.First(ByType(apnxml.ObjectBaseTypeDefault))
	if !ok {
		t.Fatal("expected default record")
	}
	if record.Base == nil || record.Base.Apn == nil || *record.Base.Apn != "internet" {
		t.Fatal("Data must return a clone")
	}

	data[0].GroupMapByType[apnxml.ObjectBaseTypeDefault].Base.Apn = stringPtr("source changed")
	record, ok = tool.First(ByType(apnxml.ObjectBaseTypeDefault))
	if !ok {
		t.Fatal("expected default record")
	}
	if record.Base == nil || record.Base.Apn == nil || *record.Base.Apn != "internet" {
		t.Fatal("From must clone input by default")
	}
}

func TestArrayMethodsDoNotMutateSource(t *testing.T) {
	data := testData()

	result, err := From(data).
		Filter(ByPLMN(250, 1)).
		Apply(func(record *apnxml.Object) error {
			base := EnsureBase(record)
			Set(&base.ProfileID, 77, SetAlways)
			return nil
		})
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	if result.CountRecords() != 2 {
		t.Fatalf("expected two records, got %d", result.CountRecords())
	}

	resultRecord, ok := result.First(ByType(apnxml.ObjectBaseTypeDefault))
	if !ok || resultRecord.Base == nil || resultRecord.Base.ProfileID == nil || *resultRecord.Base.ProfileID != 77 {
		t.Fatal("result record was not mutated")
	}

	sourceRecord := data[0].GroupMapByType[apnxml.ObjectBaseTypeDefault]
	if sourceRecord.Base.ProfileID != nil {
		t.Fatal("source data must not be mutated")
	}
}

func TestArrayMapReturnsFlatMappedRecords(t *testing.T) {
	mapped, err := From(testData()).Map(func(record apnxml.Object) (apnxml.Object, error) {
		if record.Base != nil && record.Base.Apn != nil {
			*record.Base.Apn = strings.ToUpper(*record.Base.Apn)
		}
		return record, nil
	})
	if err != nil {
		t.Fatalf("Map returned error: %v", err)
	}

	if mapped.Len() != 3 || mapped.CountRecords() != 3 {
		t.Fatalf("Map must return flat records, got len=%d count=%d", mapped.Len(), mapped.CountRecords())
	}

	record, ok := mapped.First(ByAPN("INTERNET"))
	if !ok || record.ObjectRoot == nil {
		t.Fatal("mapped materialized record not found")
	}
}

func TestPredicatesCoverRootAndSections(t *testing.T) {
	protocol := apnxml.ObjectBearerProtocolIPv4v6
	network := apnxml.ObjectNetworkTypeLTE
	record := apnxml.Object{
		ObjectRoot: &apnxml.ObjectRoot{
			Carrier:   "Carrier LTE",
			CarrierID: intPtr(42),
			Mcc:       intPtr(250),
			Mnc:       intPtr(1),
		},
		Base: &apnxml.ObjectBase{
			Apn:  stringPtr("internet"),
			Type: baseTypePtr(apnxml.ObjectBaseTypeDefault),
		},
		Bearer: &apnxml.ObjectBearer{
			Type: &protocol,
		},
		Mms: &apnxml.ObjectMMS{
			Center: stringPtr("http://mms.example"),
		},
		Other: &apnxml.ObjectOther{
			NetworkTypeBitmask: &network,
		},
	}

	predicate := And(
		ByMCC(250),
		ByMNC(1),
		ByCarrierID(42),
		ByAPNContains("net"),
		ByProtocol(protocol),
		ByNetwork(network),
		HasMMS,
	)
	if !predicate(record) {
		t.Fatal("expected predicate chain to match")
	}
	if ByCarrierID(7)(record) || ByAPNContains("ims")(record) {
		t.Fatal("unexpected predicate match")
	}
}

func TestArrayStatsAndIndexes(t *testing.T) {
	tool := From(testData())
	stats := tool.Stats()
	if stats.Groups != 2 || stats.Records != 3 || stats.Invalid != 0 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
	if stats.ByPLMN["25001"] != 2 || stats.ByPLMN["25102"] != 1 {
		t.Fatalf("unexpected PLMN stats: %+v", stats.ByPLMN)
	}
	if stats.ByType[apnxml.ObjectBaseTypeDefault] != 1 ||
		stats.ByType[apnxml.ObjectBaseTypeMMS] != 1 ||
		stats.ByType[apnxml.ObjectBaseTypeIMS] != 1 {
		t.Fatalf("unexpected type stats: %+v", stats.ByType)
	}

	plmns := tool.PLMNs()
	if len(plmns) != 2 || plmns[0] != "25001" || plmns[1] != "25102" {
		t.Fatalf("unexpected PLMNs: %#v", plmns)
	}

	types := tool.Types()
	if len(types) != 3 ||
		types[0] != apnxml.ObjectBaseTypeDefault ||
		types[1] != apnxml.ObjectBaseTypeMMS ||
		types[2] != apnxml.ObjectBaseTypeIMS {
		t.Fatalf("unexpected types: %#v", types)
	}
}

func TestArrayDedupeAndMergePatch(t *testing.T) {
	apnType := apnxml.ObjectBaseTypeDefault
	source := apnxml.Array{
		{
			ObjectRoot: &apnxml.ObjectRoot{Carrier: "Carrier A", Mcc: intPtr(250), Mnc: intPtr(1)},
			Base:       &apnxml.ObjectBase{Apn: stringPtr("internet"), Type: &apnType},
		},
		{
			ObjectRoot: &apnxml.ObjectRoot{Carrier: "Carrier A duplicate", Mcc: intPtr(250), Mnc: intPtr(1)},
			Base:       &apnxml.ObjectBase{Apn: stringPtr("changed"), Type: &apnType, ProfileID: intPtr(10)},
		},
	}

	deduped := From(source).DedupeByIdentity()
	if deduped.Len() != 1 || deduped.CountRecords() != 1 {
		t.Fatalf("unexpected dedupe result len=%d count=%d", deduped.Len(), deduped.CountRecords())
	}
	record, ok := deduped.First(ByType(apnType))
	if !ok || record.Base == nil || record.Base.Apn == nil || *record.Base.Apn != "internet" {
		t.Fatal("dedupe must keep first record for duplicate APN type")
	}

	merged := deduped.Merge(apnxml.Array{source[1]})
	record, ok = merged.First(ByType(apnType))
	if !ok || record.Base == nil || record.Base.ProfileID == nil || *record.Base.ProfileID != 10 {
		t.Fatal("merge must fill missing fields from source")
	}
	if record.Base.Apn == nil || *record.Base.Apn != "internet" {
		t.Fatal("merge must not overwrite existing fields")
	}

	patched := deduped.Patch(apnxml.Array{source[1]})
	record, ok = patched.First(ByType(apnType))
	if !ok || record.Base == nil || record.Base.Apn == nil || *record.Base.Apn != "changed" {
		t.Fatal("patch must overwrite non-zero source fields")
	}
}
