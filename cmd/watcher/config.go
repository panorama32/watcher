package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/panorama32/watcher/internal/config"
	"github.com/spf13/cobra"
)

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	cmd.AddCommand(configInitCmd())
	cmd.AddCommand(configShowCmd())
	cmd.AddCommand(configSetCmd())
	cmd.AddCommand(configDeleteCmd())

	return cmd
}

func configInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create config file with template",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := config.InitFile(); err != nil {
				return err
			}
			path, _ := config.Path()
			fmt.Printf("Config file created: %s\n", path)
			return nil
		},
	}
}

func configShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			raw, err := config.LoadRaw()
			if err != nil {
				return err
			}
			if len(raw) == 0 {
				fmt.Println("(no configuration set)")
				return nil
			}
			for _, key := range config.ValidKeys() {
				v, ok := raw[key]
				if !ok {
					continue
				}
				display := fmt.Sprintf("%v", v)
				if key == "slack_user_token" && len(display) > 10 {
					display = display[:6] + "..." + display[len(display)-4:]
				}
				fmt.Printf("%s = %q\n", key, display)
			}
			return nil
		},
	}
}

func configSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value := args[0], args[1]
			if !slices.Contains(config.ValidKeys(), key) {
				return fmt.Errorf("unknown key %q (valid keys: %s)", key, strings.Join(config.ValidKeys(), ", "))
			}
			raw, err := config.LoadRaw()
			if err != nil {
				return err
			}
			raw[key] = value
			if err := config.SaveRaw(raw); err != nil {
				return err
			}
			fmt.Printf("%s = %q\n", key, value)
			return nil
		},
	}
}

func configDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <key>",
		Short: "Delete a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			if !slices.Contains(config.ValidKeys(), key) {
				return fmt.Errorf("unknown key %q (valid keys: %s)", key, strings.Join(config.ValidKeys(), ", "))
			}
			raw, err := config.LoadRaw()
			if err != nil {
				return err
			}
			if _, ok := raw[key]; !ok {
				return fmt.Errorf("key %q is not set", key)
			}
			delete(raw, key)
			if err := config.SaveRaw(raw); err != nil {
				return err
			}
			fmt.Printf("Deleted %q\n", key)
			return nil
		},
	}
}
