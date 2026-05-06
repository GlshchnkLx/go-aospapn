package apntool

import (
	"fmt"

	"github.com/GlshchnkLx/go-aospapn/pkg/apnxml"
)

type options struct {
	trustedInput bool
}

type Option func(*options)

func WithTrustedInput() Option {
	return func(options *options) {
		options.trustedInput = true
	}
}

type Array struct {
	data apnxml.Array
}

type Visitor func(record apnxml.Object) error
type GroupVisitor func(group apnxml.Object) error
type EntryVisitor func(group apnxml.Object, record apnxml.Object) error
type Mapper func(record apnxml.Object) (apnxml.Object, error)
type Mutator func(record *apnxml.Object) error
type EntryMutator func(group *apnxml.Object, record *apnxml.Object) error

func From(data apnxml.Array, optionList ...Option) Array {
	var config options
	for _, option := range optionList {
		if option != nil {
			option(&config)
		}
	}

	if config.trustedInput {
		return Array{data: data}
	}

	return Array{data: data.Clone()}
}

func (array Array) Data() apnxml.Array {
	return array.data.Clone()
}

func (array Array) Clone() Array {
	return Array{data: array.data.Clone()}
}

func (array Array) Len() int {
	return len(array.data)
}

func (array Array) CountRecords() int {
	return array.data.CountRecords()
}

func (array Array) Filter(predicate Predicate) Array {
	return Array{data: include(array.data, predicate)}
}

func (array Array) Exclude(predicate Predicate) Array {
	return Array{data: exclude(array.data, predicate)}
}

func (array Array) Flatten() Array {
	return Array{data: flatten(array.data)}
}

func (array Array) GroupByPLMN() Array {
	return Array{data: groupByPLMN(array.data)}
}

func (array Array) GroupByIdentity() Array {
	return Array{data: groupByIdentity(array.data)}
}

func (array Array) Normalize() (Array, error) {
	return array.Apply(func(record *apnxml.Object) error {
		record.Normalize()
		return nil
	})
}

func (array Array) ForEach(visitor Visitor) error {
	if visitor == nil {
		return nil
	}

	for index, record := range flatten(array.data) {
		if err := visitor(record); err != nil {
			return fmt.Errorf("for each record %d: %w", index, err)
		}
	}

	return nil
}

func (array Array) ForEachGroup(visitor GroupVisitor) error {
	if visitor == nil {
		return nil
	}

	for index := range array.data {
		group := MaterializeRecord(nil, &array.data[index])
		if err := visitor(group); err != nil {
			return fmt.Errorf("for each group %d: %w", index, err)
		}
	}

	return nil
}

func (array Array) ForEachEntry(visitor EntryVisitor) error {
	if visitor == nil {
		return nil
	}

	for groupIndex := range array.data {
		group := &array.data[groupIndex]
		groupClone := MaterializeRecord(nil, group)
		records := group.Records()
		for recordIndex, record := range records {
			if record == nil {
				continue
			}

			materializedRecord := MaterializeRecord(group, record)
			if err := visitor(groupClone, materializedRecord); err != nil {
				return fmt.Errorf("for each group %d record %d: %w", groupIndex, recordIndex, err)
			}
		}
	}

	return nil
}

func (array Array) Map(mapper Mapper) (Array, error) {
	records := flatten(array.data)
	if mapper == nil {
		return Array{data: records}, nil
	}

	result := make(apnxml.Array, 0, len(records))
	for index, record := range records {
		mapped, err := mapper(record)
		if err != nil {
			return Array{}, fmt.Errorf("map record %d: %w", index, err)
		}
		result = append(result, mapped)
	}

	return Array{data: result}, nil
}

func (array Array) Apply(mutator Mutator) (Array, error) {
	if mutator == nil {
		return array.Clone(), nil
	}

	return array.ApplyEntries(func(_ *apnxml.Object, record *apnxml.Object) error {
		return mutator(record)
	})
}

func (array Array) ApplyEntries(mutator EntryMutator) (Array, error) {
	result := array.data.Clone()
	if mutator == nil {
		return Array{data: result}, nil
	}

	err := forEachEntryPointer(result, func(group *apnxml.Object, record *apnxml.Object) error {
		return mutator(group, record)
	})
	if err != nil {
		return Array{}, err
	}

	return Array{data: result}, nil
}

func (array Array) First(predicate Predicate) (apnxml.Object, bool) {
	if predicate == nil {
		predicate = All
	}

	for _, record := range flatten(array.data) {
		if predicate(record) {
			return record, true
		}
	}

	return apnxml.Object{}, false
}

func (array Array) Any(predicate Predicate) bool {
	_, ok := array.First(predicate)
	return ok
}

func (array Array) Count(predicate Predicate) int {
	if predicate == nil {
		predicate = All
	}

	count := 0
	_ = array.ForEach(func(record apnxml.Object) error {
		if predicate(record) {
			count++
		}
		return nil
	})

	return count
}
