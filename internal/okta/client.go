package okta

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/josh-hogle/okta-openvpn-auth-plugin/internal/app"
	"github.com/josh-hogle/okta-openvpn-auth-plugin/internal/errors"
	"github.com/josh-hogle/okta-openvpn-auth-plugin/internal/util"
	"go.innotegrity.dev/zerolog/log"
	"gopkg.in/resty.v1"
)

// Client is a client for making Okta API requests.
type Client struct{}

// NewClient returns a new Client object.
func NewClient() *Client {
	return &Client{}
}

// Authenticate attempts to authenticate the user credentials in the client request using the Okta API.
//
// The following errors are returned by this function:
// OktaRequestFailure, OktaResponseFailure, OktaAuthFailure
func (c *Client) Authenticate(req *util.OpenVPNClientRequest) error {
	config := app.Config.Auth
	logger := log.With().
		Str("username", req.Username).
		Str("ip", req.ClientIP).
		Str("location", req.Location).
		Logger()

	// check the password and parse out TOTP if enabled
	totp := ""
	if (config.MFAMethods & app.MFATOTP) > 0 {
		regex := regexp.MustCompile(`(?i)(.*?)(\+([0-9]{6}|push))?$`)
		matches := regex.FindStringSubmatch(req.Password)
		if matches != nil {
			req.Password = matches[1]
			totp = matches[3]
		}
	}

	// perform authentication via Okta API
	body := map[string]interface{}{
		"username": req.Username,
		"password": req.Password,
		"options": map[string]interface{}{
			"warnBeforePasswordExpired": true,
		},
	}
	resp, err := c.postRequest(fmt.Sprintf("%s/authn", fmt.Sprintf(OktaAPIBaseURL, config.OrgName)), body)
	if err != nil {
		return err
	}
	if logger.IsDebugEnabled() {
		fullResponse := spew.Sdump(resp)
		logger.Debug().Str("response", fullResponse).Msgf("auth response: %s", fullResponse)
	}

	// if HTTP status is not HTTP 200, log the error summary and code and fail
	if resp.StatusCode() != 200 {
		var r ErrorResponse
		err := json.Unmarshal(resp.Body(), &r)
		if err != nil {
			e := &errors.OktaResponseFailure{
				Err: err,
			}
			logger.Error().Err(e.InternalError()).Msg(e.Error())
			return e
		}
		e := &errors.OktaAuthFailure{
			Username:     req.Username,
			ErrorCode:    r.ErrorCode,
			ErrorSummary: r.ErrorSummary,
		}
		logger.Error().Err(e.InternalError()).Str("error_code", e.ErrorCode).Str("error_summary", e.ErrorSummary).
			Msg(e.Error())
		return e
	}

	// parse the response into an object
	var pr PrimaryAuthResponse
	err = json.Unmarshal(resp.Body(), &pr)
	if err != nil {
		e := &errors.OktaResponseFailure{
			Err: err,
		}
		logger.Error().Err(e.InternalError()).Msg(e.Error())
		return e
	}
	switch pr.Status {
	case "SUCCESS":
		logger.Info().Msgf("authentication succeeded for '%s' (No MFA required)", req.Username)
		return nil

	case "PASSWORD_WARN":
		logger.Warn().Msgf("password for '%s' is about to expire", req.Username)
		logger.Info().Msgf("authentication succeeded for '%s' (No MFA required)", req.Username)
		return nil

	case "PASSWORD_EXPIRED":
		e := &errors.OktaAuthFailure{
			Username:     req.Username,
			ErrorCode:    PasswordExpiredExceptionCode,
			ErrorSummary: PasswordExpiredSummary,
		}
		logger.Error().Err(e.InternalError()).Str("error_code", e.ErrorCode).Str("error_summary", e.ErrorSummary).
			Msg(e.Error())
		return e

	case "MFA_REQUIRED":
		logger.Info().Msg("primary authentication succeeded")

		// perform MFA validation
		if totp != "" {
			if (config.MFAMethods&app.MFAPush) > 0 && strings.EqualFold(totp, "push") {
				return c.validatePush(req, pr)
			} else if (config.MFAMethods & app.MFATOTP) > 0 {
				return c.validateTOTP(req, totp, pr)
			}
		}

		// no supported MFA methods available
		e := &errors.OktaAuthFailure{
			Username:     req.Username,
			ErrorCode:    AuthExceptionCode,
			ErrorSummary: "MFA is required but no supported methods are available.",
		}
		logger.Error().Err(e.InternalError()).Str("error_code", e.ErrorCode).Str("error_summary", e.ErrorSummary).
			Msg(e.Error())
		return e
	}

	// authentication failed
	e := &errors.OktaAuthFailure{
		Username:     req.Username,
		ErrorCode:    AuthExceptionCode,
		ErrorSummary: fmt.Sprintf("status returned was '%s'", pr.Status),
	}
	logger.Error().Err(e.InternalError()).Str("error_code", e.ErrorCode).Str("error_summary", e.ErrorSummary).
		Msg(e.Error())
	return e
}

