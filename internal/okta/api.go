package okta

import (
	"encoding/json"
)

// API constants.
const (
	AuthExceptionCode            = "E000004"
	OktaAPIBaseURL               = "https://%s.okta.com/api/v1"
	PasswordExpiredExceptionCode = "E000064"
	PasswordExpiredSummary       = "Password is expired and must be changed."
)

// EmbeddedResource contains embedded resource information.
type EmbeddedResource struct {
	User    UserObject     `json:"user"`
	Factors []FactorObject `json:"factors"`
	Policy  PolicyObject   `json:"policy"`
}

// ErrorResponse contains error information when a REST authentication request fails.
type ErrorResponse struct {
	ErrorCode    string          `json:"errorCode"`
	ErrorSummary string          `json:"errorSummary"`
	ErrorLink    string          `json:"errorLink"`
	ErrorID      string          `json:"errorId"`
	ErrorCauses  []ErrorResponse `json:"errorCauses"`
}

// FactorObject contains information about a particular MFA factor.
type FactorObject struct {
	ID         string                  `json:"id"`
	FactorType string                  `json:"factorType"`
	Provider   string                  `json:"provider"`
	VendorName string                  `json:"vendorsName"`
	Profile    json.RawMessage         `json:"profile"`
	Links      map[string]LinkResource `json:"_links"`
}

// LinkResource describes links to other resources or API calls.
type LinkResource struct {
	Href  string              `json:"href"`
	Hints map[string][]string `json:"hints"`
}

// PolicyObject contains  policy information.
type PolicyObject struct {
	AllowRememberDevice             bool            `json:"allowRememberDevice"`
	RememberDeviceLifetimeInMinutes uint32          `json:"rememberDeviceLifetimeInMinutes"`
	RememberDeviceByDefault         bool            `json:"rememberDeviceByDefault"`
	FactorsPolicyInfo               json.RawMessage `json:"factorsPolicyInfo"`
}

// PrimaryAuthResponse contains primary authentication information when authentication succeeds.
type PrimaryAuthResponse struct {
	StateToken string                  `json:"stateToken"`
	ExpiresAt  string                  `json:"expiresAt"`
	Status     string                  `json:"status"`
	Embedded   EmbeddedResource        `json:"_embedded"`
	Links      map[string]LinkResource `json:"_links"`
}

// SecondaryAuthResponse contains secondary authentication information when MFA succeeds.
type SecondaryAuthResponse struct {
	ExpiresAt    string `json:"expiresAt"`
	Status       string `json:"status"`
	FactorResult string `json:"factorResult"`
	SessionToken string `json:"sessionToken"`
}

// UserObject holds information about a user.
type UserObject struct {
	ID              string      `json:"id"`
	PasswordChanged string      `json:"passwordChanged"`
	Profile         UserProfile `json:"profile"`
}

// UserProfile holds user profile information such as login, first and last name, etc.
type UserProfile struct {
	Login     string `json:"login"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Locale    string `json:"locale"`
	TimeZone  string `json:"timeZone"`
}
