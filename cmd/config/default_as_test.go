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

func setupDefaultAsTestConfig(t *testing.T, defaultAs core.Identity) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", dir)
	multi := &core.MultiAppConfig{
		Apps: []core.AppConfig{{
			AppId:     "test-app",
			AppSecret: core.PlainSecret("secret"),
			Brand:     core.BrandFeishu,
			DefaultAs: defaultAs,
		}},
	}
	if err := core.SaveMultiAppConfig(multi); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultAs_Show_TextDefault(t *testing.T) {
	setupDefaultAsTestConfig(t, "")
	f, stdout, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd := NewCmdConfigDefaultAs(f)
	cmd.SetArgs([]string{})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if got, want := stdout.String(), "default-as: auto\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDefaultAs_Show_ExplicitText(t *testing.T) {
	setupDefaultAsTestConfig(t, core.Identity("bot"))
	f, stdout, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd := NewCmdConfigDefaultAs(f)
	cmd.SetArgs([]string{"--output", "text"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if got, want := stdout.String(), "default-as: bot\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDefaultAs_Show_JSON(t *testing.T) {
	setupDefaultAsTestConfig(t, core.Identity("bot"))
	f, stdout, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd := NewCmdConfigDefaultAs(f)
	cmd.SetArgs([]string{"--output", "json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if got, want := stdout.String(), "{\n  \"default_as\": \"bot\"\n}\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDefaultAs_Show_YAML(t *testing.T) {
	setupDefaultAsTestConfig(t, core.Identity("user"))
	f, stdout, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd := NewCmdConfigDefaultAs(f)
	cmd.SetArgs([]string{"--output", "yaml"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if got, want := stdout.String(), "default_as: user\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestDefaultAs_Show_InvalidOutput(t *testing.T) {
	setupDefaultAsTestConfig(t, "")
	f, _, _, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd := NewCmdConfigDefaultAs(f)
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

func TestDefaultAs_Set_OutputIsNoop(t *testing.T) {
	setupDefaultAsTestConfig(t, "")
	f, _, stderr, _ := cmdutil.TestFactory(t, &core.CliConfig{AppID: "test-app", AppSecret: "secret"})
	cmd := NewCmdConfigDefaultAs(f)
	cmd.SetArgs([]string{"bot", "--output", "json"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if got, want := stderr.String(), "Default identity set to: bot\n"; got != want {
		t.Errorf("stderr got %q, want %q (--output must not change the set path's plain-text message)", got, want)
	}
	multi, err := core.LoadMultiAppConfig()
	if err != nil {
		t.Fatal(err)
	}
	if app := multi.CurrentAppConfig(""); app.DefaultAs != core.Identity("bot") {
		t.Errorf("expected DefaultAs=bot, got %v", app.DefaultAs)
	}
}
