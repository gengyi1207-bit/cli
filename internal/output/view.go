// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package output

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

// ViewFormat is the output format for single-object "view" commands
// (config strict-mode / default-as, ...). It is distinct from Format,
// which targets paginated API-result lists (json/ndjson/table/csv) and
// has no yaml variant.
type ViewFormat int

const (
	ViewFormatText ViewFormat = iota
	ViewFormatJSON
	ViewFormatYAML
)

// String returns the string representation of a ViewFormat.
func (f ViewFormat) String() string {
	switch f {
	case ViewFormatJSON:
		return "json"
	case ViewFormatYAML:
		return "yaml"
	default:
		return "text"
	}
}

// ParseViewFormat parses "text"/"json"/"yaml" (case-insensitive; "" -> text).
// The second return value is false if the string was not recognized, in
// which case ViewFormatText is returned as default.
func ParseViewFormat(s string) (ViewFormat, bool) {
	switch strings.ToLower(s) {
	case "", "text":
		return ViewFormatText, true
	case "json":
		return ViewFormatJSON, true
	case "yaml":
		return ViewFormatYAML, true
	default:
		return ViewFormatText, false
	}
}

// RenderView writes data in the requested format. For ViewFormatText it
// calls textFn — the caller's existing human-readable rendering, unchanged.
// For JSON/YAML it marshals data directly. Marshal/write failures are
// reported to stderr and never fail the command, mirroring FormatValue.
func RenderView(w io.Writer, format ViewFormat, data interface{}, textFn func(io.Writer) error) {
	var err error
	switch format {
	case ViewFormatJSON:
		err = WriteJSON(w, data)
	case ViewFormatYAML:
		err = WriteYAML(w, data)
	default:
		err = textFn(w)
	}
	if isOutputMarshalError(err) {
		legacyStderrf("%s marshal error: %v\n", format, err)
	}
}

// WriteYAML writes data as YAML to w and returns marshal or write errors.
func WriteYAML(w io.Writer, data interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = &outputMarshalError{err: fmt.Errorf("yaml marshal panic: %v", r)}
		}
	}()

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if encErr := enc.Encode(data); encErr != nil {
		_ = enc.Close()
		return &outputMarshalError{err: encErr}
	}
	if closeErr := enc.Close(); closeErr != nil {
		return &outputMarshalError{err: closeErr}
	}
	_, writeErr := w.Write(buf.Bytes())
	return writeErr
}
