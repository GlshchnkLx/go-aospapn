package apnxml

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
)

//--------------------------------------------------------------------------------//
// Helper Method
//--------------------------------------------------------------------------------//

func isValidXML(data []byte) bool {
	return xml.Unmarshal(data, new(interface{})) == nil
}

func isValidJSON(data []byte) bool {
	return json.Unmarshal(data, new(interface{})) == nil
}

//--------------------------------------------------------------------------------//
// APNXML Method
//--------------------------------------------------------------------------------//

func ParseApnArrayFromXmlByte(apnArrayXmlByte []byte) (apnArray APNArray, err error) {
	err = xml.Unmarshal(apnArrayXmlByte, &apnArray)
	return
}

func ParseApnArrayFromJsonByte(apnArrayJsonByte []byte) (apnArray APNArray, err error) {
	err = json.Unmarshal(apnArrayJsonByte, &apnArray)
	return
}

func ParseApnArrayFromByte(apnArrayB64Byte []byte) (apnArray APNArray, err error) {
	var (
		apnArrayByte []byte
	)

	apnArrayByte, err = base64.StdEncoding.DecodeString(string(apnArrayB64Byte))
	if err != nil {
		err = nil
		apnArrayByte = apnArrayB64Byte
	}

	if isValidXML(apnArrayByte) {
		return ParseApnArrayFromXmlByte(apnArrayByte)
	}

	if isValidJSON(apnArrayByte) {
		return ParseApnArrayFromJsonByte(apnArrayByte)
	}

	return nil, fmt.Errorf("apn byte array must be xml or json")
}

func ParseApnArrayFromFile(filename string) (apnArray APNArray, err error) {
	var (
		apnArrayByte []byte
	)

	apnArrayByte, err = os.ReadFile(filename)
	if err != nil {
		return
	}

	return ParseApnArrayFromByte(apnArrayByte)
}

//--------------------------------------------------------------------------------//
