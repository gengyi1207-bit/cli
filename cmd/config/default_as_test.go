// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package config

import (
	"encoding/json"
	"errors"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/larksuite/cli/errs"
	"github.com/larksuite/cli/internal/cmdutil"
	"github.com/larksuite/cli/internal/core"
)

func TestDefaultAs_Show_Default(t *testing.T) {
	setupStrictModeTestConfig(t)
	f, stdout, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd := NewCmdConfigDefaultAs(f)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if stdout.String() != "default-as: auto\n" {
		t.Errorf("stdout = %q, want %q", stdout.String(), "default-as: auto\n")
	}
}

func TestDefaultAs_Show_JSON(t *testing.T) {
	setupStrictModeTestConfig(t)
	f, stdout, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd := NewCmdConfigDefaultAs(f)
	cmd.SetArgs([]string{"--output", "json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var got map[string]string
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v (output: %s)", err, stdout.String())
	}
	if got["default_as"] != "auto" {
		t.Errorf("default_as = %q, want %q", got["default_as"], "auto")
	}
}

func TestDefaultAs_Show_YAML_AfterSet(t *testing.T) {
	setupStrictModeTestConfig(t)
	f, _, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd := NewCmdConfigDefaultAs(f)
	cmd.SetArgs([]string{"user"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	f2, stdout, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd2 := NewCmdConfigDefaultAs(f2)
	cmd2.SetArgs([]string{"--output", "yaml"})
	if err := cmd2.Execute(); err != nil {
		t.Fatal(err)
	}
	var got map[string]string
	if err := yaml.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid YAML: %v (output: %s)", err, stdout.String())
	}
	if got["default_as"] != "user" {
		t.Errorf("default_as = %q, want %q", got["default_as"], "user")
	}
}

func TestDefaultAs_Show_InvalidOutputValue(t *testing.T) {
	setupStrictModeTestConfig(t)
	f, _, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd := NewCmdConfigDefaultAs(f)
	cmd.SetArgs([]string{"--output", "xml"})
	err := cmd.Execute()
	var verr *errs.ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("expected *errs.ValidationError, got %T: %v", err, err)
	}
	if verr.Param != "--output" {
		t.Errorf("Param = %q, want %q", verr.Param, "--output")
	}
}

func TestDefaultAs_Set_RejectsNonTextOutput(t *testing.T) {
	setupStrictModeTestConfig(t)
	f, _, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd := NewCmdConfigDefaultAs(f)
	cmd.SetArgs([]string{"user", "--output", "json"})
	err := cmd.Execute()
	var verr *errs.ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("expected *errs.ValidationError, got %T: %v", err, err)
	}
	if verr.Param != "--output" {
		t.Errorf("Param = %q, want %q", verr.Param, "--output")
	}

	multi, _ := core.LoadMultiAppConfig()
	app := multi.CurrentAppConfig("")
	if app.DefaultAs != "" {
		t.Errorf("expected DefaultAs to remain unset, got %q", app.DefaultAs)
	}
}

func TestDefaultAs_Set_InvalidIdentity(t *testing.T) {
	setupStrictModeTestConfig(t)
	f, _, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd := NewCmdConfigDefaultAs(f)
	cmd.SetArgs([]string{"xml"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid identity type 'xml'")
	}
}
