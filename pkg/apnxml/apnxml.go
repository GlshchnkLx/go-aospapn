package apnxml

import (
	"context"
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
// Format
//--------------------------------------------------------------------------------//

type Format string

const (
	FormatJSON Format = "json"
	FormatXML  Format = "xml"
)

func formatFromFilename(filename string) (Format, error) {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".json":
		return FormatJSON, nil
	case ".xml":
		return FormatXML, nil
	default:
		return "", fmt.Errorf("unsupported apn file extension: %s", filepath.Ext(filename))
	}
}

//--------------------------------------------------------------------------------//
// Decode
//--------------------------------------------------------------------------------//

func decode(data []byte, format Format) (Array, error) {
	var records Array

	switch format {
	case FormatJSON:
		if err := json.Unmarshal(data, &records); err != nil {
			return nil, err
		}
	case FormatXML:
		if err := xml.Unmarshal(data, &records); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported apn format: %s", format)
	}

	return records, nil
}

func ImportFromJSONByte(jsonByte []byte) (apnArray Array, err error) {
	return decode(jsonByte, FormatJSON)
}

func ImportFromXMLByte(xmlByte []byte) (apnArray Array, err error) {
	return decode(xmlByte, FormatXML)
}

func ImportFromReader(reader io.Reader, format Format) (Array, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read apn data: %w", err)
	}

	return decode(data, format)
}

func ImportFromFile(filename string) (apnArray Array, err error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	format, err := formatFromFilename(filename)
	if err != nil {
		return nil, err
	}

	return decode(data, format)
}

func ImportFromURL(ctx context.Context, httpClient *http.Client, url string, format Format, isBase64 bool) (apnArray Array, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create apn request: %w", err)
	}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fetch apn url: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch apn url: unexpected status %d", response.StatusCode)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read apn response body: %w", err)
	}

	if isBase64 {
		data, err = base64.StdEncoding.DecodeString(string(data))
		if err != nil {
			return nil, fmt.Errorf("decode apn base64 payload: %w", err)
		}
	}

	return decode(data, format)
}

func ImportFromSimpleURL(url string, isBase64 bool) (apnArray Array, err error) {
	return ImportFromURL(context.Background(), http.DefaultClient, url, FormatXML, isBase64)
}

//--------------------------------------------------------------------------------//
// Encode
//--------------------------------------------------------------------------------//

func encode(records Array, format Format) ([]byte, error) {
	switch format {
	case FormatJSON:
		return json.MarshalIndent(records, "", "\t")
	case FormatXML:
		return xml.MarshalIndent(records, "", "\t")
	default:
		return nil, fmt.Errorf("unsupported apn format: %s", format)
	}
}

func ExportToJSONByte(apnArray Array) (jsonByte []byte, err error) {
	return encode(apnArray, FormatJSON)
}

func ExportToXMLByte(apnArray Array) (xmlByte []byte, err error) {
	return encode(apnArray, FormatXML)
}

func ExportToWriter(apnArray Array, writer io.Writer, format Format) error {
	data, err := encode(apnArray, format)
	if err != nil {
		return err
	}

	if _, err := writer.Write(data); err != nil {
		return fmt.Errorf("write apn data: %w", err)
	}

	return nil
}

func ExportToFile(apnArray Array, filename string) error {
	format, err := formatFromFilename(filename)
	if err != nil {
		return err
	}

	data, err := encode(apnArray, format)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

//--------------------------------------------------------------------------------//
