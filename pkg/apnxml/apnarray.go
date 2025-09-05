package apnxml

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
)

// --------------------------------------------------------------------------------//
// APNArray
// --------------------------------------------------------------------------------//

type APNArray []APNObject

func (apnArray APNArray) String() string {
	jsonData, err := json.MarshalIndent(apnArray, "", "\t")
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	return string(jsonData)
}

func (apnArray APNArray) MarshalXML(xmlEncoder *xml.Encoder, start xml.StartElement) error {
	xmlEncoder.Indent("", "\t")

	start.Name.Local = "apns"

	start.Attr = append(start.Attr, xml.Attr{
		Name: xml.Name{
			Local: "version",
		},
		Value: "8",
	})

	err := xmlEncoder.EncodeToken(start)
	if err != nil {
		return err
	}

	for _, apnCore := range apnArray {
		if apnCore.GroupMapByType == nil {
			err := xmlEncoder.EncodeElement(apnCore, xml.StartElement{
				Name: xml.Name{
					Local: "apn",
				},
			})

			if err != nil {
				return err
			}
		} else {
			apnCoreBaseTypeArray := []string{}
			for apnCoreBaseTypeString := range apnCore.GroupMapByType {
				apnCoreBaseTypeArray = append(apnCoreBaseTypeArray, apnCoreBaseTypeString)
			}

			sort.Strings(apnCoreBaseTypeArray)

			for _, apnCoreBaseTypeString := range apnCoreBaseTypeArray {
				apnObject := apnCore.GroupMapByType[apnCoreBaseTypeString]

				apnObject = apnObject.Clone()
				apnObject.APNObjectRoot = apnCore.APNObjectRoot

				if err := xmlEncoder.EncodeElement(apnObject, xml.StartElement{Name: xml.Name{Local: "apn"}}); err != nil {
					return err
				}
			}
		}
	}

	return xmlEncoder.EncodeToken(start.End())
}

func (apnArray *APNArray) UnmarshalXML(xmlDecoder *xml.Decoder, xmlStart xml.StartElement) error {
	var (
		apnGroupArrayMap = map[string][]APNObject{}
		apnCoreMap       = map[string]APNObject{}
	)

	if xmlStart.Name.Local != "apns" {
		return fmt.Errorf("apn xml has uncorrected root element: %s", xmlStart.Name.Local)
	}

	for {
		token, err := xmlDecoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}

		switch elem := token.(type) {
		case xml.StartElement:
			if elem.Name.Local == "apn" {
				var (
					apnObject       APNObject
					apnObjectBaseID string
					apnCore         APNObject
					err             error
				)

				err = xmlDecoder.DecodeElement(&apnObject, &elem)
				if err != nil {
					return err
				}

				if apnObject.APNObjectRoot != nil {
					apnObjectBaseID = apnObject.GetID()

					apnGroupArrayMap[apnObjectBaseID] = append(apnGroupArrayMap[apnObjectBaseID], apnObject)

					apnCore = apnCoreMap[apnObjectBaseID]
					if apnCore.Base == nil || apnCore.Base.Type == nil || *apnCore.Base.Type&APNTYPE_BASE_TYPE_DEFAULT != APNTYPE_BASE_TYPE_DEFAULT {
						apnCoreMap[apnObjectBaseID] = apnObject.Clone()
					}
				}
			}

		case xml.EndElement:
			if elem.Name == xmlStart.Name {
				break
			}
		}
	}

	for apnObjectBaseID, apnGroupArray := range apnGroupArrayMap {
		apnCore := apnCoreMap[apnObjectBaseID]

		apnCore = APNObject{
			APNObjectRoot:  apnCore.APNObjectRoot,
			GroupMapByType: map[string]APNObject{},
		}

		apnCore.APNObjectRoot.Carrier = apnCore.GetCarrier()

		for _, apnObject := range apnGroupArray {
			if apnObject.Base == nil || apnObject.Base.Type == nil {
				continue
			}

			apnObject.APNObjectRoot = nil
			apnCore.GroupMapByType[apnObject.Base.Type.String()] = apnObject
		}

		*apnArray = append(*apnArray, apnCore)
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

func (apnArray APNArray) MarshalJSON() ([]byte, error) {
	return json.MarshalIndent(apnArray, "", "\t")
}

//--------------------------------------------------------------------------------//
