// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package config

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/larksuite/cli/errs"
	"github.com/larksuite/cli/internal/cmdutil"
	"github.com/larksuite/cli/internal/core"
)

func setupStrictModeTestConfig(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", dir)
	multi := &core.MultiAppConfig{
		Apps: []core.AppConfig{{
			AppId:     "test-app",
			AppSecret: core.PlainSecret("secret"),
			Brand:     core.BrandFeishu,
		}},
	}
	if err := core.SaveMultiAppConfig(multi); err != nil {
		t.Fatal(err)
	}
}

func TestStrictMode_Show_Default(t *testing.T) {
	setupStrictModeTestConfig(t)
	f, stdout, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd := NewCmdConfigStrictMode(f)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), "off") {
		t.Errorf("expected 'off' in output, got: %s", stdout.String())
	}
}

func TestStrictMode_SetBot_Profile(t *testing.T) {
	setupStrictModeTestConfig(t)
	f, _, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd := NewCmdConfigStrictMode(f)
	cmd.SetArgs([]string{"bot"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	multi, _ := core.LoadMultiAppConfig()
	app := multi.CurrentAppConfig("")
	if app.StrictMode == nil || *app.StrictMode != core.StrictModeBot {
		t.Error("expected StrictMode=bot on profile")
	}
}

func TestStrictMode_SetUser_Profile(t *testing.T) {
	setupStrictModeTestConfig(t)
	f, _, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd := NewCmdConfigStrictMode(f)
	cmd.SetArgs([]string{"user"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	multi, _ := core.LoadMultiAppConfig()
	app := multi.CurrentAppConfig("")
	if app.StrictMode == nil || *app.StrictMode != core.StrictModeUser {
		t.Error("expected StrictMode=user on profile")
	}
}

func TestStrictMode_SetOff_Profile(t *testing.T) {
	setupStrictModeTestConfig(t)
	f, _, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd := NewCmdConfigStrictMode(f)
	cmd.SetArgs([]string{"bot"})
	cmd.Execute()
	cmd = NewCmdConfigStrictMode(f)
	cmd.SetArgs([]string{"off"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	multi, _ := core.LoadMultiAppConfig()
	app := multi.CurrentAppConfig("")
	if app.StrictMode == nil || *app.StrictMode != core.StrictModeOff {
		t.Error("expected StrictMode=off on profile")
	}
}

func TestStrictMode_SetBot_Global(t *testing.T) {
	setupStrictModeTestConfig(t)
	f, _, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd := NewCmdConfigStrictMode(f)
	cmd.SetArgs([]string{"bot", "--global"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	multi, _ := core.LoadMultiAppConfig()
	if multi.StrictMode != core.StrictModeBot {
		t.Error("expected global StrictMode=bot")
	}
}

func TestStrictMode_SetGlobal_DoesNotRequireActiveProfile(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", dir)
	multi := &core.MultiAppConfig{
		CurrentApp: "missing-profile",
		Apps: []core.AppConfig{{
			Name:      "default",
			AppId:     "test-app",
			AppSecret: core.PlainSecret("secret"),
			Brand:     core.BrandFeishu,
		}},
	}
	if err := core.SaveMultiAppConfig(multi); err != nil {
		t.Fatal(err)
	}

	f, _, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd := NewCmdConfigStrictMode(f)
	cmd.SetArgs([]string{"bot", "--global"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	saved, err := core.LoadMultiAppConfig()
	if err != nil {
		t.Fatalf("LoadMultiAppConfig() error = %v", err)
	}
	if saved.StrictMode != core.StrictModeBot {
		t.Fatalf("StrictMode = %q, want %q", saved.StrictMode, core.StrictModeBot)
	}
}

func TestStrictMode_Reset(t *testing.T) {
	setupStrictModeTestConfig(t)
	f, _, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd := NewCmdConfigStrictMode(f)
	cmd.SetArgs([]string{"bot"})
	cmd.Execute()
	cmd = NewCmdConfigStrictMode(f)
	cmd.SetArgs([]string{"--reset"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	multi, _ := core.LoadMultiAppConfig()
	app := multi.CurrentAppConfig("")
	if app.StrictMode != nil {
		t.Errorf("expected nil StrictMode after reset, got %v", *app.StrictMode)
	}
}

func TestStrictMode_InvalidValue(t *testing.T) {
	setupStrictModeTestConfig(t)
	f, _, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd := NewCmdConfigStrictMode(f)
	cmd.SetArgs([]string{"on"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid value 'on'")
	}
}

func TestStrictMode_Show_JSON(t *testing.T) {
	setupStrictModeTestConfig(t)
	f, stdout, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd := NewCmdConfigStrictMode(f)
	cmd.SetArgs([]string{"--output", "json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var got map[string]string
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid JSON: %v (output: %s)", err, stdout.String())
	}
	if got["mode"] != "off" {
		t.Errorf("mode = %q, want %q", got["mode"], "off")
	}
	if got["source"] != "global (default)" {
		t.Errorf("source = %q, want %q", got["source"], "global (default)")
	}
}

func TestStrictMode_Show_YAML(t *testing.T) {
	setupStrictModeTestConfig(t)
	f, stdout, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd := NewCmdConfigStrictMode(f)
	cmd.SetArgs([]string{"--output", "yaml"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	var got map[string]string
	if err := yaml.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("output is not valid YAML: %v (output: %s)", err, stdout.String())
	}
	if got["mode"] != "off" {
		t.Errorf("mode = %q, want %q", got["mode"], "off")
	}
	if got["source"] != "global (default)" {
		t.Errorf("source = %q, want %q", got["source"], "global (default)")
	}
}

func TestStrictMode_Show_InvalidOutputValue(t *testing.T) {
	setupStrictModeTestConfig(t)
	f, _, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd := NewCmdConfigStrictMode(f)
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

func TestStrictMode_Set_RejectsNonTextOutput(t *testing.T) {
	setupStrictModeTestConfig(t)
	f, _, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd := NewCmdConfigStrictMode(f)
	cmd.SetArgs([]string{"bot", "--output", "json"})
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
	if app.StrictMode != nil {
		t.Errorf("expected StrictMode to remain unset, got %v", *app.StrictMode)
	}
}

func TestStrictMode_Reset_RejectsNonTextOutput(t *testing.T) {
	setupStrictModeTestConfig(t)
	f, _, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd := NewCmdConfigStrictMode(f)
	cmd.SetArgs([]string{"bot"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	cmd = NewCmdConfigStrictMode(f)
	cmd.SetArgs([]string{"--reset", "--output", "yaml"})
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
	if app.StrictMode == nil || *app.StrictMode != core.StrictModeBot {
		t.Errorf("expected StrictMode to remain %q, got %v", core.StrictModeBot, app.StrictMode)
	}
}
