# cmd/apnctl

`apnctl` is a small command-line tool for AOSP APN XML/JSON workflows.

It is intentionally built around stdin/stdout so commands can be used in
pipelines.

## Build

```sh
go build ./cmd/apnctl
```

## Example Data

The examples below can be run from the repository root with the bundled files
in `cmd/apnctl/storage`:

- `apns-full-conf.xml`: compact AOSP-like source file.
- `vendor-apns.json`: standalone vendor source for conversion examples.
- `vendor.json`: curated patch file with new and missing records.
- `vendor.xml`: XML variant of vendor patch data.
- `ru-missing-defaults.json`: minimal patch file for a missing default APN.
- `out/`: scratch output directory used by examples.

## Import Current AOSP XML

Android Gitiles returns the file body as base64 when `?format=TEXT` is used:

```sh
go run ./cmd/apnctl fetch \
	--url 'https://android.googlesource.com/device/sample/+/main/etc/apns-full-conf.xml?format=TEXT' \
	--base64 \
	--out cmd/apnctl/storage/out/apns-full-conf.xml
```

If the source is already a local XML or JSON file, use it directly with
`--in`. Most commands infer input format from `.xml` / `.json`; use
`--input-format` when reading from stdin or from a file with a non-standard
extension.

## Inspect and Search

```sh
go run ./cmd/apnctl stats --in cmd/apnctl/storage/apns-full-conf.xml

go run ./cmd/apnctl list \
	--in cmd/apnctl/storage/apns-full-conf.xml \
	--kind plmn \
	--output-format text

go run ./cmd/apnctl find \
	--in cmd/apnctl/storage/apns-full-conf.xml \
	--plmn 25001 \
	--type default \
	--output-format table \
	--limit 10

go run ./cmd/apnctl inspect \
	--in cmd/apnctl/storage/apns-full-conf.xml \
	--plmn 25001
```

Supported search flags include `--plmn`, `--mcc`, `--mnc`, `--carrier-id`,
`--carrier`, `--apn`, `--apn-contains`, `--type`, `--protocol`, `--network`,
`--valid-only`, `--invalid-only`, `--not` and repeated `--has` / `--without`.

`list --kind` supports `plmn`, `type`, `carrier-id`, `carrier` and `apn`.
Output formats include `text`, `json` and `csv`.

Useful inspection examples:

```sh
# Show all PLMNs present in the file.
go run ./cmd/apnctl list \
	--in cmd/apnctl/storage/apns-full-conf.xml \
	--kind plmn

# Show operators for one country by MCC.
go run ./cmd/apnctl list \
	--in cmd/apnctl/storage/apns-full-conf.xml \
	--mcc 250 \
	--kind carrier

# Inspect all APN records for one operator identity.
go run ./cmd/apnctl inspect \
	--in cmd/apnctl/storage/apns-full-conf.xml \
	--plmn 25001

# Export the first 20 default APNs for one country as CSV.
go run ./cmd/apnctl find \
	--in cmd/apnctl/storage/apns-full-conf.xml \
	--mcc 250 \
	--type default \
	--flat \
	--limit 20 \
	--output-format csv \
	--out cmd/apnctl/storage/out/ru-default-apns.csv
```

## Convert

```sh
go run ./cmd/apnctl convert \
	--in cmd/apnctl/storage/apns-full-conf.xml \
	--output-format json \
	--out cmd/apnctl/storage/out/apns.json

cat cmd/apnctl/storage/out/apns.json | go run ./cmd/apnctl convert \
	--stdin \
	--input-format json \
	--output-format xml \
	--out cmd/apnctl/storage/out/apns.xml
```

Shape flags:

- `--flat`
- `--group-by plmn|identity`
- `--dedupe-by plmn|identity`
- `--normalize`
- `--offset N`
- `--limit N`

Conversion examples:

