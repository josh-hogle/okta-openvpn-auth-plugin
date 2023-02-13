package auth

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/josh-hogle/okta-openvpn-auth-plugin/internal/app"
	"github.com/josh-hogle/okta-openvpn-auth-plugin/internal/errors"
	"github.com/josh-hogle/okta-openvpn-auth-plugin/internal/okta"
	"github.com/josh-hogle/okta-openvpn-auth-plugin/internal/util"
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
	cmd.Use = "auth"
	cmd.Short = "Perform user authentication via Okta."
	cmd.Long = "This command performs the actual authentication (username+password) with optional MFA verification via Okta."
	cmd.RunE = cmd.runE
	// we are not relying on Persistent*RunE functions here so we can better control the order in which functions
	// are called by sub-commands
	cmd.PostRunE = cmd.postRunE
	cmd.PreRunE = cmd.preRunE
	cmd.SilenceErrors = true
	flags := cmd.Flags()

	// flags stored by viper
	flags.String("api-key-file", "", "File containing Okta API key")
	viper.BindPFlag("auth.api_key_file", flags.Lookup("api-key-file"))
	viper.BindEnv("auth.api_key_file", fmt.Sprintf("%sAUTH_API_KEY_FILE", app.EnvVarPrefix))

	flags.String("geoip-db-path", "", "Path to MaxMind GeoIP database files")
	viper.BindPFlag("auth.geoip_db_path", flags.Lookup("geoip-db-path"))
	viper.BindEnv("auth.geoip_db_path", fmt.Sprintf("%sAUTH_GEOIP_DB_PATH", app.EnvVarPrefix))

	flags.String("geoip-locale", app.DefaultGeoIPLocale, "Locale to load from GeoIP database")
	viper.BindPFlag("auth.geoip_locale", flags.Lookup("geoip-locale"))
	viper.BindEnv("auth.geoip_locale", fmt.Sprintf("%sAUTH_GEOIP_LOCALE", app.EnvVarPrefix))

	flags.Bool("interactive", false, "Perform an interactive authentication test to verify configuration settings")
	viper.BindPFlag("auth.interactive", flags.Lookup("interactive"))
	viper.BindEnv("auth.interactive", fmt.Sprintf("%sAUTH_INTERACTIVE", app.EnvVarPrefix))

	flags.StringArray("mfa-methods", []string{}, "Authorized methods for MFA verification")
	viper.BindPFlag("auth.mfa_methods", flags.Lookup("mfa-methods"))
	viper.BindEnv("auth.mfa_methods", fmt.Sprintf("%sAUTH_MFA_METHODS", app.EnvVarPrefix))

	flags.String("mfa-timeout", app.DefaultMFATimeout, "Maximum time to wait for entering MFA code")
	viper.BindPFlag("auth.mfa_timeout", flags.Lookup("mfa-timeout"))
	viper.BindEnv("auth.mfa_timeout", fmt.Sprintf("%sAUTH_MFA_TIMEOUT", app.EnvVarPrefix))

	flags.String("org-name", "", "Okta organization name")
	viper.BindPFlag("auth.org_name", flags.Lookup("org-name"))
	viper.BindEnv("auth.org_name", fmt.Sprintf("%sAUTH_ORG_NAME", app.EnvVarPrefix))

	return cmd
}

// runE simply executes the command.
func (c *Command) runE(cmd *cobra.Command, args []string) error {
	config := app.Config.Auth

	// perform interactive authentication test
	if config.Interactive {
		return c.doInteractiveAuth()
	}

	// authenticate the user
	data := "1"
	req := util.NewOpenVPNClientRequest()
	client := okta.NewClient()
	err := client.Authenticate(req)
	if err != nil {
		data = "0"
	}

	// write the status
	controlFile := os.Getenv("auth_control_file")
	writeErr := ioutil.WriteFile(controlFile, []byte(data), 0644)
	if writeErr != nil {
		e := &errors.GeneralFailure{
			Err: writeErr,
			Msg: fmt.Sprintf("failed to write control file '%s': %s", controlFile, writeErr.Error()),
		}
		log.Error().Err(e.InternalError()).Str("control_file", controlFile).Msg(e.Error())
		return e
	}
	return err
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
	if err := app.Config.Auth.Validate(); err != nil {
		return err
	}
	log.Debug().Msgf("'auth' command settings: %+v", app.Config.Auth)
	return nil
}

// doInteractiveAuth performs an interactive authentication and is used for testing configuration settings to make
// make sure they work.
func (c *Command) doInteractiveAuth() error {
	// prompt for credentials
	r := bufio.NewReader(os.Stdin)
	//w := bufio.NewWriter(os.Stdout)
	//rw := bufio.NewReadWriter(r, w)
	//term := terminal.NewTerminal(rw, "")

	fmt.Printf("Username: ")
	username, err := r.ReadString('\n')
	if err != nil {
		e := &errors.GeneralFailure{
			Err: err,
			Msg: fmt.Sprintf("failed to read username from terminal: %s", err.Error()),
		}
		log.Error().Err(e.InternalError()).Msg(e.Error())
		return e
	}
	fmt.Printf("Password: ")
	//	password, err := term.ReadPassword("Password: ")
	password, err := r.ReadString('\n')
	if err != nil {
		e := &errors.GeneralFailure{
			Err: err,
			Msg: fmt.Sprintf("failed to read password from terminal: %s", err.Error()),
		}
		log.Error().Err(e.InternalError()).Msg(e.Error())
		return e
	}

	// get public IP address
	var clientIP string
	if r, err := http.Get("http://ip-api.com/json/"); err == nil {
		defer r.Body.Close()
		if body, err := ioutil.ReadAll(r.Body); err == nil {
			type IP struct {
				Query string
			}
			var ip IP
			if err := json.Unmarshal(body, &ip); err == nil {
				clientIP = ip.Query
			}
		}
	}

	// setup environment to replicate OpenVPN request
	if err := os.Setenv("username", strings.TrimSpace(username)); err != nil {
		e := &errors.GeneralFailure{
			Err: err,
			Msg: fmt.Sprintf("failed to set environment variable 'username': %s", err.Error()),
		}
		log.Error().Err(e.InternalError()).Msg(e.Error())
		return e
	}
	if err := os.Setenv("password", strings.TrimSpace(password)); err != nil {
		e := &errors.GeneralFailure{
			Err: err,
			Msg: fmt.Sprintf("failed to set environment variable 'password': %s", err.Error()),
		}
		log.Error().Err(e.InternalError()).Msg(e.Error())
		return e
	}
	if err := os.Setenv("untrusted_ip", clientIP); err != nil {
		e := &errors.GeneralFailure{
			Err: err,
			Msg: fmt.Sprintf("failed to set environment variable 'untrusted_ip': %s", err.Error()),
		}
		log.Error().Err(e.InternalError()).Msg(e.Error())
		return e
	}

	// perform the authentication
	req := util.NewOpenVPNClientRequest()
	client := okta.NewClient()
	return client.Authenticate(req)
}
