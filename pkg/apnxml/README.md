# pkg/apnxml

`apnxml` is the low-level package for reading, writing and representing AOSP
APN data.

It supports AOSP-style XML, package JSON, file/reader/writer helpers, URL
imports and a Go model for grouped APN records.

## Import

```go
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/GlshchnkLx/go-aospapn/pkg/apnxml"
)

func main() {
	apns, err := apnxml.ImportFromFile("apns-conf.xml")
	if err != nil {
		log.Fatal(err)
	}

	_ = apns

	apns, err = apnxml.ImportFromURL(
		context.Background(),
		http.DefaultClient,
		"https://example.com/apns-conf.xml",
		apnxml.FormatXML,
		false,
	)
	if err != nil {
		log.Fatal(err)
	}
}
```

Available import helpers:

- `ImportFromXMLByte([]byte) (Array, error)`
- `ImportFromJSONByte([]byte) (Array, error)`
- `ImportFromReader(io.Reader, Format) (Array, error)`
- `ImportFromFile(string) (Array, error)`
- `ImportFromURL(context.Context, *http.Client, string, Format, bool) (Array, error)`
- `ImportFromSimpleURL(string, bool) (Array, error)`
- `FormatFromFilename(string) (Format, error)`
- `ParseFormat(string) (Format, error)`

`ImportFromFile` detects the format from the filename extension. Supported
extensions are `.xml` and `.json`.

`ImportFromURL` falls back to `context.Background()` and `http.DefaultClient`
when the context or client argument is nil.

`ImportFromSimpleURL` is a compatibility helper for XML URLs. It uses
`context.Background()`, `http.DefaultClient` and `FormatXML`.

The URL helpers can decode a base64 response body when `isBase64` is `true`.

## Export

```go
data, err := apnxml.ExportToXMLByte(apns)
if err != nil {
	log.Fatal(err)
}

err = apnxml.ExportToFile(apns, "apns-conf.json")
if err != nil {
	log.Fatal(err)
}

_ = data
```

Available export helpers:

- `ExportToXMLByte(Array) ([]byte, error)`
- `ExportToJSONByte(Array) ([]byte, error)`
- `ExportToWriter(Array, io.Writer, Format) error`
- `ExportToFile(Array, string) error`

`ExportToFile` also detects the output format from `.xml` or `.json`.

XML export writes an `<apns version="8">` root element. Grouped objects are
expanded back to one `<apn>` element per APN type.

## Data Model

`Array` is a slice of `Object`.

`Object` contains the root APN identity and optional sections:

- `ObjectRoot`: carrier name, carrier ID, MCC and MNC.
- `ObjectBase`: APN name, APN type and profile ID.
- `ObjectAuth`: auth type, username and password.
- `ObjectBearer`: protocol, roaming protocol, MTU and server.
- `ObjectProxy`: proxy server and port.
- `ObjectMMS`: MMSC, MMS proxy and MMS port.
- `ObjectMVNO`: MVNO type and match data.
- `ObjectLimit`: max connections and max connection time.
- `ObjectOther`: network bitmask and carrier/user flags.
- `GroupMapByType`: grouped APN records keyed by `ObjectBaseType`.

Most section fields are pointers. A nil pointer means the value is absent and
will be omitted from JSON/XML output.

## XML Grouping

XML import groups valid `<apn>` records by `ObjectRoot.GetID()`. The ID includes
`carrier_id` when present and always includes PLMN:

```text
CID:<carrier_id>;PLMN:<mcc><mnc>;
PLMN:<mcc><mnc>;
```

Only records with both MCC and MNC pass root validation and enter the imported
array.

Inside one grouped `Object`, concrete APN entries are stored in
`GroupMapByType`. The map key is the record's `ObjectBase.Type`.

If a group contains multiple records with the same APN type, the first record is
kept and later records of that type are ignored.

For each group, the root carrier name is taken from the record with the longest
carrier string. After grouping, `Carrier` is normalized with `GetCarrier()`,
which removes common technical suffixes such as APN type names and radio
generation markers.

The imported array is sorted by MCC, MNC and grouping ID.

## Common Helpers

```go
query := &apnxml.Object{
	ObjectRoot: &apnxml.ObjectRoot{
		Mcc: ptr(250),
		Mnc: ptr(1),
	},
}

for i := range apns {
	if apns[i].Match(query) {
		for _, record := range apns[i].Records() {
			// record is either the object itself or one grouped APN entry.
			_ = record
		}
	}
}
```

`Array` helpers:

- `Clone() Array`: returns a deep copy.
- `CountRecords() int`: counts concrete records, including grouped records.
- `String() string`: returns indented JSON or an error string.

`Object` helpers:

- `Clone() *Object`: returns a deep copy.
- `HasGroup() bool`: reports whether `GroupMapByType` has entries.
- `CountRecords() int`: returns 1 for an ungrouped object or group size.
- `GroupTypes() []ObjectBaseType`: returns grouped types in sorted order.
- `Records() []*Object`: returns grouped records in sorted type order, or the
  receiver itself for an ungrouped object.
- `Normalize()`: mutates the object and nested records.
- `NormalizedClone() *Object`: clones and normalizes.
- `Update(*Object, ObjectUpdateMode) bool`: updates fields according to the
  selected update mode.