```sh
# Produce flat JSON records instead of grouped objects.
go run ./cmd/apnctl convert \
	--in cmd/apnctl/storage/apns-full-conf.xml \
	--flat \
	--output-format json \
	--out cmd/apnctl/storage/out/apns-flat.json

# Normalize and group records by PLMN only.
go run ./cmd/apnctl convert \
	--in cmd/apnctl/storage/vendor-apns.json \
	--input-format json \
	--normalize \
	--group-by plmn \
	--output-format xml \
	--out cmd/apnctl/storage/out/vendor-apns.xml
```

## Patch

```sh
go run ./cmd/apnctl patch \
	--in cmd/apnctl/storage/apns-full-conf.xml \
	--plmn 25001 \
	--type default \
	--set base.profileID=42 \
	--set other.carrierEnabled=false \
	--mode patch \
	--output-format xml \
	--out cmd/apnctl/storage/out/patched.xml
```

Patch modes are `merge`, `patch` and `apply`.
Use `--strict` to fail when a `--set` patch matches no records.

Mode behavior:

- `merge`: fills only missing target fields.
- `patch`: overwrites target fields when the source field is present.
- `apply`: applies the source shape exactly and can clear fields.

Common `--set` fields:

- root: `carrier`, `carrierID`, `mcc`, `mnc`
- base: `apn`, `type`, `profileID`
- auth: `auth.type`, `auth.username`, `auth.password`
- bearer: `protocol`, `roamingProtocol`, `mtu`, `bearer.server`
- proxy/MMS: `proxy.server`, `proxy.port`, `mmsc`, `mms.server`, `mms.port`
- other: `network`, `enabled`, `visible`, `editable`

Whole-file updates are supported through `--patch-file`:

```sh
go run ./cmd/apnctl patch \
	--in cmd/apnctl/storage/apns-full-conf.xml \
	--patch-file cmd/apnctl/storage/vendor.json \
	--patch-format json \
	--mode merge \
	--output-format xml \
	--out cmd/apnctl/storage/out/merged.xml
```

Batch patch examples:

```sh
# Update a single operator by PLMN.
go run ./cmd/apnctl patch \
	--in cmd/apnctl/storage/apns-full-conf.xml \
	--plmn 25001 \
	--type default \
	--set apn=internet.example \
	--set protocol=ipv4v6 \
	--set roamingProtocol=ipv4v6 \
	--mode patch \
	--output-format xml \
	--out cmd/apnctl/storage/out/apns-25001.xml

# Fill missing defaults for existing default APN records in one country.
# Existing values are preserved because mode=merge only writes empty fields.
go run ./cmd/apnctl patch \
	--in cmd/apnctl/storage/apns-full-conf.xml \
	--mcc 250 \
	--type default \
	--set apn=internet.ru \
	--set auth.type=pap \
	--set auth.username=user \
	--set auth.password=pass \
	--set protocol=ipv4v6 \
	--set roamingProtocol=ipv4v6 \
	--mode merge \
	--output-format xml \
	--out cmd/apnctl/storage/out/apns-ru-defaults.xml

# Apply a curated vendor/country patch file. This is the preferred way to add
# records that do not exist in the source, such as a missing default APN for an
# operator.
go run ./cmd/apnctl patch \
	--in cmd/apnctl/storage/apns-full-conf.xml \
	--patch-file cmd/apnctl/storage/vendor.json \
	--patch-format json \
	--mode merge \
	--output-format xml \
	--out cmd/apnctl/storage/out/apns-with-ru-overrides.xml
```

## Build One Record

```sh
go run ./cmd/apnctl build \
	--carrier "Programmatic Operator" \
	--mcc 999 \
	--mnc 99 \
	--apn iot.example \
	--type default,supl \
	--protocol ipv4v6 \
	--network lte,nr \
	--enabled true \
	--visible false \
	--output-format xml \
	--out cmd/apnctl/storage/out/built.xml
```

Build is useful for creating a small patch file:

```sh
go run ./cmd/apnctl build \
	--carrier "Example RU" \
	--carrier-id 10001 \
	--mcc 250 \
	--mnc 99 \
	--apn internet.ru \
	--type default \
	--protocol ipv4v6 \
	--enabled true \
	--output-format json \
	--out cmd/apnctl/storage/out/ru-overrides.json
```