// getFactorVerifyLink retrieves the verify link for the given MFA factor from the response
func (c *Client) getFactorVerifyLink(factorType string, pr PrimaryAuthResponse) (string, error) {
	for _, factor := range pr.Embedded.Factors {
		if factor.FactorType == factorType {
			verifyLink, ok := factor.Links["verify"]
			if !ok {
				return "", fmt.Errorf("'%s': MFA factor has no verification link", factorType)
			}
			return verifyLink.Href, nil
		}
	}
	return "", fmt.Errorf("'%s': not a valid MFA factor type", factorType)
}

// postRequest performs a POST request rendering the given map to a JSON object
//
// The following errors are returned by this function:
// OktaRequestFailure
func (c *Client) postRequest(url string, body map[string]interface{}) (*resty.Response, error) {
	config := app.Config.Auth
	logger := log.With().
		Str("url", url).
		Logger()
	if logger.IsDebugEnabled() {
		logger = logger.With().Interface("body", body).Logger()
	}

	// Marshal the body into a JSON object
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// Make the request
	request := resty.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json")
	if config.APIKey != "" {
		request = request.SetHeader("Authorization", fmt.Sprintf("SSWS %s", config.APIKey))
	}
	resp, err := request.SetBody(jsonBody).Post(url)
	if err != nil {
		e := &errors.OktaRequestFailure{
			Err: err,
		}
		logger.Error().Err(e.InternalError()).Msg(e.Error())
		return nil, e
	}
	return resp, nil
}

// validatePush performs an MFA PUSH validation for the user
//
// The following errors are returned by this function:
// OktaRequestFailure, OktaResponseFailure, OktaAuthFailure
func (c *Client) validatePush(req *util.OpenVPNClientRequest, pr PrimaryAuthResponse) error {
	config := app.Config.Auth
	logger := log.With().
		Str("mfa_method", "push").
		Logger()

	// get the verification link
	link, err := c.getFactorVerifyLink("push", pr)
	if err != nil {
		e := &errors.OktaRequestFailure{
			Err: err,
		}
		logger.Error().Err(e.InternalError()).Msg(e.Error())
		return e
	}

	// poll until we see a response
	timeout := int(config.MFATimeout.Seconds())
	for i := 1; i <= timeout; i++ {
		if (i % 5) == 0 {
			logger.Info().Msgf("still waiting on MFA reply after %v seconds", i)
		}

		// POST the verification request
		resp, err := c.postRequest(link, map[string]interface{}{
			"stateToken": pr.StateToken,
		})
		if err != nil {
			return err
		}
		if logger.IsDebugEnabled() {
			fullResponse := spew.Sdump(resp)
			logger.Debug().Str("response", fullResponse).Msgf("MFA auth response: %s", fullResponse)
		}

		// error occurred
		if resp.StatusCode() != 200 {
			var r ErrorResponse
			err := json.Unmarshal(resp.Body(), &r)
			if err != nil {
				e := &errors.OktaResponseFailure{
					Err: err,
				}
				logger.Error().Err(e.InternalError()).Msg(e.Error())
				return e
			}
			e := &errors.OktaAuthFailure{
				Username:     req.Username,
				ErrorCode:    r.ErrorCode,
				ErrorSummary: r.ErrorSummary,
			}
			logger.Error().Err(e.InternalError()).Str("error_code", e.ErrorCode).Str("error_summary", e.ErrorSummary).
				Msg(e.Error())
			return e
		}

		// check the status
		var sr SecondaryAuthResponse
		err = json.Unmarshal(resp.Body(), &sr)
		if err != nil {
			e := &errors.OktaResponseFailure{
				Err: err,
			}
			logger.Error().Err(e.InternalError()).Msg(e.Error())
			return e
		}
		switch sr.Status {
		case "SUCCESS":
			logger.Info().Msg("MFA authentication succeeded")
			return nil
		case "MFA_CHALLENGE":
			if sr.FactorResult == "REJECTED" {
				e := &errors.OktaAuthFailure{
					Username:     req.Username,
					ErrorCode:    AuthExceptionCode,
					ErrorSummary: "user declined MFA request",
				}
				logger.Error().Err(e.InternalError()).Str("error_code", e.ErrorCode).Str("error_summary", e.ErrorSummary).
					Msg(e.Error())
				return e
			}
		default:
			e := &errors.OktaAuthFailure{
				Username:     req.Username,
				ErrorCode:    AuthExceptionCode,
				ErrorSummary: fmt.Sprintf("MFA authentication failed: status returned was '%s'", sr.Status),
			}
			logger.Error().Err(e.InternalError()).Str("error_code", e.ErrorCode).Str("error_summary", e.ErrorSummary).
				Str("status", sr.Status).Msg(e.Error())
			return e
		}
		time.Sleep(time.Second)
	}

	// timed out
	e := &errors.OktaAuthFailure{
		Username:     req.Username,
		ErrorCode:    AuthExceptionCode,
		ErrorSummary: "timed out waiting for reply to PUSH request",
	}
	logger.Error().Err(e.InternalError()).Str("error_code", e.ErrorCode).Str("error_summary", e.ErrorSummary).
		Msg(e.Error())
	return e
}

