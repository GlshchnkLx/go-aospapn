package apnxml

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

//--------------------------------------------------------------------------------//
// EnumMap
//--------------------------------------------------------------------------------//

type EnumMap[Type ~int] struct {
	None    Type
	Max     Type
	Indexes []Type
	Names   map[Type]string
	Values  map[string]Type
}

func NewEnumMap[Type ~int](noneIndex Type, maxIndex Type, mapByIndex map[Type]string) *EnumMap[Type] {
	coreMapStorage := &EnumMap[Type]{
		None:    noneIndex,
		Max:     maxIndex,
		Indexes: []Type{},
		Names:   map[Type]string{},
		Values:  map[string]Type{},
	}

	for apnTypeIndex, apnTypeString := range mapByIndex {
		apnTypeString = strings.TrimSpace(apnTypeString)

		if noneIndex < apnTypeIndex && apnTypeIndex < maxIndex {
			coreMapStorage.Indexes = append(coreMapStorage.Indexes, apnTypeIndex)
		}

		coreMapStorage.Names[apnTypeIndex] = apnTypeString
		coreMapStorage.Values[apnTypeString] = apnTypeIndex
		coreMapStorage.Values[strings.ToLower(apnTypeString)] = apnTypeIndex
	}

	sort.Slice(coreMapStorage.Indexes, func(i, j int) bool {
		return coreMapStorage.Indexes[i] < coreMapStorage.Indexes[j]
	})

	return coreMapStorage
}

func (coreMapStorage *EnumMap[Type]) Index(apnTypeValue Type) Type {
	if _, ok := coreMapStorage.Names[apnTypeValue]; ok {
		return apnTypeValue
	}

	return coreMapStorage.None
}

func (coreMapStorage *EnumMap[Type]) SetIndex(apnTypeValue *Type, apnTypeIndex Type) error {
	if _, ok := coreMapStorage.Names[apnTypeIndex]; !ok {
		return fmt.Errorf("apn type has incorrect index: %d", apnTypeIndex)
	}

	*apnTypeValue = apnTypeIndex

	return nil
}

func (coreMapStorage *EnumMap[Type]) IndexesOf(apnTypeValue Type) []Type {
	apnTypeIndexes := []Type{}
	for _, apnTypeIndex := range coreMapStorage.Indexes {
		if apnTypeValue&apnTypeIndex == apnTypeIndex {
			apnTypeIndexes = append(apnTypeIndexes, apnTypeIndex)
		}
	}

	if len(apnTypeIndexes) == 0 {
		apnTypeIndexes = append(apnTypeIndexes, coreMapStorage.None)
	}

	return apnTypeIndexes
}

func (coreMapStorage *EnumMap[Type]) SetIndexes(apnTypeValue *Type, apnTypeIndexes []Type) error {
	*apnTypeValue = coreMapStorage.None
	for _, apnTypeIndex := range apnTypeIndexes {
		if _, ok := coreMapStorage.Names[apnTypeIndex]; !ok {
			return fmt.Errorf("apn type has incorrect index: %d", apnTypeIndex)
		}

		*apnTypeValue |= apnTypeIndex
	}

	return nil
}

func (coreMapStorage *EnumMap[Type]) Name(apnTypeValue Type) string {
	return coreMapStorage.Names[coreMapStorage.Index(apnTypeValue)]
}

func (coreMapStorage *EnumMap[Type]) SetName(apnTypeValue *Type, apnTypeString string) error {
	apnTypeString = strings.ToLower(strings.TrimSpace(apnTypeString))
	apnTypeIndex, ok := coreMapStorage.Values[apnTypeString]
	if !ok {
		return fmt.Errorf("apn type has incorrect string: %q", apnTypeString)
	}

	return coreMapStorage.SetIndex(apnTypeValue, apnTypeIndex)
}

func (coreMapStorage *EnumMap[Type]) NamesOf(apnTypeValue Type) []string {
	var apnTypeStringArray []string
	for _, apnTypeIndex := range coreMapStorage.IndexesOf(apnTypeValue) {
		apnTypeStringArray = append(apnTypeStringArray, coreMapStorage.Names[apnTypeIndex])
	}

	return apnTypeStringArray
}

