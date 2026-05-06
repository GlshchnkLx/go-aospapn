package apnxml

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
)

//--------------------------------------------------------------------------------//
// Array
//--------------------------------------------------------------------------------//

type Array []Object

func (apnArray Array) Clone() Array {
	if apnArray == nil {
		return nil
	}

	apnArrayClone := make(Array, 0, len(apnArray))
	for index := range apnArray {
		apnPointer := apnArray[index].Clone()
		if apnPointer == nil {
			continue
		}

		apnArrayClone = append(apnArrayClone, *apnPointer)
	}

	return apnArrayClone
}

func (apnArray Array) CountRecords() int {
	total := 0
	for index := range apnArray {
		total += apnArray[index].CountRecords()
	}

	return total
}

func (apnArray Array) String() string {
	jsonData, err := json.MarshalIndent(apnArray, "", "\t")
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	return string(jsonData)
}

func (apnArray Array) MarshalXML(xmlEncoder *xml.Encoder, _ xml.StartElement) error {
	var (
		xmlStart xml.StartElement
		err      error
	)

	xmlEncoder.Indent("", "\t")

	xmlStart = xml.StartElement{
		Name: xml.Name{
			Local: "apns",
		},
		Attr: []xml.Attr{
			{
				Name: xml.Name{
					Local: "version",
				},
				Value: "8",
			},
		},
	}

	err = xmlEncoder.EncodeToken(xmlStart)
	if err != nil {
		return err
	}

	for _, apnObjectRoot := range apnArray {
		apnPointerRoot := apnObjectRoot.NormalizedClone()
		if apnPointerRoot == nil {
			continue
		}

		if apnPointerRoot.GroupMapByType == nil {
			err = xmlEncoder.EncodeElement(apnPointerRoot, xml.StartElement{
				Name: xml.Name{
					Local: "apn",
				},
			})
		} else {
			var (
				apnPointerBaseTypeArray []ObjectBaseType
				apnPointer              *Object
			)

			for apnPointerBaseTypeString := range apnPointerRoot.GroupMapByType {
				apnPointerBaseTypeArray = append(apnPointerBaseTypeArray, apnPointerBaseTypeString)
			}

			sort.Slice(apnPointerBaseTypeArray, func(i, j int) bool {
				return apnPointerBaseTypeArray[i] < apnPointerBaseTypeArray[j]
			})

			for _, apnPointerBaseTypeString := range apnPointerBaseTypeArray {
				apnPointer = apnPointerRoot.GroupMapByType[apnPointerBaseTypeString].NormalizedClone()
				if apnPointer == nil {
					continue
				}
				apnPointer.ObjectRoot = apnPointerRoot.ObjectRoot.Clone()

				err = xmlEncoder.EncodeElement(apnPointer, xml.StartElement{
					Name: xml.Name{
						Local: "apn",
					},
				})
			}
		}

		if err != nil {
			return err
		}
	}

	err = xmlEncoder.EncodeToken(xmlStart.End())
	if err != nil {
		return err
	}

	return xmlEncoder.Flush()
}

func (apnArray *Array) UnmarshalXML(xmlDecoder *xml.Decoder, xmlStart xml.StartElement) error {
	var (
		apnPointerArrayMap = map[string][]*Object{}
		apnPointerRootMap  = map[string]*ObjectRoot{}
	)

	if xmlStart.Name.Local != "apns" {
		return fmt.Errorf("apn xml has incorrect root element: %q", xmlStart.Name.Local)
	}

	for {
		var (
			xmlDecoderToken xml.Token
			apnObject       Object
			apnObjectBaseID string
			apnPointerRoot  *ObjectRoot
			err             error
		)

		xmlDecoderToken, err = xmlDecoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		switch xmlDecoderElement := xmlDecoderToken.(type) {
		case xml.StartElement:
			if xmlDecoderElement.Name.Local == "apn" {
				err = xmlDecoder.DecodeElement(&apnObject, &xmlDecoderElement)
				if err != nil {
					return err
				}

				if apnObject.ObjectRoot.Validate() {
					apnObjectBaseID = apnObject.GetID()
					apnPointerArrayMap[apnObjectBaseID] = append(apnPointerArrayMap[apnObjectBaseID], &apnObject)

					apnPointerRoot = apnPointerRootMap[apnObjectBaseID]
					if apnPointerRoot == nil || len(apnPointerRoot.Carrier) < len(apnObject.Carrier) {
						apnPointerRootMap[apnObjectBaseID] = apnObject.ObjectRoot
					}
				}
			}
		case xml.EndElement:
			if xmlDecoderElement.Name == xmlStart.Name {
				break
			}
		}
	}

	for apnObjectBaseID, apnPointerArray := range apnPointerArrayMap {
		apnPointerRoot := apnPointerRootMap[apnObjectBaseID]

		apnObject := Object{
			ObjectRoot:     apnPointerRoot.Clone(),
			GroupMapByType: map[ObjectBaseType]*Object{},
		}

		apnObject.Carrier = apnObject.GetCarrier()

		for _, apnPointer := range apnPointerArray {
			if apnPointer.Base == nil || apnPointer.Base.Type == nil {
				continue
			}

			apnPointer.ObjectRoot = nil
			if _, ok := apnObject.GroupMapByType[*apnPointer.Base.Type]; !ok {
				apnObject.GroupMapByType[*apnPointer.Base.Type] = apnPointer.Clone()
			}
		}

		*apnArray = append(*apnArray, apnObject)
	}

	sort.Slice(*apnArray, func(i, j int) bool {
		var (
			mccA, mncA = *(*apnArray)[i].Mcc, *(*apnArray)[i].Mnc
			mccB, mncB = *(*apnArray)[j].Mcc, *(*apnArray)[j].Mnc
			cidA, cidB = (*apnArray)[i].GetID(), (*apnArray)[j].GetID()
		)

		if mccA != mccB {
			return mccA < mccB
		} else {
			if mncA != mncB {
				return mncA < mncB
			} else {
				return cidA < cidB
			}
		}
	})

	return nil
}

//--------------------------------------------------------------------------------//
