# pkg/apntool

`apntool` is the processing layer for APN data after it has been loaded into
the `pkg/apnxml` representation. It does not parse XML, export files, download
data, or contain CLI wiring.

The preferred API is the clone-safe `apntool.Array` wrapper:

```go
result, err := apntool.From(apns).
	Filter(apntool.ByPLMN(250, 1)).
	Apply(func(record *apnxml.Object) error {
		bearer := apntool.EnsureBearer(record)
		apntool.Set(&bearer.Type, protocol, apntool.SetIfEmpty)
		return nil
	})
if err != nil {
	return err
}

output := result.Data()
```

## Data Safety

`From(data)` clones the input by default, and `Data()` returns a clone. All
`Array` methods return a new `Array` or a read-only view of materialized clones.
This keeps processing pipelines from mutating source data accidentally.

Use `From(data, WithTrustedInput())` only when the caller owns `data` and accepts
sharing it with the wrapper.

The package does not expose pointer-based walkers for application code. Mutable
operations are available through `Apply` and `ApplyEntries`; both operate on a
clone and return a new `Array`.

## Record Views

APN data has two useful shapes:

- grouped view: one root object per PLMN/carrier identity with records in
  `GroupMapByType`;
- flat view: one materialized `apnxml.Object` per APN record.

`Array.Flatten` produces flat records. `Array.GroupByPLMN` groups flat records
by MCC/MNC. `Array.GroupByIdentity` groups by `ObjectRoot.GetID()`, which
includes `CarrierID` when present and PLMN.

`MaterializeRecord(group, record)` clones a grouped entry and attaches the group
root when the entry does not have its own root fields.

## Array API

Construction and export:

- `From(data apnxml.Array, opts ...Option) Array`
- `WithTrustedInput() Option`
- `Array.Data() apnxml.Array`
- `Array.Clone() Array`
- `Array.Len() int`
- `Array.CountRecords() int`

Filtering and grouping:

- `Array.Filter(predicate Predicate) Array`
- `Array.Exclude(predicate Predicate) Array`
- `Array.Flatten() Array`
- `Array.GroupByPLMN() Array`
- `Array.GroupByIdentity() Array`
- `Array.DedupeByPLMN() Array`
- `Array.DedupeByIdentity() Array`
- `Array.Merge(other apnxml.Array) Array`
- `Array.Patch(other apnxml.Array) Array`

Iteration and transformation:

- `Array.ForEach(visitor Visitor) error`
- `Array.ForEachGroup(visitor GroupVisitor) error`
- `Array.ForEachEntry(visitor EntryVisitor) error`
- `Array.Map(mapper Mapper) (Array, error)`
- `Array.Apply(mutator Mutator) (Array, error)`
- `Array.ApplyEntries(mutator EntryMutator) (Array, error)`
- `Array.Normalize() (Array, error)`

Lookup helpers:

- `Array.First(predicate Predicate) (apnxml.Object, bool)`
- `Array.Any(predicate Predicate) bool`
- `Array.Count(predicate Predicate) int`
- `Array.Stats() Stats`
- `Array.Types() []apnxml.ObjectBaseType`
- `Array.PLMNs() []string`
- `Array.CarrierIDs() []int`

`ForEach`/`ForEachGroup`/`ForEachEntry` pass cloned values to the callback.
`Map` works like a real map over materialized records and returns a flat array.
`Apply` and `ApplyEntries` preserve the current grouped/flat shape while
mutating a clone.

## Predicates

```go
type Predicate func(record apnxml.Object) bool
```

Available predicates and combinators:

- `All`
- `Not`
- `And`
- `Or`
- `ByPLMN`
- `ByMCC`
- `ByMNC`
- `ByCarrierID`
- `ByType`
- `ByProtocol`
- `ByNetwork`
- `ByAPN`
- `ByAPNContains`
- `ByCarrierName`
- `HasRoot`
- `HasValidRoot`
- `HasBase`
- `HasAuth`
- `HasBearer`
- `HasProxy`
- `HasMMS`
- `HasMVNO`
- `IsValid`
- `Match`

Predicates are intentionally error-free. Operations that can fail should use
`ForEach`, `Map`, or `Apply`.

## Mutation Helpers

Field writes are explicit:

- `SetIfEmpty`: write when the pointer is nil or the value is the zero value;
- `SetIfExists`: write only when the pointer already exists;
- `SetAlways`: overwrite unconditionally.

Helpers:

- `Set`
- `EnsureRoot`
- `EnsureBase`
- `EnsureAuth`
- `EnsureBearer`
- `EnsureProxy`
- `EnsureMMS`
- `EnsureMVNO`
- `EnsureLimit`
- `EnsureOther`