func (coreMapStorage *EnumMap[Type]) SetNames(apnTypeValue *Type, apnTypeStringArray []string) error {
	var apnTypeIndexes []Type
	for _, apnTypeString := range apnTypeStringArray {
		apnTypeString = strings.ToLower(strings.TrimSpace(apnTypeString))
		apnTypeIndex, ok := coreMapStorage.Values[apnTypeString]
		if !ok {
			return fmt.Errorf("apn type has incorrect string: %q", apnTypeString)
		}

		apnTypeIndexes = append(apnTypeIndexes, apnTypeIndex)
	}

	return coreMapStorage.SetIndexes(apnTypeValue, apnTypeIndexes)
}

func (coreMapStorage *EnumMap[Type]) Value(apnTypeValue Type) Type {
	if apnTypeValue <= coreMapStorage.None || coreMapStorage.Max <= apnTypeValue {
		return coreMapStorage.None
	}

	return apnTypeValue
}

func (coreMapStorage *EnumMap[Type]) SetValue(apnTypeValue *Type, apnTypeIndex Type) error {
	if !(coreMapStorage.None <= apnTypeIndex && apnTypeIndex < coreMapStorage.Max) {
		return fmt.Errorf("apn type has incorrect value: %d", apnTypeIndex)
	}

	*apnTypeValue = apnTypeIndex

	return nil
}

//--------------------------------------------------------------------------------//
// enumCodecOptions
//--------------------------------------------------------------------------------//

type enumCodecOptions struct {
	jsonIsArray bool

	xmlIsArray           bool
	xmlArrayHasSeparator string
	xmlIsString          bool
	xmlStringIsUpper     bool
	xmlIsNumber          bool
	xmlNumberIsOrder     bool
	xmlNumberIsIndex     bool
}

func newEnumCodecOptions() enumCodecOptions {
	return enumCodecOptions{
		jsonIsArray: false,
		xmlIsArray:  false,
		xmlIsString: true,
	}
}

func (coreProxyOption enumCodecOptions) SetJSONIsArray(jsonIsArray bool) enumCodecOptions {
	coreProxyOption.jsonIsArray = jsonIsArray
	return coreProxyOption
}

func (coreProxyOption enumCodecOptions) SetXMLIsArray(xmlArrayHasSeparator string) enumCodecOptions {
	coreProxyOption.xmlIsArray = true
	coreProxyOption.xmlArrayHasSeparator = xmlArrayHasSeparator
	return coreProxyOption
}

func (coreProxyOption enumCodecOptions) SetXMLIsString(xmlStringIsUpper bool) enumCodecOptions {
	coreProxyOption.xmlIsString = true
	coreProxyOption.xmlStringIsUpper = xmlStringIsUpper
	coreProxyOption.xmlIsNumber = false
	return coreProxyOption
}

func (coreProxyOption enumCodecOptions) SetXMLIsNumber(xmlNumberIsOrder bool) enumCodecOptions {
	coreProxyOption.xmlIsNumber = true
	coreProxyOption.xmlNumberIsOrder = xmlNumberIsOrder
	coreProxyOption.xmlNumberIsIndex = !xmlNumberIsOrder
	coreProxyOption.xmlIsString = false
	return coreProxyOption
}

//--------------------------------------------------------------------------------//
// enumCodec
//--------------------------------------------------------------------------------//

type enumCodec[Type ~int] struct {
	json    *EnumMap[Type]
	xml     *EnumMap[Type]
	options enumCodecOptions
}

func newEnumCodec[Type ~int](noneIndex Type, maxIndex Type, mapByIndex map[Type]string, options enumCodecOptions) *enumCodec[Type] {
	coreProxyStorage := &enumCodec[Type]{
		json:    NewEnumMap(noneIndex, maxIndex, mapByIndex),
		options: options,
	}

	xmlNames := map[Type]string{}

	if options.xmlIsString {
		for apnTypeIndex, apnTypeString := range coreProxyStorage.json.Names {
			if options.xmlStringIsUpper {
				apnTypeString = strings.ToUpper(apnTypeString)
			}

			xmlNames[apnTypeIndex] = apnTypeString
		}
	}

	if options.xmlIsNumber {
		for apnTypeOrder, apnTypeIndex := range coreProxyStorage.json.Indexes {
			apnTypeString := strconv.Itoa(apnTypeOrder + 1)
			xmlNames[apnTypeIndex] = apnTypeString
		}
	}

	coreProxyStorage.xml = NewEnumMap(
		coreProxyStorage.json.None,
		coreProxyStorage.json.Max,
		xmlNames,
	)

	return coreProxyStorage
}

