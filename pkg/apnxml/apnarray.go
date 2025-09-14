// # APN Array
//
// File defines the APNArray type, which represents a collection of APNObject
// entries. It supports marshaling to and unmarshaling from XML with a root element
// named "apns" and version attribute "8". It also provides a JSON string representation
// via the String method.
//
// The XML structure is expected to contain one or more "apn" elements, each representing
// a carrier's APN configuration. During unmarshaling, entries are grouped by ID and
// sorted by MCC, MNC, and ID. During marshaling, if an APNObject contains grouped
// entries by type, each is encoded as a separate "apn" element while preserving the
// root carrier information.
package apnxml

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
)

//--------------------------------------------------------------------------------//
// APNArray
//--------------------------------------------------------------------------------//

// APNArray is a slice of APNObject, representing a collection of APN configurations.
// It implements custom XML marshaling and unmarshaling to handle the hierarchical
// structure of APN entries, including grouping by type. It also provides a JSON
// string representation via the String method.
type APNArray []APNObject

// String returns a pretty-printed JSON representation of the APNArray.
// If marshaling fails, it returns an error string.
func (apnArray APNArray) String() string {
	jsonData, err := json.MarshalIndent(apnArray, "", "\t")
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}

	return string(jsonData)
}

// MarshalXML encodes the APNArray into XML format with a root element <apns version="8">.
// Each APNObject is encoded as an <apn> element. If the APNObject has grouped entries
// (GroupMapByType), each group is encoded as a separate <apn> element, sorted by type.
// The encoder is set to indent with tabs for readability.
func (apnArray APNArray) MarshalXML(xmlEncoder *xml.Encoder, _ xml.StartElement) error {
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
		apnPointerRoot := apnObjectRoot.Clone()

		if apnPointerRoot.GroupMapByType == nil {
			err = xmlEncoder.EncodeElement(apnPointerRoot, xml.StartElement{
				Name: xml.Name{
					Local: "apn",
				},
			})
		} else {
			var (
				apnPointerBaseTypeArray []string
				apnPointer              *APNObject
			)

			for apnPointerBaseTypeString := range apnPointerRoot.GroupMapByType {
				apnPointerBaseTypeArray = append(apnPointerBaseTypeArray, apnPointerBaseTypeString)
			}

			sort.Strings(apnPointerBaseTypeArray)

			for _, apnPointerBaseTypeString := range apnPointerBaseTypeArray {
				apnPointer = apnPointerRoot.GroupMapByType[apnPointerBaseTypeString]
				apnPointer.APNObjectRoot = apnPointerRoot.APNObjectRoot

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

// UnmarshalXML decodes XML data into the APNArray. It expects a root element <apns>.
// Each <apn> element is decoded into an APNObject. Objects are grouped by their ID
// (derived from MCC/MNC), and grouped entries are consolidated under a single APNObject
// with GroupMapByType. After processing, the array is sorted by MCC, then MNC, then ID.
//
// Returns an error if the root element is not "apns", if XML decoding fails,
// or if an APNObject fails validation.
func (apnArray *APNArray) UnmarshalXML(xmlDecoder *xml.Decoder, xmlStart xml.StartElement) error {
	var (
		apnPointerArrayMap = map[string][]*APNObject{}
		apnPointerRootMap  = map[string]*APNObjectRoot{}
	)

	if xmlStart.Name.Local != "apns" {
		return fmt.Errorf("apn xml has incorrect root element: %q", xmlStart.Name.Local)
	}

	for {
		var (
			xmlDecoderToken xml.Token
			apnObject       APNObject
			apnObjectBaseID string
			apnPointerRoot  *APNObjectRoot
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

				if apnObject.APNObjectRoot.Validate() {
					apnObjectBaseID = apnObject.GetID()
					apnPointerArrayMap[apnObjectBaseID] = append(apnPointerArrayMap[apnObjectBaseID], &apnObject)

					apnPointerRoot = apnPointerRootMap[apnObjectBaseID]
					if apnPointerRoot == nil || len(apnPointerRoot.Carrier) < len(apnObject.Carrier) {
						apnPointerRootMap[apnObjectBaseID] = apnObject.APNObjectRoot
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

		apnObject := APNObject{
			APNObjectRoot:  apnPointerRoot.Clone(),
			GroupMapByType: map[string]*APNObject{},
		}

		apnObject.Carrier = apnObject.GetCarrier()

		for _, apnPointer := range apnPointerArray {
			if apnPointer.Base == nil || apnPointer.Base.Type == nil {
				continue
			}

			apnPointer.APNObjectRoot = nil
			apnObject.GroupMapByType[apnPointer.Base.Type.String()] = apnPointer
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
