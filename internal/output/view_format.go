// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package output

import (
	"encoding/json"
	"io"

	"gopkg.in/yaml.v3"

	"github.com/larksuite/cli/errs"
)

// ViewFormat is the output format for simple read-only "view" commands
// (e.g. `config default-as`, `config strict-mode`). It is intentionally
// separate from Format/ParseFormat/Emitter, which serve a different set
// of commands and already use the string "yaml" in tests as a stand-in
// for an unrecognized format.
type ViewFormat string

const (
	ViewFormatText ViewFormat = "text"
	ViewFormatJSON ViewFormat = "json"
	ViewFormatYAML ViewFormat = "yaml"
)

// ParseViewFormat validates the raw --output flag value. Empty string is
// treated as the default ("text"). Unknown values are a hard validation
// error (not a silent fallback) — this flag is new, so there is no legacy
// leniency to preserve, and silent fallback would hide typos from AI agents.
func ParseViewFormat(raw string) (ViewFormat, error) {
	switch ViewFormat(raw) {
	case "", ViewFormatText:
		return ViewFormatText, nil
	case ViewFormatJSON:
		return ViewFormatJSON, nil
	case ViewFormatYAML:
		return ViewFormatYAML, nil
	default:
		return "", errs.NewValidationError(errs.SubtypeInvalidArgument,
			"invalid --output value %q, valid values: text | json | yaml", raw).WithParam("output")
	}
}

// WriteView renders data according to format. For ViewFormatText, it
// delegates to renderText (the command's existing fmt.Fprintf logic,
// unchanged). For JSON/YAML, it marshals data directly — callers pass a
// struct with json/yaml tags (snake_case, matching config's other
// structured output in policy.go/plugins.go). Use a struct, not a map:
// encoding/json sorts map keys alphabetically, which would silently
// reorder fields.
func WriteView(w io.Writer, format ViewFormat, data interface{}, renderText func(io.Writer) error) error {
	switch format {
	case ViewFormatJSON:
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return wrapWriteViewErr(enc.Encode(data), "failed to render json: %v")
	case ViewFormatYAML:
		out, err := yaml.Marshal(data)
		if err != nil {
			return errs.NewInternalError(errs.SubtypeUnknown, "failed to render yaml: %v", err).WithCause(err)
		}
		_, err = w.Write(out)
		return wrapWriteViewErr(err, "failed to write yaml output: %v")
	default:
		return wrapWriteViewErr(renderText(w), "failed to render text: %v")
	}
}

// wrapWriteViewErr preserves already-typed errs errors (e.g. from a caller's
// renderText closure) and wraps everything else as an internal error.
func wrapWriteViewErr(err error, format string) error {
	if err == nil {
		return nil
	}
	if _, ok := errs.ProblemOf(err); ok {
		return err
	}
	return errs.NewInternalError(errs.SubtypeUnknown, format, err).WithCause(err)
}
