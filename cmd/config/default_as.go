// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package config

import (
	"fmt"
	"io"

	"github.com/larksuite/cli/errs"
	"github.com/larksuite/cli/internal/cmdutil"
	"github.com/larksuite/cli/internal/core"
	"github.com/larksuite/cli/internal/output"
	"github.com/spf13/cobra"
)

// DefaultAsView is the structured (json/yaml) representation of `config
// default-as`'s view-mode output.
type DefaultAsView struct {
	DefaultAs string `json:"default_as" yaml:"default_as"`
}

// NewCmdConfigDefaultAs creates the "config default-as" subcommand.
func NewCmdConfigDefaultAs(f *cmdutil.Factory) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "default-as [user|bot|auto]",
		Short: "View or set default identity type",
		Long:  "Without arguments, shows the current default identity. Pass user, bot, or auto to set a new default.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			multi, err := core.LoadOrNotConfigured()
			if err != nil {
				return err
			}

			app := multi.CurrentAppConfig(f.Invocation.Profile)
			if app == nil {
				return core.NoActiveProfileError()
			}

			if len(args) == 0 {
				format, ok := output.ParseViewFormat(outputFormat)
				if !ok {
					return errs.NewValidationError(errs.SubtypeInvalidArgument,
						"invalid output format %q, valid values: text | json | yaml", outputFormat).WithParam("--output")
				}
				current := app.DefaultAs
				if current == "" {
					current = "auto"
				}
				output.RenderView(f.IOStreams.Out, format, DefaultAsView{DefaultAs: string(current)}, func(w io.Writer) error {
					_, err := fmt.Fprintf(w, "default-as: %s\n", current)
					return err
				})
				return nil
			}

			if outputFormat != "text" {
				return errs.NewValidationError(errs.SubtypeInvalidArgument,
					"--output only applies when viewing (no value argument)").WithParam("--output")
			}

			value := args[0]
			if value != "user" && value != "bot" && value != "auto" {
				return errs.NewValidationError(errs.SubtypeInvalidArgument, "invalid identity type %q, valid values: user | bot | auto", value)
			}

			app.DefaultAs = core.Identity(value)
			if err := core.SaveMultiAppConfig(multi); err != nil {
				return errs.NewInternalError(errs.SubtypeStorage, "failed to save config: %v", err).WithCause(err)
			}
			fmt.Fprintf(f.IOStreams.ErrOut, "Default identity set to: %s\n", value)
			return nil
		},
	}
	cmd.Flags().StringVar(&outputFormat, "output", "text", "output format for viewing: text|json|yaml")
	cmdutil.RegisterFlagCompletion(cmd, "output", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"text", "json", "yaml"}, cobra.ShellCompDirectiveNoFileComp
	})
	cmdutil.SetRisk(cmd, "write")
	return cmd
}