func (coreProxyStorage *enumCodec[Type]) marshalText(apnTypeValue Type) (textByte []byte, err error) {
	if coreProxyStorage.options.jsonIsArray {
		return []byte(strings.Join(coreProxyStorage.json.NamesOf(apnTypeValue), ",")), nil
	}

	return []byte(coreProxyStorage.json.Name(apnTypeValue)), nil
}

func (coreProxyStorage *enumCodec[Type]) unmarshalText(apnTypeValue *Type, textByte []byte) error {
	if coreProxyStorage.options.jsonIsArray {
		return coreProxyStorage.json.SetNames(apnTypeValue, strings.Split(string(textByte), ","))
	}

	return coreProxyStorage.json.SetName(apnTypeValue, string(textByte))
}

func (coreProxyStorage *enumCodec[Type]) marshalJSON(apnTypeValue Type) (jsonByte []byte, err error) {
	if coreProxyStorage.options.jsonIsArray {
		return json.Marshal(coreProxyStorage.json.NamesOf(apnTypeValue))
	}

	return json.Marshal(coreProxyStorage.json.Name(apnTypeValue))
}

func (coreProxyStorage *enumCodec[Type]) unmarshalJSON(apnTypeValue *Type, jsonByte []byte) error {
	if len(jsonByte) == 0 {
		return fmt.Errorf("apn type has empty json value")
	}

	if coreProxyStorage.options.jsonIsArray {
		if jsonByte[0] == '[' {
			var apnTypeStringArray []string

			err := json.Unmarshal(jsonByte, &apnTypeStringArray)
			if err != nil {
				return err
			}

			return coreProxyStorage.json.SetNames(apnTypeValue, apnTypeStringArray)
		} else {
			if len(jsonByte) < 2 {
				return fmt.Errorf("apn type has invalid json value: %q", string(jsonByte))
			}
			return coreProxyStorage.unmarshalText(apnTypeValue, jsonByte[1:len(jsonByte)-1])
		}
	} else {
		var apnTypeString string

		err := json.Unmarshal(jsonByte, &apnTypeString)
		if err != nil {
			return err
		}

		return coreProxyStorage.json.SetName(apnTypeValue, apnTypeString)
	}
}

func (coreProxyStorage *enumCodec[Type]) marshalXMLAttr(apnTypeValue Type, xmlAttrName xml.Name) (xmlAttr xml.Attr, err error) {
	var apnTypeString string

	if coreProxyStorage.options.xmlIsArray {
		apnTypeString = strings.Join(coreProxyStorage.xml.NamesOf(apnTypeValue), coreProxyStorage.options.xmlArrayHasSeparator)
	} else if coreProxyStorage.options.xmlIsString {
		apnTypeString = coreProxyStorage.xml.Name(apnTypeValue)
	} else if coreProxyStorage.options.xmlIsNumber {
		if coreProxyStorage.options.xmlNumberIsOrder {
			apnTypeString = coreProxyStorage.xml.Name(apnTypeValue)
		} else if coreProxyStorage.options.xmlNumberIsIndex {
			apnTypeString = strconv.Itoa(int(coreProxyStorage.xml.Value(apnTypeValue)))
		}
	}

	return xml.Attr{
		Name:  xmlAttrName,
		Value: apnTypeString,
	}, nil
}

func (coreProxyStorage *enumCodec[Type]) unmarshalXMLAttr(apnTypeValue *Type, xmlAttr xml.Attr) error {
	if coreProxyStorage.options.xmlIsArray {
		apnTypeStringArray := strings.Split(xmlAttr.Value, coreProxyStorage.options.xmlArrayHasSeparator)
		return coreProxyStorage.xml.SetNames(apnTypeValue, apnTypeStringArray)
	} else if coreProxyStorage.options.xmlIsString {
		return coreProxyStorage.xml.SetName(apnTypeValue, xmlAttr.Value)
	} else if coreProxyStorage.options.xmlIsNumber {
		if coreProxyStorage.options.xmlNumberIsOrder {
			return coreProxyStorage.xml.SetName(apnTypeValue, xmlAttr.Value)
		} else if coreProxyStorage.options.xmlNumberIsIndex {
			apnTypeIndex, err := strconv.Atoi(xmlAttr.Value)
			if err != nil {
				return fmt.Errorf("apn type has invalid number: %v", err)
			}

			if apnTypeIndex < int(coreProxyStorage.xml.None) || int(coreProxyStorage.xml.Max) <= apnTypeIndex {
				return fmt.Errorf("apn type has out of range number: %d", apnTypeIndex)
			}

			*apnTypeValue = Type(apnTypeIndex)
		}
	}
	return nil
}
