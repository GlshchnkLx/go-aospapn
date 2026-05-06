package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/GlshchnkLx/go-aospapn/pkg/apnxml"
)

func loadAPNs(flags *commonFlags) (apnxml.Array, error) {
	format, err := inputFormat(flags)
	if err != nil {
		return nil, err
	}
	if flags.url != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return apnxml.ImportFromURL(ctx, http.DefaultClient, flags.url, format, flags.base64)
	}
	if flags.stdin {
		return apnxml.ImportFromReader(os.Stdin, format)
	}
	if flags.in != "" {
		return loadFile(flags.in, flags.inputFormat)
	}
	return nil, fmt.Errorf("input is required: use --in, --stdin, or --url")
}

func loadFile(path string, formatValue string) (apnxml.Array, error) {
	if formatValue == "" {
		return apnxml.ImportFromFile(path)
	}
	format, err := apnxml.ParseFormat(formatValue)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return apnxml.ImportFromReader(file, format)
}

func inputFormat(flags *commonFlags) (apnxml.Format, error) {
	if flags.inputFormat != "" {
		return apnxml.ParseFormat(flags.inputFormat)
	}
	if flags.in != "" {
		return apnxml.FormatFromFilename(flags.in)
	}
	if flags.url != "" {
		return apnxml.FormatXML, nil
	}
	return "", fmt.Errorf("--input-format is required for stdin")
}
