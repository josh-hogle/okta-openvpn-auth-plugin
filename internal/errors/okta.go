package errors

import "fmt"

// OktaRequestFailure occurs when an error is detected while decoding a response from the Okta API.
type OktaRequestFailure struct {
	Err error
}

// InternalError returns the internal error object.
func (e *OktaRequestFailure) InternalError() error {
	return e.Err
}

// Error returns the string version of the error.
func (e *OktaRequestFailure) Error() string {
	return fmt.Sprintf("error while making Okta API request: %s", e.Err.Error())
}

// Code returns the corresponding error code.
func (e *OktaRequestFailure) Code() int {
	return OktaRequestFailureCode
}

// OktaResponseFailure occurs when an error is detected while decoding a response from the Okta API.
type OktaResponseFailure struct {
	Err error
}

// InternalError returns the internal error object.
func (e *OktaResponseFailure) InternalError() error {
	return e.Err
}

// Error returns the string version of the error.
func (e *OktaResponseFailure) Error() string {
	return fmt.Sprintf("failed to decode Okta API response: %s", e.Err.Error())
}

// Code returns the corresponding error code.
func (e *OktaResponseFailure) Code() int {
	return OktaResponseFailureCode
}

// OktaAuthFailure occurs when an authentication failure occurs.
type OktaAuthFailure struct {
	Username     string
	ErrorCode    string
	ErrorSummary string
}

// InternalError returns the internal error object.
func (e *OktaAuthFailure) InternalError() error {
	return fmt.Errorf("%s (%s)", e.ErrorSummary, e.ErrorCode)
}

// Error returns the string version of the error.
func (e *OktaAuthFailure) Error() string {
	return fmt.Sprintf("authentication failed for user '%s': %s (%s)", e.Username, e.ErrorSummary, e.ErrorCode)
}

// Code returns the corresponding error code.
func (e *OktaAuthFailure) Code() int {
	return OktaAuthFailureCode
}
