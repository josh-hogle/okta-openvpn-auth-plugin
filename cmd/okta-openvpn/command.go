package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/josh-hogle/okta-openvpn-auth-plugin/internal/app"
	"github.com/josh-hogle/okta-openvpn-auth-plugin/internal/commands/auth"
	"github.com/josh-hogle/okta-openvpn-auth-plugin/internal/commands/version"
	"github.com/josh-hogle/okta-openvpn-auth-plugin/internal/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.innotegrity.dev/zerolog"
	"go.innotegrity.dev/zerolog/log"
)

// RootCommand is the base command for the application.
type RootCommand struct {
	cobra.Command

	// unexported variables
	configFile   string
	visitedFlags map[string]bool
}

// NewRootCommand creates a new Command object.
func NewRootCommand() *RootCommand {
	cmd := &RootCommand{
		visitedFlags: map[string]bool{},
	}
	cmd.Use = "okta-openvpn"
	cmd.Short = app.Title
	cmd.Long = fmt.Sprintf("%s allows you to authenticate OpenVPN users utilizing Okta's SSO and MFA capabilities.",
		app.Title)
	cmd.PersistentPreRunE = cmd.persistentPreRunE
	cmd.PersistentPostRunE = cmd.persistentPostRunE
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	persistentFlags := cmd.PersistentFlags()

	// flags stored by the command
	persistentFlags.StringVarP(&cmd.configFile, "config-file", "f", "",
		"alternate path to configuration file")

	// flags stored by viper
	persistentFlags.BoolP("enable-json-logging", "j", false, "Format output messages as JSON")
	viper.BindPFlag("global.enable_json_logging", persistentFlags.Lookup("enable-json-logging"))
	viper.BindEnv("global.enable_json_logging", fmt.Sprintf("%sENABLE_JSON_LOGGING", app.EnvVarPrefix))

	persistentFlags.StringP("log-level", "l", app.DefaultLogLevel,
		"Set logging level to debug, info, warn, error, fatal, panic or none")
	viper.BindPFlag("global.log_level", persistentFlags.Lookup("log-level"))
	viper.BindEnv("global.log_level", fmt.Sprintf("%sLOG_LEVEL", app.EnvVarPrefix))

	// add commands
	cmd.AddCommand(&auth.NewCommand().Command)
	cmd.AddCommand(&version.NewCommand().Command)

	return cmd
}

// persistentPostRunE is called after the command is executed.
func (c *RootCommand) persistentPostRunE(cmd *cobra.Command, args []string) error {
	return nil
}

// persistentPreRunE is called before any command is executed.
func (c *RootCommand) persistentPreRunE(cmd *cobra.Command, args []string) error {
	// update the list of flags specified on the command-line
	c.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		if f.Changed {
			c.visitedFlags[f.Name] = true
		}
	})
	c.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Changed {
			c.visitedFlags[f.Name] = true
		}
	})

	// set special environment variables
	env := map[string]string{
		fmt.Sprintf("%sVERSION", app.EnvVarPrefix): app.Version,
		fmt.Sprintf("%sBUILD", app.EnvVarPrefix):   app.Build,
		"GOOS":                                     runtime.GOOS,
		"GOARCH":                                   runtime.GOARCH,
	}
	for k, v := range env {
		if err := os.Setenv(k, v); err != nil {
			e := &errors.GeneralFailure{
				Err: err,
				Msg: fmt.Sprintf("failed to set environment variable '%s': %s", k, err.Error()),
			}
			log.Error().Err(e.InternalError()).Msg(e.Error())
			return e
		}
	}

	// load any settings from config file
	configFile := ""
	if _, ok := c.visitedFlags["config-file"]; ok {
		configFile = c.configFile
	}
	if err := app.Config.Load(configFile); err != nil {
		return err
	}

	// configure the global logger
	cfg := app.Config.Global
	if cfg.EnableJSONLogging {
		stdoutLevelWriter := zerolog.NewFilteredLevelWriter([]zerolog.Level{
			zerolog.DebugLevel, zerolog.InfoLevel, zerolog.WarnLevel,
		}, os.Stdout)
		stderrLevelWriter := zerolog.NewFilteredLevelWriter([]zerolog.Level{
			zerolog.ErrorLevel, zerolog.FatalLevel, zerolog.PanicLevel,
		}, os.Stderr)
		writer := zerolog.MultiLevelWriter(stdoutLevelWriter, stderrLevelWriter)
		l := zerolog.New(writer).With().Timestamp().Logger()
		l.SetLevel(cfg.LogLevel)
		log.ReplaceGlobal(l)
	} else {
		log.SetLevel(cfg.LogLevel)
	}
	log.Debug().Msgf("global settings: %+v", cfg)
	return nil
}