- `Merge(*Object) bool`: fills only zero target fields from non-zero source
  fields.
- `Patch(*Object) bool`: overwrites target fields with non-zero source fields.
- `Apply(*Object) bool`: replaces target fields, including zero values.
- `Validate() bool`: validates the root identity.
- `Match(*Object) bool`: checks whether the object matches a query object.
- `GetMatchPointer(*Object) *Object`: returns the matching object or grouped
  record.
- `GetID() string`: returns the root grouping ID.
- `GetPLMN() string`: returns MCC/MNC as `%03d%02d`, or `00000` when absent.
- `GetCarrier() string`: returns a simplified carrier name.
- `String() string`: returns indented JSON or an error string.

Section helpers:

- `ObjectRoot`, `ObjectBase`, `ObjectAuth`, `ObjectBearer`, `ObjectProxy`,
  `ObjectMMS`, `ObjectMVNO`, `ObjectLimit` and `ObjectOther` expose section
  versions of `Clone`, `Update`, `Merge`, `Patch`, `Apply`, `Validate` and
  `Match`. Non-root sections also expose `Normalize`.

## Updating

`ObjectUpdateMode` controls how `Update` copies fields from a source object:

```go
const (
	ObjectUpdateMerge ObjectUpdateMode = iota
	ObjectUpdatePatch
	ObjectUpdateApply
)
```

- `ObjectUpdateMerge` fills only zero target fields with non-zero source
  fields.
- `ObjectUpdatePatch` overwrites target fields only when the source field is
  non-zero.
- `ObjectUpdateApply` copies source fields exactly and can clear target fields.

`Object.Merge`, `Object.Patch` and `Object.Apply` are convenience wrappers
around `Object.Update`.

For grouped objects, `Apply` replaces the entire group map. `Merge` and `Patch`
merge source group entries into the target map by `ObjectBaseType`, creating
missing entries and updating existing entries with the same mode.

`ParseObjectUpdateMode` converts CLI-style strings to `ObjectUpdateMode`.
Supported values are `merge`, `patch` and `apply`; an empty value maps to
`patch`.

## Parsing Enum Values

The enum types expose text and JSON unmarshalling. For command-line and config
layers, package-level parse helpers wrap the text unmarshalling API:

- `ParseObjectBaseType(string) (ObjectBaseType, error)`
- `ParseObjectAuthType(string) (ObjectAuthType, error)`
- `ParseObjectBearerProtocol(string) (ObjectBearerProtocol, error)`
- `ParseObjectNetworkType(string) (ObjectNetworkType, error)`

These helpers accept the same token syntax as XML attributes, for example
`default,mms,supl`, `ipv4v6` and `lte,nr`.

## Matching

`Match` treats the right-hand object as a query:

- nil query sections match only nil data sections;
- string fields are case-insensitive substring matches;
- integer fields require equality;
- enum bitmask fields require all query bits to be present;
- grouped objects check the root first, then return a matching grouped
  record.

`ObjectOther.Match` currently matches `NetworkTypeBitmask`; carrier/user boolean
flags are serialized and updated but are not part of matching.

Example:

```go
query := &apnxml.Object{
	ObjectRoot: &apnxml.ObjectRoot{
		Mcc: ptr(250),
		Mnc: ptr(1),
	},
}

for i := range apns {
	if apns[i].Match(query) {
		// matched
	}
}
```

## Normalization

`Normalize` mutates the receiver.

Current normalization rules:

- `ObjectBase.Normalize` sets `Type` to `ObjectBaseTypeDefault` when `Apn` is
  present and `Type` is absent.
- `ObjectAuth.Normalize` fills missing username/password with empty strings
  when auth type is present and at least one credential field is present.
- grouped records are normalized recursively.
- XML unmarshal normalizes each imported `<apn>` record after decoding.
- XML marshal writes normalized clones and does not mutate the original values.

`Validate` is side-effect free and does not normalize data.

## Enum Encoding

The package exposes enum-like integer types:

- `ObjectBaseType`
- `ObjectAuthType`
- `ObjectNetworkType`
- `ObjectBearerProtocol`

They implement text, JSON and XML attribute marshal/unmarshal methods.

JSON uses strings for single-value enums and arrays for bitmask enums. XML uses
the AOSP attribute representation:

- APN `type`: comma-separated names, for example `default,mms`.
- `authtype`: names such as `pap` or `chap`.
- `network_type_bitmask`: pipe-separated numeric order values.
- `protocol` and `roaming_protocol`: upper-case protocol names such as
  `IPV4V6`.

Invalid enum names, invalid enum numbers and empty JSON enum payloads return
errors.

## Format Values

`Format` selects the serializer for reader/writer and URL helpers:

```go
const (
	FormatJSON Format = "json"
	FormatXML  Format = "xml"
)
```

Unsupported formats and unsupported file extensions return errors.

## Package Boundary

Keep this package focused on:

- APN data structures;
- XML/JSON serialization;
- enum conversion;
- basic normalization, cloning and matching;
- stable XML grouping rules.

Broader workflows such as complex filtering, batch mutation, reporting and
application-level import/export orchestration belong outside `pkg/apnxml`.
