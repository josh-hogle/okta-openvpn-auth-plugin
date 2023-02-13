package errors

// Object error codes (0-1000)
const (
	// general errors (0-20)
	NoneCode           = 0
	UsageCode          = 1
	GeneralFailureCode = 2

	// configuration errors (21-40)
	ConfigLoadFailureCode     = 21
	ConfigParseFailureCode    = 22
	ConfigValidateFailureCode = 23

	// GeoIP errors (41-60)
	GeoIPDatabaseFailureCode = 41
	GeoIPLookupFailureCode   = 42

	// Okta errors (61-80)
	OktaRequestFailureCode  = 61
	OktaResponseFailureCode = 62
	OktaAuthFailureCode     = 63
)
