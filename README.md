# go-aospapn

`go-aospapn` is a Go library and CLI toolkit for working with AOSP APN
(Access Point Name) configuration data.

It can import Android-style `apns-conf.xml`, export XML or JSON, group APN
records by operator identity, transform records safely in Go code, and run
common inspection/conversion/patch workflows from the command line.

## Packages

- [`pkg/apnxml`](pkg/apnxml): low-level APN data model, XML/JSON import/export,
  enum parsing, cloning, normalization, validation, matching and stable XML
  grouping rules.
- [`pkg/apntool`](pkg/apntool): clone-safe processing layer for filtering,
  flattening, grouping, deduplication, mutation and patch-style updates after
  data has been loaded into `apnxml`.
- [`cmd/apnctl`](cmd/apnctl): stdin/stdout-friendly CLI for fetching,
  inspecting, searching, converting, patching, building and validating APN
  files.

## Core Use Cases

- Import and export AOSP APN XML.
- Convert APN tables between XML and JSON.
- Search APN profiles by PLMN (`MCC` + `MNC`), carrier, APN name, type,
  protocol, network bitmask or section presence.
- Process APN records with clone-safe Go pipelines.
- Patch individual fields or merge curated vendor/country APN overrides.
- Build small APN patch files programmatically or from CLI flags.
- Validate APN data before shipping or feeding it to downstream tooling.

## Install

```sh
go get github.com/GlshchnkLx/go-aospapn
```

Build the CLI from the repository root:

```sh
go build ./cmd/apnctl
```

## Go Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/GlshchnkLx/go-aospapn/pkg/apntool"
	"github.com/GlshchnkLx/go-aospapn/pkg/apnxml"
)

func main() {
	apns, err := apnxml.ImportFromFile("apns-conf.xml")
	if err != nil {
		log.Fatal(err)
	}

	matches := apntool.From(apns).
		Filter(apntool.ByPLMN(250, 1)).
		Data()

	for _, match := range matches {
		fmt.Println(match.GetPLMN(), match.Carrier, match.CountRecords())
	}
}
```

`apnxml.ImportFromFile` detects `.xml` and `.json` input by extension. Reader,
byte slice and URL helpers are also available, including base64 response body
decoding for Android Gitiles `?format=TEXT` URLs.

## CLI Examples

Inspect bundled example data:

```sh
go run ./cmd/apnctl stats --in cmd/apnctl/storage/apns-full-conf.xml

go run ./cmd/apnctl find \
	--in cmd/apnctl/storage/apns-full-conf.xml \
	--plmn 25001 \
	--type default \
	--output-format table \
	--limit 10
```

Convert XML to JSON:

```sh
go run ./cmd/apnctl convert \
	--in cmd/apnctl/storage/apns-full-conf.xml \
	--output-format json \
	--out cmd/apnctl/storage/out/apns.json
```

Patch matching records:

```sh
go run ./cmd/apnctl patch \
	--in cmd/apnctl/storage/apns-full-conf.xml \
	--plmn 25001 \
	--type default \
	--set apn=internet.example \
	--set protocol=ipv4v6 \
	--mode patch \
	--output-format xml \
	--out cmd/apnctl/storage/out/patched.xml
```

Fetch current AOSP XML from Android Gitiles:

```sh
go run ./cmd/apnctl fetch \
	--url 'https://android.googlesource.com/device/sample/+/main/etc/apns-full-conf.xml?format=TEXT' \
	--base64 \
	--out cmd/apnctl/storage/out/apns-full-conf.xml
```

## Data Model

`pkg/apnxml` represents APN data as an `apnxml.Array`, a slice of
`apnxml.Object`. Each object contains a root identity section and optional
sections for base APN settings, authentication, bearer/protocol settings,
proxy, MMS, MVNO, limits and carrier/user flags.

Most fields are pointers. A nil pointer means the value is absent and will be
omitted from JSON/XML output.

XML import groups valid `<apn>` records by `ObjectRoot.GetID()`. The grouping
ID includes `carrier_id` when present and always includes PLMN. Inside a group,
`GroupMapByType` keeps one concrete APN record per APN type. If the source XML
contains multiple records with the same group identity and APN type, the first
record is kept.

The imported array is sorted by MCC, MNC and grouping ID. XML export expands
grouped objects back to one `<apn>` element per APN type and writes an
`<apns version="8">` root element.

## Processing API

Use `apntool.From(data)` to start a clone-safe processing pipeline:

```go
result, err := apntool.From(apns).
	Filter(apntool.ByPLMN(250, 1)).
	Apply(func(record *apnxml.Object) error {
		bearer := apntool.EnsureBearer(record)
		apntool.Set(&bearer.Type, apnxml.ObjectBearerProtocolIPv4v6, apntool.SetIfEmpty)
		return nil
	})
if err != nil {
	log.Fatal(err)
}

output := result.Data()
```

`From` clones input by default, and `Data` returns a clone. Mutable operations
are exposed through `Apply`, `ApplyEntries` and update helpers, so application
code does not need pointer-based walkers over shared source data.

Common operations include:

- `Filter`, `Exclude`, `First`, `Any`, `Count`.
- `Flatten`, `GroupByPLMN`, `GroupByIdentity`.
- `DedupeByPLMN`, `DedupeByIdentity`.
- `Merge`, `Patch`, `ApplyUpdate`.
- `Normalize`, `Stats`, `Types`, `PLMNs`, `CarrierIDs`.

## Patch Semantics

Both `apnxml` and `apntool` use the same update modes:

- `merge`: fill only missing target fields.
- `patch`: overwrite target fields when the source field is present.
- `apply`: copy the source shape exactly and allow fields to be cleared.

The CLI exposes these modes through `apnctl patch --mode`.

## More Documentation

- [`pkg/apnxml/README.md`](pkg/apnxml/README.md) covers import/export helpers,
  grouping, normalization, matching, enum encoding and format handling.
- [`pkg/apntool/README.md`](pkg/apntool/README.md) covers clone-safety,
  predicates, grouping/flattening, mutation helpers and field patch
  expressions.
- [`cmd/apnctl/README.md`](cmd/apnctl/README.md) covers CLI commands, flags and
  end-to-end APN update pipelines.
