// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package output

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestParseViewFormat(t *testing.T) {
	tests := []struct {
		input  string
		want   ViewFormat
		wantOK bool
	}{
		{"", ViewFormatText, true},
		{"text", ViewFormatText, true},
		{"TEXT", ViewFormatText, true},
		{"Text", ViewFormatText, true},
		{"json", ViewFormatJSON, true},
		{"JSON", ViewFormatJSON, true},
		{"yaml", ViewFormatYAML, true},
		{"YAML", ViewFormatYAML, true},
		{"xml", ViewFormatText, false},
		{"foobar", ViewFormatText, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := ParseViewFormat(tt.input)
			if got != tt.want {
				t.Errorf("ParseViewFormat(%q) format = %v, want %v", tt.input, got, tt.want)
			}
			if ok != tt.wantOK {
				t.Errorf("ParseViewFormat(%q) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
		})
	}
}

func TestViewFormatString(t *testing.T) {
	tests := []struct {
		format ViewFormat
		want   string
	}{
		{ViewFormatText, "text"},
		{ViewFormatJSON, "json"},
		{ViewFormatYAML, "yaml"},
		{ViewFormat(99), "text"}, // unknown falls back
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.format.String()
			if got != tt.want {
				t.Errorf("ViewFormat(%d).String() = %q, want %q", tt.format, got, tt.want)
			}
		})
	}
}

type testView struct {
	Mode   string `json:"mode" yaml:"mode"`
	Source string `json:"source" yaml:"source"`
}

func TestRenderView_JSON(t *testing.T) {
	var buf bytes.Buffer
	RenderView(&buf, ViewFormatJSON, testView{Mode: "bot", Source: "global"}, func(io.Writer) error {
		t.Fatal("textFn must not be called for ViewFormatJSON")
		return nil
	})

	var got testView
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v (output: %s)", err, buf.String())
	}
	if got.Mode != "bot" || got.Source != "global" {
		t.Errorf("got %+v, want {bot global}", got)
	}
}

func TestRenderView_YAML(t *testing.T) {
	var buf bytes.Buffer
	RenderView(&buf, ViewFormatYAML, testView{Mode: "user", Source: `profile "default"`}, func(io.Writer) error {
		t.Fatal("textFn must not be called for ViewFormatYAML")
		return nil
	})

	var got testView
	if err := yaml.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid YAML: %v (output: %s)", err, buf.String())
	}
	if got.Mode != "user" || got.Source != `profile "default"` {
		t.Errorf("got %+v, want {user profile \"default\"}", got)
	}
}

func TestRenderView_Text(t *testing.T) {
	var buf bytes.Buffer
	called := false
	RenderView(&buf, ViewFormatText, testView{Mode: "off"}, func(w io.Writer) error {
		called = true
		_, err := w.Write([]byte("strict-mode: off (source: global (default))\n"))
		return err
	})

	if !called {
		t.Fatal("textFn was not called for ViewFormatText")
	}
	if buf.String() != "strict-mode: off (source: global (default))\n" {
		t.Errorf("got %q", buf.String())
	}
}
