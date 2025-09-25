package apnxml

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//--------------------------------------------------------------------------------//
// APNXML Import Method
//--------------------------------------------------------------------------------//

func ImportFromJSONByte(jsonByte []byte) (apnArray APNArray, err error) {
	err = json.Unmarshal(jsonByte, &apnArray)
	return
}

func ImportFromXMLByte(xmlByte []byte) (apnArray APNArray, err error) {
	err = xml.Unmarshal(xmlByte, &apnArray)
	return
}

func ImportFromFile(filename string) (apnArray APNArray, err error) {
	var (
		apnArrayByte []byte
		filenameExt  = strings.ToLower(filepath.Ext(filename))
	)

	apnArrayByte, err = os.ReadFile(filename)
	if err != nil {
		return
	}

	switch filenameExt {
	case ".json":
		return ImportFromJSONByte(apnArrayByte)
	case ".xml":
		return ImportFromXMLByte(apnArrayByte)
	default:
		return nil, fmt.Errorf("apn array file has unsupported ext: %s", filenameExt)
	}
}

func ImportFromUrl(urllink string, isBase64 bool) (apnArray APNArray, err error) {
	var (
		httpGetResponse *http.Response
		apnArrayByte    []byte
	)

	httpGetResponse, err = http.Get(urllink)
	if err != nil {
		return nil, fmt.Errorf("apn array url has fetch error: %v", err)
	}
	defer httpGetResponse.Body.Close()

	if httpGetResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("apn array url has status error: %d", httpGetResponse.StatusCode)
	}

	apnArrayByte, err = io.ReadAll(httpGetResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("apn array url has body error: %v", err)
	}

	if isBase64 {
		apnArrayByte, err = base64.StdEncoding.DecodeString(string(apnArrayByte))
		if err != nil {
			return nil, fmt.Errorf("apn array url has base64 error: %v", err)
		}
	}

	return ImportFromXMLByte(apnArrayByte)
}

//--------------------------------------------------------------------------------//
// APNXML Export Method
//--------------------------------------------------------------------------------//

func ExportToJSONByte(apnArray APNArray) (jsonByte []byte, err error) {
	return json.MarshalIndent(apnArray, "", "\t")
}

func ExportToXMLByte(apnArray APNArray) (xmlByte []byte, err error) {
	return xml.MarshalIndent(apnArray, "", "\t")
}

func ExportToFile(apnArray APNArray, filename string) error {
	var (
		apnArrayByte []byte
		filenameExt  = strings.ToLower(filepath.Ext(filename))
		err          error
	)

	switch filenameExt {
	case ".json":
		apnArrayByte, err = ExportToJSONByte(apnArray)
	case ".xml":
		apnArrayByte, err = ExportToXMLByte(apnArray)
	default:
		err = fmt.Errorf("apn array file has unsupported ext: %s", filenameExt)
	}

	if err != nil {
		return err
	}

	return os.WriteFile(filename, apnArrayByte, 0644)
}

//--------------------------------------------------------------------------------//
