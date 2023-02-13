package errors

import "fmt"

// GeoIPDatabaseFailure occurs when an error is detected while querying the GeoIP database.
type GeoIPDatabaseFailure struct {
	DatabaseFile string
	Err          error
}

// InternalError returns the internal error object.
func (e *GeoIPDatabaseFailure) InternalError() error {
	return e.Err
}

// Error returns the string version of the error.
func (e *GeoIPDatabaseFailure) Error() string {
	return fmt.Sprintf("error while querying the GeoIP database '%s': %s", e.DatabaseFile, e.Err.Error())
}

// Code returns the corresponding error code.
func (e *GeoIPDatabaseFailure) Code() int {
	return GeoIPDatabaseFailureCode
}

// GeoIPLookupFailure occurs when an error is detected while looking up an IP in the GeoIP database.
type GeoIPLookupFailure struct {
	ClientIP string
	Err      error
}

// InternalError returns the internal error object.
func (e *GeoIPLookupFailure) InternalError() error {
	return e.Err
}

// Error returns the string version of the error.
func (e *GeoIPLookupFailure) Error() string {
	return fmt.Sprintf("error while retrieving data for client IP '%s': %s", e.ClientIP, e.Err.Error())
}

// Code returns the corresponding error code.
func (e *GeoIPLookupFailure) Code() int {
	return GeoIPLookupFailureCode
}
