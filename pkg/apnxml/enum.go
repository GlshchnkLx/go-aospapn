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
	NoneIndex   Type
	MaxIndex    Type
	IndexArray  []Type
	MapByIndex  map[Type]string
	MapByString map[string]Type
}

func NewEnumMap[Type ~int](noneIndex Type, maxIndex Type, mapByIndex map[Type]string) *EnumMap[Type] {
	coreMapStorage := &EnumMap[Type]{
		NoneIndex:   noneIndex,
		MaxIndex:    maxIndex,
		IndexArray:  []Type{},
		MapByIndex:  map[Type]string{},
		MapByString: map[string]Type{},
	}

	for apnTypeIndex, apnTypeString := range mapByIndex {
		apnTypeString = strings.TrimSpace(apnTypeString)

		if noneIndex < apnTypeIndex && apnTypeIndex < maxIndex {
			coreMapStorage.IndexArray = append(coreMapStorage.IndexArray, apnTypeIndex)
		}

		coreMapStorage.MapByIndex[apnTypeIndex] = apnTypeString
		coreMapStorage.MapByString[apnTypeString] = apnTypeIndex
		coreMapStorage.MapByString[strings.ToLower(apnTypeString)] = apnTypeIndex
	}

	sort.Slice(coreMapStorage.IndexArray, func(i, j int) bool {
		return coreMapStorage.IndexArray[i] < coreMapStorage.IndexArray[j]
	})

	return coreMapStorage
}

func (coreMapStorage *EnumMap[Type]) GetIndex(apnTypeValue Type) Type {
	if _, ok := coreMapStorage.MapByIndex[apnTypeValue]; ok {
		return apnTypeValue
	}

	return coreMapStorage.NoneIndex
}

func (coreMapStorage *EnumMap[Type]) SetIndex(apnTypeValue *Type, apnTypeIndex Type) error {
	if _, ok := coreMapStorage.MapByIndex[apnTypeIndex]; !ok {
		return fmt.Errorf("apn type has incorrect index: %d", apnTypeIndex)
	}

	*apnTypeValue = apnTypeIndex

	return nil
}

func (coreMapStorage *EnumMap[Type]) GetIndexArray(apnTypeValue Type) []Type {
	apnTypeIndexArray := []Type{}
	for _, apnTypeIndex := range coreMapStorage.IndexArray {
		if apnTypeValue&apnTypeIndex == apnTypeIndex {
			apnTypeIndexArray = append(apnTypeIndexArray, apnTypeIndex)
		}
	}

	if len(apnTypeIndexArray) == 0 {
		apnTypeIndexArray = append(apnTypeIndexArray, coreMapStorage.NoneIndex)
	}

	return apnTypeIndexArray
}

func (coreMapStorage *EnumMap[Type]) SetIndexArray(apnTypeValue *Type, apnTypeIndexArray []Type) error {
	*apnTypeValue = coreMapStorage.NoneIndex
	for _, apnTypeIndex := range apnTypeIndexArray {
		if _, ok := coreMapStorage.MapByIndex[apnTypeIndex]; !ok {
			return fmt.Errorf("apn type has incorrect index: %d", apnTypeIndex)
		}

		*apnTypeValue |= apnTypeIndex
	}

	return nil
}

func (coreMapStorage *EnumMap[Type]) GetString(apnTypeValue Type) string {
	return coreMapStorage.MapByIndex[coreMapStorage.GetIndex(apnTypeValue)]
}

func (coreMapStorage *EnumMap[Type]) SetString(apnTypeValue *Type, apnTypeString string) error {
	apnTypeString = strings.ToLower(strings.TrimSpace(apnTypeString))
	apnTypeIndex, ok := coreMapStorage.MapByString[apnTypeString]
	if !ok {
		return fmt.Errorf("apn type has incorrect string: %q", apnTypeString)
	}

	return coreMapStorage.SetIndex(apnTypeValue, apnTypeIndex)
}

func (coreMapStorage *EnumMap[Type]) GetStringArray(apnTypeValue Type) []string {
	var apnTypeStringArray []string
	for _, apnTypeIndex := range coreMapStorage.GetIndexArray(apnTypeValue) {
		apnTypeStringArray = append(apnTypeStringArray, coreMapStorage.MapByIndex[apnTypeIndex])
	}

	return apnTypeStringArray
}

func (coreMapStorage *EnumMap[Type]) SetStringArray(apnTypeValue *Type, apnTypeStringArray []string) error {
	var apnTypeIndexArray []Type
	for _, apnTypeString := range apnTypeStringArray {
		apnTypeString = strings.ToLower(strings.TrimSpace(apnTypeString))
		apnTypeIndex, ok := coreMapStorage.MapByString[apnTypeString]
		if !ok {
			return fmt.Errorf("apn type has incorrect string: %q", apnTypeString)
		}

		apnTypeIndexArray = append(apnTypeIndexArray, apnTypeIndex)
	}

	return coreMapStorage.SetIndexArray(apnTypeValue, apnTypeIndexArray)
}