// validateTOTP performs an MFA TOTP validation for the user
//
// The following errors are returned by this function:
// OktaRequestFailure, OktaResponseFailure, OktaAuthFailure
func (c *Client) validateTOTP(req *util.OpenVPNClientRequest, totp string, pr PrimaryAuthResponse) error {
	logger := log.With().
		Str("mfa_method", "totp").
		Logger()

	// get the verification link
	link, err := c.getFactorVerifyLink("token:software:totp", pr)
	if err != nil {
		e := &errors.OktaRequestFailure{
			Err: err,
		}
		logger.Error().Err(e.InternalError()).Msg(e.Error())
		return e
	}

	// POST the verification request
	resp, err := c.postRequest(link, map[string]interface{}{
		"stateToken": pr.StateToken,
	})
	if err != nil {
		return err
	}
	if logger.IsDebugEnabled() {
		fullResponse := spew.Sdump(resp)
		logger.Debug().Str("response", fullResponse).Msgf("MFA auth response: %s", fullResponse)
	}

	// error occurred
	if resp.StatusCode() != 200 {
		var r ErrorResponse
		err := json.Unmarshal(resp.Body(), &r)
		if err != nil {
			e := &errors.OktaResponseFailure{
				Err: err,
			}
			logger.Error().Err(e.InternalError()).Msg(e.Error())
			return e
		}
		e := &errors.OktaAuthFailure{
			Username:     req.Username,
			ErrorCode:    r.ErrorCode,
			ErrorSummary: r.ErrorSummary,
		}
		logger.Error().Err(e.InternalError()).Str("error_code", e.ErrorCode).Str("error_summary", e.ErrorSummary).
			Msg(e.Error())
		return e
	}

	// check the status
	var sr SecondaryAuthResponse
	err = json.Unmarshal(resp.Body(), &sr)
	if err != nil {
		e := &errors.OktaResponseFailure{
			Err: err,
		}
		logger.Error().Err(e.InternalError()).Msg(e.Error())
		return e
	}
	switch sr.Status {
	case "SUCCESS":
		logger.Info().Msg("MFA authentication succeeded")
		return nil
	}

	// MFA authentication failed
	e := &errors.OktaAuthFailure{
		Username:     req.Username,
		ErrorCode:    AuthExceptionCode,
		ErrorSummary: fmt.Sprintf("MFA authentication failed: status returned was '%s'", sr.Status),
	}
	logger.Error().Err(e.InternalError()).Str("error_code", e.ErrorCode).Str("error_summary", e.ErrorSummary).
		Str("status", sr.Status).Msg(e.Error())
	return e
}
