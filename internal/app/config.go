package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/josh-hogle/okta-openvpn-auth-plugin/internal/errors"
	"github.com/josh-hogle/zerolog/v2/log"
	"github.com/spf13/viper"
)

// Default configuration settings.
const (
	ConfigDir = "/opt/okta-openvpn-auth-plugin/etc"
)

func init() {
	Config = &config{}

	// initialize default settings
	viper.SetDefault("auth.api_key_file", "")
	viper.SetDefault("auth.geoip_db_path", "")
	viper.SetDefault("auth.geoip_locale", DefaultGeoIPLocale)
	viper.SetDefault("auth.interactive", false)
	viper.SetDefault("auth.mfa_methods", []string{})
	viper.SetDefault("auth.mfa_timeout", DefaultMFATimeout)
	viper.SetDefault("auth.org_name", "")

	viper.SetDefault("global.log_level", DefaultLogLevel)
	viper.SetDefault("global.enable_json_logging", false)

	viper.SetDefault("version.short", false)
}

// config is an internal structure to hold the application configuration.
type config struct {
	// Auth stores auth command configuration options
	Auth AuthOptions `mapstructure:"auth"`

	// Global stores the global configuration options
	Global GlobalOptions `mapstructure:"global"`

	// Version stores version command configuration options
	Version VersionOptions `mapstructure:"version"`

	// unexported fields
	sourceFile string
}

// Load simply loads the configuration settings into memory.
//
// The config file is determined as follows:
//
//	◽ If the --config-file option is specified on the command-line, use that file.
//	◽ If the PLUGIN_CONFIG_FILE environment variable is set, use that file.
//	◽ Use the default /opt/okta-openvpn-auth-plugin/etc/config.yaml file if it exists.
//
// The following errors are returned by this function:
// ConfigLoadFailure, ConfigParseFailure, ConfigValidateFailure
func (c *config) Load(file string) error {
	// config file was specified on the command-line
	if file != "" {
		return c.loadFile(file)
	}

	// config file was specified via environment variable
	file = os.Getenv(fmt.Sprintf("%sCONFIG_FILE", EnvVarPrefix))
	if file != "" {
		return c.loadFile(file)
	}

	// use the default config file
	return c.loadDefaultFile()
}

// loadDefaultFile attempts to load a default configuration file from the user's configuration folder.
//
// The following errors are returned by this function:
// ConfigLoadFailure, ConfigParseFailure, ConfigValidateFailure
func (c *config) loadDefaultFile() error {
	// no specific config file was specified so we'll check for a default config file
	viper.AddConfigPath(ConfigDir)
	viper.SupportedExts = []string{"yaml", "yml"}
	viper.SetConfigName(DefaultConfigFile)

	// read the configuration file
	if err := viper.ReadInConfig(); err != nil {
		// since the default config file is being used but was not found, do not return an error
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return c.unmarshal()
		}
		e := &errors.ConfigLoadFailure{
			ConfigFile: viper.ConfigFileUsed(),
			Err:        err,
		}
		log.Error().Err(e.InternalError()).Str("config_file", e.ConfigFile).Msg(e.Error())
		return e
	}

	c.sourceFile = viper.ConfigFileUsed()
	return c.unmarshal()
}

// loadFile loads the specified configuration file.
//
// The following errors are returned by this function:
// ConfigLoadFailure, ConfigParseFailure, ConfigValidateFailure
func (c *config) loadFile(file string) error {
	file, err := filepath.Abs(os.ExpandEnv(file))
	if err != nil {
		e := &errors.ConfigLoadFailure{
			ConfigFile: file,
			Err:        err,
		}
		log.Error().Err(e.InternalError()).Str("config_file", e.ConfigFile).Msg(e.Error())
		return e
	}

	// read the configuration file
	viper.SetConfigFile(file)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			e := &errors.ConfigLoadFailure{
				ConfigFile: file,
				Err:        fmt.Errorf("file not found"),
			}
			log.Error().Err(e.InternalError()).Str("config_file", e.ConfigFile).Msg(e.Error())
			return e
		}
		e := &errors.ConfigLoadFailure{
			ConfigFile: viper.ConfigFileUsed(),
			Err:        err,
		}
		log.Error().Err(e.InternalError()).Str("config_file", e.ConfigFile).Msg(e.Error())
		return e
	}
	c.sourceFile = viper.ConfigFileUsed()

	// save the absolute path to the directory in which the config file is located
	absFile, err := filepath.Abs(c.sourceFile)
	if err != nil {
		e := &errors.ConfigLoadFailure{
			ConfigFile: c.sourceFile,
			Err:        err,
		}
		log.Error().Err(e.InternalError()).Str("config_file", e.ConfigFile).Msg(e.Error())
		return e
	}
	c.Global.ConfigDir = filepath.Dir(absFile)

	return c.unmarshal()
}

// unmarshal simply unmarshals the data from the config file into the object.
//
// The following errors are returned by this function:
// ConfigParseFailure, ConfigValidateFailure
func (c *config) unmarshal() error {
	if err := viper.Unmarshal(c); err != nil {
		e := &errors.ConfigParseFailure{
			ConfigFile: c.sourceFile,
			Err:        err,
		}
		log.Error().Err(e.InternalError()).Str("config_file", e.ConfigFile).Msg(e.Error())
		return e
	}
	return c.Global.Validate()
}