## Validate

```sh
go run ./cmd/apnctl validate \
	--in cmd/apnctl/storage/apns-full-conf.xml \
	--strict
```

`validate` prints the same counters as `stats` and returns an error in strict
mode when invalid records are present.

## End-to-End Country Update Pipeline

The following pipeline imports AOSP APNs, patches a batch of PLMN-specific
operator values, fills missing fields for existing default APNs in country MCC
`250`, merges a curated file with new/missing records, validates the result and
exports JSON.

```sh
# 1. Start from the bundled local XML fixture.
cp cmd/apnctl/storage/apns-full-conf.xml cmd/apnctl/storage/out/apns-full-conf.xml

# For a live AOSP import, replace the cp command above with:
# go run ./cmd/apnctl fetch \
# 	--url 'https://android.googlesource.com/device/sample/+/main/etc/apns-full-conf.xml?format=TEXT' \
# 	--base64 \
# 	--out cmd/apnctl/storage/out/apns-full-conf.xml
#
# For a JSON source file, replace the cp command above with:
# go run ./cmd/apnctl convert \
# 	--in cmd/apnctl/storage/vendor-apns.json \
# 	--input-format json \
# 	--output-format xml \
# 	--out cmd/apnctl/storage/out/apns-full-conf.xml

# 2. Patch a batch of known operator updates by PLMN.
go run ./cmd/apnctl patch \
	--in cmd/apnctl/storage/out/apns-full-conf.xml \
	--plmn 25001 \
	--type default \
	--set apn=internet.operator-a.ru \
	--set protocol=ipv4v6 \
	--mode patch \
	--output-format xml \
	--out cmd/apnctl/storage/out/step-25001.xml

go run ./cmd/apnctl patch \
	--in cmd/apnctl/storage/out/step-25001.xml \
	--plmn 25002 \
	--type default \
	--set apn=internet.operator-b.ru \
	--set auth.type=pap \
	--set auth.username=user \
	--set auth.password=pass \
	--mode patch \
	--output-format xml \
	--out cmd/apnctl/storage/out/step-25002.xml

# 3. Fill default values for existing default APN records in MCC 250.
# This does not overwrite already populated fields.
go run ./cmd/apnctl patch \
	--in cmd/apnctl/storage/out/step-25002.xml \
	--mcc 250 \
	--type default \
	--set apn=internet.ru \
	--set auth.type=pap \
	--set auth.username=user \
	--set auth.password=pass \
	--set protocol=ipv4v6 \
	--set roamingProtocol=ipv4v6 \
	--mode merge \
	--output-format xml \
	--out cmd/apnctl/storage/out/step-250-defaults.xml

# 4. Add records that are absent in the source through a patch file.
# vendor.json is a curated JSON array with missing default APNs and new records.
go run ./cmd/apnctl patch \
	--in cmd/apnctl/storage/out/step-250-defaults.xml \
	--patch-file cmd/apnctl/storage/vendor.json \
	--patch-format json \
	--mode merge \
	--output-format xml \
	--out cmd/apnctl/storage/out/apns-ru-final.xml

# 5. Validate and export the final JSON.
go run ./cmd/apnctl validate \
	--in cmd/apnctl/storage/out/apns-ru-final.xml \
	--strict

go run ./cmd/apnctl convert \
	--in cmd/apnctl/storage/out/apns-ru-final.xml \
	--output-format json \
	--out cmd/apnctl/storage/out/apns-ru-final.json
```

Notes:

- `--plmn` accepts MCC+MNC, for example `25001`.
- `--mcc 250` applies to all operators in the country.
- Use `--mode merge` for defaults because it fills only absent fields.
- Use `--patch-file` when a whole APN record is missing; `--set` updates only
  records matched by filters.

## Source Layout

The command is split by responsibility:

- `main.go`: command dispatch.
- `flags.go`, `types.go`, `parse.go`: CLI flag and small parse helpers.
- `input.go`, `output.go`, `process.go`, `filter.go`: shared pipeline logic.
- `command_*.go`: individual command implementations.
