package version

import (
	"fmt"

	"github.com/josh-hogle/okta-openvpn-auth-plugin/internal/app"
	"github.com/josh-hogle/zerolog/v2"
	"github.com/josh-hogle/zerolog/v2/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Command is the object for executing the actual command.
type Command struct {
	cobra.Command

	// unexported variables
	visitedFlags map[string]bool
}

// NewCommand creates a new Command object.
func NewCommand() *Command {
	cmd := &Command{
		visitedFlags: map[string]bool{},
	}
	cmd.Use = "version"
	cmd.Short = "Display application version information"
	cmd.Long = "This command allows you to see simple or detailed version information."
	cmd.Run = cmd.run
	// we are not relying on Persistent*RunE functions here so we can better control the order in which functions
	// are called by sub-commands
	cmd.PostRunE = cmd.postRunE
	cmd.PreRunE = cmd.preRunE
	cmd.SilenceErrors = true
	flags := cmd.Flags()

	// command-line flags
	flags.BoolP("short", "s", false, "show version only")
	viper.BindPFlag("version.short", flags.Lookup("short"))
	viper.BindEnv("version.short", fmt.Sprintf("%sVERSION_SHORT", app.EnvVarPrefix))

	return cmd
}

// run simply executes the command.
func (c *Command) run(cmd *cobra.Command, args []string) {
	config := app.Config.Version

	// disable logger
	log.SetLevel(zerolog.Disabled)

	// show just the version
	if config.Short {
		fmt.Printf("%s\n", app.Version)
		return
	}

	// show version and build
	fmt.Printf("%s build %s", app.Version, app.Build)
	if app.DevBuild {
		fmt.Printf(" [Developer Build]")
	}
	fmt.Printf("\n")
}

// postRunE is called after the command is executed.
func (c *Command) postRunE(cmd *cobra.Command, args []string) error {
	return nil
}

// preRunE is called before the command is executed.
func (c *Command) preRunE(cmd *cobra.Command, args []string) error {
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

	// validate options
	if err := app.Config.Version.Validate(); err != nil {
		return err
	}
	log.Debug().Msgf("'version' command settings: %+v", app.Config.Version)
	return nil
}
