// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package output

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/larksuite/cli/errs"
)

func TestParseViewFormat(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want ViewFormat
	}{
		{"empty defaults to text", "", ViewFormatText},
		{"explicit text", "text", ViewFormatText},
		{"json", "json", ViewFormatJSON},
		{"yaml", "yaml", ViewFormatYAML},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseViewFormat(tc.raw)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseViewFormat_Invalid(t *testing.T) {
	_, err := ParseViewFormat("xml")
	var valErr *errs.ValidationError
	if !errors.As(err, &valErr) {
		t.Fatalf("expected *errs.ValidationError, got %T: %v", err, err)
	}
	if valErr.Subtype != errs.SubtypeInvalidArgument {
		t.Errorf("Subtype = %q, want %q", valErr.Subtype, errs.SubtypeInvalidArgument)
	}
	if valErr.Param != "output" {
		t.Errorf("Param = %q, want %q", valErr.Param, "output")
	}
	if p, ok := errs.ProblemOf(err); !ok || p.Category != errs.CategoryValidation || p.Subtype != errs.SubtypeInvalidArgument {
		t.Errorf("ProblemOf = %+v (ok=%v), want Category=%q Subtype=%q", p, ok, errs.CategoryValidation, errs.SubtypeInvalidArgument)
	}
}

type testView struct {
	Foo string `json:"foo" yaml:"foo"`
}

func TestWriteView_Text(t *testing.T) {
	var buf bytes.Buffer
	called := false
	err := WriteView(&buf, ViewFormatText, testView{Foo: "bar"}, func(w io.Writer) error {
		called = true
		_, err := fmt.Fprint(w, "plain text\n")
		return err
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected renderText to be invoked for ViewFormatText")
	}
	if got, want := buf.String(), "plain text\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriteView_JSON(t *testing.T) {
	var buf bytes.Buffer
	err := WriteView(&buf, ViewFormatJSON, testView{Foo: "bar"}, func(w io.Writer) error {
		t.Fatal("renderText must not be invoked for ViewFormatJSON")
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := buf.String(), "{\n  \"foo\": \"bar\"\n}\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWriteView_YAML(t *testing.T) {
	var buf bytes.Buffer
	err := WriteView(&buf, ViewFormatYAML, testView{Foo: "off"}, func(w io.Writer) error {
		t.Fatal("renderText must not be invoked for ViewFormatYAML")
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// "off" collides with a YAML 1.1 boolean literal; yaml.v3 quotes it to
	// preserve string semantics — verified against the real dependency in
	// the design spec, not assumed.
	if got, want := buf.String(), "foo: \"off\"\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
