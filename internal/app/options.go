package app

import (
	"encoding/base64"
	goerrors "errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/josh-hogle/okta-openvpn-auth-plugin/internal/errors"
	"github.com/oschwald/geoip2-golang"
	"go.innotegrity.dev/zerolog"
	"go.innotegrity.dev/zerolog/log"
)

// Default configuration settings.
const (
	DefaultConfigFile  = "config"
	DefaultGeoIPLocale = "en"
	DefaultLogLevel    = "info"
	DefaultMFATimeout  = "30s"

	MinMFATimeout = 15
)

// Supported MFA methods
const (
	MFANone = 0
	MFATOTP = 1
	MFAPush = 2
)

var mfaMethodStrings = map[string]uint8{
	"none": MFANone,
	"totp": MFATOTP,
	"push": MFAPush,
}

// AuthOptions holds the options for the auth command.
type AuthOptions struct {
	// APIKeyFile holds the path to the Okta API key.
	APIKeyFile string `mapstructure:"api_key_file"`

	// APIKey holds the actual API key read from the API key file.
	APIKey string

	// GeoIPDBPath holds the path to the GeoIP data files.
	GeoIPDBPath string `mapstructure:"geoip_db_path"`

	// GeoIPLocale holds the locale to use for retrieving GeoIP data.
	GeoIPLocale string `mapstructure:"geoip_locale"`

	// Interactive determines whether or not to perform an interactive authentication.
	Interactive bool `mapstructure:"interactive"`

	// MFAMethods holds a bitmask for the allowed methods for MFA.
	MFAMethods uint8

	// MFATimeout holds the length of time to wait for a user to respond to an MFA request before timing out.
	MFATimeout time.Duration

	// OrgName holds the name of the Okta organization.
	OrgName string `mapstructure:"org_name"`

	// RawMFAMethods holds the list of unvalidated MFA methods.
	RawMFAMethods []string `mapstructure:"mfa_methods"`

	// RawMFATimeout holds the unparsed duration of how long to wait for a user to respond to an MFA request
	// before timing out.
	RawMFATimeout string `mapstructure:"mfa_timeout"`
}

// Validate checks and saves any configuration settings from viper and ensures that all values are sane.
//
// The following errors are returned by this function:
// ConfigValidateFailure
func (o *AuthOptions) Validate() error {
	// validate organization
	if err := requireSetting(o.OrgName, "auth.org_name"); err != nil {
		return err
	}

	// read the API key, if present
	if o.APIKeyFile != "" {
		setting := "auth.api_key_file"
		absPath, err := filepath.Abs(o.APIKeyFile)
		if err != nil {
			e := &errors.ConfigValidateFailure{
				Setting: setting,
				Value:   o.APIKeyFile,
				Err:     err,
			}
			log.Error().Err(e.InternalError()).Str("setting", e.Setting).Interface("value", e.Value).Msg(e.Error())
			return e
		}

		key, err := ioutil.ReadFile(absPath)
		if err != nil {
			e := &errors.ConfigValidateFailure{
				Setting: setting,
				Value:   absPath,
				Err:     fmt.Errorf("error reading the API key file: %s", err.Error()),
			}
			log.Error().Err(e.InternalError()).Str("setting", e.Setting).Interface("value", e.Value).Msg(e.Error())
			return e
		}

		decodedKey, err := base64.StdEncoding.DecodeString(string(key))
		if err != nil {
			e := &errors.ConfigValidateFailure{
				Setting: setting,
				Value:   absPath,
				Err:     fmt.Errorf("error decoding the API key: %s", err.Error()),
			}
			log.Error().Err(e.InternalError()).Str("setting", e.Setting).Interface("value", e.Value).Msg(e.Error())
			return e
		}
		o.APIKey = strings.TrimSpace(string(decodedKey))
	}

	// test opening the GeoIP database
	if o.GeoIPDBPath != "" {
		setting := "auth.geoip_db_path"
		absPath, err := filepath.Abs(o.GeoIPDBPath)
		if err != nil {
			e := &errors.ConfigValidateFailure{
				Setting: setting,
				Value:   o.GeoIPDBPath,
				Err:     err,
			}
			log.Error().Err(e.InternalError()).Str("setting", e.Setting).Interface("value", e.Value).Msg(e.Error())
			return e
		}

		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			e := &errors.ConfigValidateFailure{
				Setting: setting,
				Value:   o.GeoIPDBPath,
				Err:     err,
			}
			log.Error().Err(e.InternalError()).Str("setting", e.Setting).Interface("value", e.Value).Msg(e.Error())
			return e
		}

		db, err := geoip2.Open(absPath)
		if err != nil {
			e := &errors.ConfigValidateFailure{
				Setting: setting,
				Value:   o.GeoIPDBPath,
				Err:     fmt.Errorf("error opening the GeoIP database: %s", err.Error()),
			}
			log.Error().Err(e.InternalError()).Str("setting", e.Setting).Interface("value", e.Value).Msg(e.Error())
			return e
		}
		db.Close()
		o.GeoIPDBPath = absPath
	}

	// validate MFA methods
	o.MFAMethods = 0
	for _, method := range o.RawMFAMethods {
		m, err := parseMFAMethod(method)
		if err != nil {
			e := &errors.ConfigValidateFailure{
				Setting: "auth.mfa_methods",
				Value:   method,
				Err:     err,
			}
			log.Error().Err(e.InternalError()).Str("setting", e.Setting).Interface("value", e.Value).Msg(e.Error())
			return e
		}
		o.MFAMethods = o.MFAMethods | m
	}

	// validate MFA timeout
	setting := "auth.mfa_timeout"
	duration, err := time.ParseDuration(o.RawMFATimeout)
	if err != nil {
		e := &errors.ConfigValidateFailure{
			Setting: setting,
			Value:   o.RawMFATimeout,
			Err:     err,
		}
		log.Error().Err(e.InternalError()).Str("setting", e.Setting).Interface("value", e.Value).Msg(e.Error())
		return e
	}
	sec := duration.Seconds()
	if sec < MinMFATimeout {
		log.Warn().Str("setting", setting).Interface("value", duration).
			Msgf("MFA timeout of %v second(s) is less than the minimum threshold; defaulting to %ds", sec, MinMFATimeout)
		o.MFATimeout = MinMFATimeout * time.Second
	} else {
		o.MFATimeout = duration
	}

	return nil
}