func (coreMapStorage *EnumMap[Type]) GetValue(apnTypeValue Type) Type {
	if apnTypeValue <= coreMapStorage.NoneIndex || coreMapStorage.MaxIndex <= apnTypeValue {
		return coreMapStorage.NoneIndex
	}

	return apnTypeValue
}

func (coreMapStorage *EnumMap[Type]) SetValue(apnTypeValue *Type, apnTypeIndex Type) error {
	if !(coreMapStorage.NoneIndex <= apnTypeIndex && apnTypeIndex < coreMapStorage.MaxIndex) {
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
		for apnTypeIndex, apnTypeString := range coreProxyStorage.json.MapByIndex {
			if options.xmlStringIsUpper {
				apnTypeString = strings.ToUpper(apnTypeString)
			}

			xmlNames[apnTypeIndex] = apnTypeString
		}
	}

	if options.xmlIsNumber {
		for apnTypeOrder, apnTypeIndex := range coreProxyStorage.json.IndexArray {
			apnTypeString := strconv.Itoa(apnTypeOrder + 1)
			xmlNames[apnTypeIndex] = apnTypeString
		}
	}

	coreProxyStorage.xml = NewEnumMap(
		coreProxyStorage.json.NoneIndex,
		coreProxyStorage.json.MaxIndex,
		xmlNames,
	)

	return coreProxyStorage
}

func (coreProxyStorage *enumCodec[Type]) marshalText(apnTypeValue Type) (textByte []byte, err error) {
	if coreProxyStorage.options.jsonIsArray {
		return []byte(strings.Join(coreProxyStorage.json.GetStringArray(apnTypeValue), ",")), nil
	}

	return []byte(coreProxyStorage.json.GetString(apnTypeValue)), nil
}

func (coreProxyStorage *enumCodec[Type]) unmarshalText(apnTypeValue *Type, textByte []byte) error {
	if coreProxyStorage.options.jsonIsArray {
		return coreProxyStorage.json.SetStringArray(apnTypeValue, strings.Split(string(textByte), ","))
	}

	return coreProxyStorage.json.SetString(apnTypeValue, string(textByte))
}

func (coreProxyStorage *enumCodec[Type]) marshalJSON(apnTypeValue Type) (jsonByte []byte, err error) {
	if coreProxyStorage.options.jsonIsArray {
		return json.Marshal(coreProxyStorage.json.GetStringArray(apnTypeValue))
	}

	return json.Marshal(coreProxyStorage.json.GetString(apnTypeValue))
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

			return coreProxyStorage.json.SetStringArray(apnTypeValue, apnTypeStringArray)
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

		return coreProxyStorage.json.SetString(apnTypeValue, apnTypeString)
	}
}

func (coreProxyStorage *enumCodec[Type]) marshalXMLAttr(apnTypeValue Type, xmlAttrName xml.Name) (xmlAttr xml.Attr, err error) {
	var apnTypeString string

	if coreProxyStorage.options.xmlIsArray {
		apnTypeString = strings.Join(coreProxyStorage.xml.GetStringArray(apnTypeValue), coreProxyStorage.options.xmlArrayHasSeparator)
	} else if coreProxyStorage.options.xmlIsString {
		apnTypeString = coreProxyStorage.xml.GetString(apnTypeValue)
	} else if coreProxyStorage.options.xmlIsNumber {
		if coreProxyStorage.options.xmlNumberIsOrder {
			apnTypeString = coreProxyStorage.xml.GetString(apnTypeValue)
		} else if coreProxyStorage.options.xmlNumberIsIndex {
			apnTypeString = strconv.Itoa(int(coreProxyStorage.xml.GetValue(apnTypeValue)))
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
		return coreProxyStorage.xml.SetStringArray(apnTypeValue, apnTypeStringArray)
	} else if coreProxyStorage.options.xmlIsString {
		return coreProxyStorage.xml.SetString(apnTypeValue, xmlAttr.Value)
	} else if coreProxyStorage.options.xmlIsNumber {
		if coreProxyStorage.options.xmlNumberIsOrder {
			return coreProxyStorage.xml.SetString(apnTypeValue, xmlAttr.Value)
		} else if coreProxyStorage.options.xmlNumberIsIndex {
			apnTypeIndex, err := strconv.Atoi(xmlAttr.Value)
			if err != nil {
				return fmt.Errorf("apn type has invalid number: %v", err)
			}

			if apnTypeIndex < int(coreProxyStorage.xml.NoneIndex) || int(coreProxyStorage.xml.MaxIndex) <= apnTypeIndex {
				return fmt.Errorf("apn type has out of range number: %d", apnTypeIndex)
			}

			*apnTypeValue = Type(apnTypeIndex)
		}
	}
	return nil
}

//--------------------------------------------------------------------------------//
