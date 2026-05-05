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

`ImportFromFile` detects the format from the filename extension. Supported
extensions are `.xml` and `.json`.

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

## Common Helpers

```go
matches := apns.FindByPLMN(250, 1)

for i := range matches {
	for _, record := range matches[i].Records() {
		// record is either the object itself or one grouped APN entry.
		_ = record
	}
}
```

`Array` helpers:

- `Clone() Array`: returns a deep copy.
- `CountRecords() int`: counts concrete records, including grouped records.
- `FindByPLMN(mcc, mnc int) Array`: returns objects with matching MCC/MNC.
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
- `Validate() bool`: validates the root identity.
- `IsLike(*Object) bool`: checks whether the object matches a query object.
- `GetIsLikePointer(*Object) *Object`: returns the matching object or grouped
  record.
- `GetID() string`: returns the root grouping ID.
- `GetPLMN() string`: returns MCC/MNC as `%03d%02d`, or `00000` when absent.
- `GetCarrier() string`: returns a simplified carrier name.
- `String() string`: returns indented JSON or an error string.

## Matching

`IsLike` treats the right-hand object as a query:

- nil query sections match only nil data sections;
- string fields are case-insensitive substring matches;
- integer fields require equality;
- enum bitmask fields require all query bits to be present;
- grouped objects check the root first, then return the first matching grouped
  record.

Example:

```go
query := &apnxml.Object{
	ObjectRoot: &apnxml.ObjectRoot{
		Mcc: ptr(250),
		Mnc: ptr(1),
	},
}

for i := range apns {
	if apns[i].IsLike(query) {
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