// GlobalOptions holds the global configuration settings.
type GlobalOptions struct {
	// ConfigDir is the directory in which the configuration file is located.
	ConfigDir string

	// EnableJSONLogging is flag which determines whether or not to log output as JSON instead of text.
	EnableJSONLogging bool `mapstructure:"enable_json_logging"`

	// LogLevel holds the the minimum level of events to log.
	LogLevel zerolog.Level

	// RawLogLevel holds the the minimum level of events to log as a string.
	RawLogLevel string `mapstructure:"log_level"`
}

// Validate checks and saves any configuration settings from viper and ensures that all values are sane.
//
// The following errors are returned by this function:
// ConfigValidateFailure
func (o *GlobalOptions) Validate() error {
	// parse string log level to actual log level
	level, err := zerolog.ParseLevel(o.RawLogLevel)
	if err != nil {
		e := &errors.ConfigValidateFailure{
			Setting: "global.log_level",
			Value:   o.RawLogLevel,
			Err:     err,
		}
		log.Error().Err(e.InternalError()).Str("setting", e.Setting).Interface("value", e.Value).Msg(e.Error())
		return e
	}
	o.LogLevel = level

	return nil
}

// VersionOptions holds specific settings for the version command.
type VersionOptions struct {
	// Short represents a flag used to determine whether to show just the version or not.
	Short bool `mapstructure:"short"`
}

// Validate checks and saves any configuration settings from viper and ensures that all values are sane.
//
// The following errors are returned by this function:
// ConfigValidateFailure
func (o *VersionOptions) Validate() error {
	return nil
}

// parseMFAMethod converts a MFA method string to an actual MFA method value
func parseMFAMethod(method string) (uint8, error) {
	normalized := strings.ToLower(method)
	if v, ok := mfaMethodStrings[normalized]; ok {
		return v, nil
	}
	return 0, fmt.Errorf("no such MFA method '%s'", method)
}

// requireSetting checks that the value is not empty and returns an error if it is.
func requireSetting(value, setting string) error {
	if value == "" {
		e := &errors.ConfigValidateFailure{
			Setting: setting,
			Value:   "",
			Err:     goerrors.New("setting is required and cannot be empty"),
		}
		log.Error().Err(e.InternalError()).Str("setting", e.Setting).Interface("value", e.Value).Msg(e.Error())
		return e
	}
	return nil
}
