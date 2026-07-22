// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package config

import (
	"errors"
	"testing"

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
	want := "strict-mode: off (source: global (default))\n"
	if got := stdout.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
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
	want := "{\n  \"strict_mode\": \"off\",\n  \"source\": \"global (default)\"\n}\n"
	if got := stdout.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
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
	// "off" collides with a YAML 1.1 boolean literal; yaml.v3 quotes it.
	want := "strict_mode: \"off\"\nsource: global (default)\n"
	if got := stdout.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStrictMode_Show_CredentialProviderSource(t *testing.T) {
	setupStrictModeTestConfig(t)
	// SupportedIdentities: 2 makes the runtime credential chain resolve to
	// "bot", diverging from the persisted profile's default "off" — this
	// exercises the runtime != configMode branch in showStrictMode.
	f, stdout, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret", SupportedIdentities: 2})
	cmd := NewCmdConfigStrictMode(f)
	cmd.SetArgs([]string{"--output", "json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	want := "{\n  \"strict_mode\": \"bot\",\n  \"source\": \"credential provider\"\n}\n"
	if got := stdout.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestStrictMode_Show_InvalidOutput(t *testing.T) {
	setupStrictModeTestConfig(t)
	f, _, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd := NewCmdConfigStrictMode(f)
	cmd.SetArgs([]string{"--output", "xml"})
	err := cmd.Execute()
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
