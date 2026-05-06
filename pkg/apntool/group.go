package apntool

import "github.com/GlshchnkLx/go-aospapn/pkg/apnxml"

func (array Array) DedupeByPLMN() Array {
	return Array{data: groupByPLMN(flatten(array.data))}
}

func (array Array) DedupeByIdentity() Array {
	return Array{data: groupByIdentity(flatten(array.data))}
}

func (array Array) Merge(other apnxml.Array) Array {
	return Array{data: combine(array.data, other, apnxml.ObjectUpdateMerge)}
}

func (array Array) Patch(other apnxml.Array) Array {
	return Array{data: combine(array.data, other, apnxml.ObjectUpdatePatch)}
}

func (array Array) ApplyUpdate(other apnxml.Array) Array {
	return Array{data: combine(array.data, other, apnxml.ObjectUpdateApply)}
}

func flatten(data apnxml.Array) apnxml.Array {
	var result apnxml.Array
	for groupIndex := range data {
		group := &data[groupIndex]
		for _, record := range group.Records() {
			result = append(result, MaterializeRecord(group, record))
		}
	}

	return result
}

func groupByPLMN(data apnxml.Array) apnxml.Array {
	return groupBy(data, func(record apnxml.Object) string {
		return record.GetPLMN()
	})
}

func groupByIdentity(data apnxml.Array) apnxml.Array {
	return groupBy(data, func(record apnxml.Object) string {
		return record.GetID()
	})
}

func groupBy(data apnxml.Array, key func(apnxml.Object) string) apnxml.Array {
	groupMap := map[string]*apnxml.Object{}
	var groupOrder []string

	for recordIndex := range data {
		record := data[recordIndex]
		if record.ObjectRoot == nil || !record.ObjectRoot.Validate() {
			continue
		}

		groupID := key(record)
		group := groupMap[groupID]
		if group == nil {
			group = &apnxml.Object{
				ObjectRoot:     record.ObjectRoot.Clone(),
				GroupMapByType: map[apnxml.ObjectBaseType]*apnxml.Object{},
			}
			group.Carrier = group.GetCarrier()
			groupMap[groupID] = group
			groupOrder = append(groupOrder, groupID)
		}

		if record.Base == nil || record.Base.Type == nil {
			continue
		}

		recordClone := record.Clone()
		recordClone.ObjectRoot = nil
		if _, exists := group.GroupMapByType[*record.Base.Type]; !exists {
			group.GroupMapByType[*record.Base.Type] = recordClone
		}
	}

	result := make(apnxml.Array, 0, len(groupOrder))
	for _, groupID := range groupOrder {
		result = append(result, *groupMap[groupID])
	}

	return result
}

func combine(left apnxml.Array, right apnxml.Array, mode apnxml.ObjectUpdateMode) apnxml.Array {
	result := groupByIdentity(flatten(left))
	indexByID := make(map[string]int, len(result))
	for index := range result {
		indexByID[result[index].GetID()] = index
	}

	source := groupByIdentity(flatten(right))
	for index := range source {
		sourceGroup := source[index].Clone()
		if sourceGroup == nil {
			continue
		}

		groupID := sourceGroup.GetID()
		targetIndex, ok := indexByID[groupID]
		if !ok {
			indexByID[groupID] = len(result)
			result = append(result, *sourceGroup)
			continue
		}

		result[targetIndex].Update(sourceGroup, mode)
	}

	return result
}
