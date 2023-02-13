package errors

import "fmt"

// ConfigLoadFailure occurs when an error is detected while loading the configuration file.
type ConfigLoadFailure struct {
	ConfigFile string
	Err        error
}

// InternalError returns the internal error object.
func (e *ConfigLoadFailure) InternalError() error {
	return e.Err
}

// Error returns the string version of the error.
func (e *ConfigLoadFailure) Error() string {
	return fmt.Sprintf("error while loading configuration file '%s': %s", e.ConfigFile, e.Err.Error())
}

// Code returns the corresponding error code.
func (e *ConfigLoadFailure) Code() int {
	return ConfigLoadFailureCode
}

// ConfigParseFailure occurs when an error is detected while parsing configuration settings.
type ConfigParseFailure struct {
	ConfigFile string
	Err        error
}

// InternalError returns the internal error object.
func (e *ConfigParseFailure) InternalError() error {
	return e.Err
}

// Error returns the string version of the error.
func (e *ConfigParseFailure) Error() string {
	return fmt.Sprintf("error while parsing configuration file '%s': %s", e.ConfigFile, e.Err.Error())
}

// Code returns the corresponding error code.
func (e *ConfigParseFailure) Code() int {
	return ConfigParseFailureCode
}

// ConfigValidateFailure occurs when an error is detected while validating configuration settings.
type ConfigValidateFailure struct {
	Setting string
	Value   interface{}
	Err     error
}

// InternalError returns the internal error object.
func (e *ConfigValidateFailure) InternalError() error {
	return e.Err
}

// Error returns the string version of the error.
func (e *ConfigValidateFailure) Error() string {
	if e.Setting != "" {
		return fmt.Sprintf("error while validating configuration setting '%s': %s", e.Setting, e.Err.Error())
	}
	return fmt.Sprintf("error while validating one or more configuration settings: %s", e.Err.Error())
}

// Code returns the corresponding error code.
func (e *ConfigValidateFailure) Code() int {
	return ConfigValidateFailureCode
}
