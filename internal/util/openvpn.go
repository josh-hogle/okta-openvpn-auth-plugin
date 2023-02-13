package util

import (
	"net"
	"os"
	"strings"

	"github.com/josh-hogle/okta-openvpn-auth-plugin/internal/app"
	"github.com/josh-hogle/okta-openvpn-auth-plugin/internal/errors"
	"github.com/josh-hogle/zerolog/v2/log"
	geoip2 "github.com/oschwald/geoip2-golang"
)

// OpenVPNClientRequest holds data from the OpenVPN connection request
type OpenVPNClientRequest struct {
	// ClientIP holds the client's untrusted IP address from the authentication request.
	ClientIP string

	// Location, if present, holds additional information about the location of the client IP.
	Location string

	// Password holds the password from the authentication request.
	Password string

	// Username holds the username from the authentication request.
	Username string
}

// NewOpenVPNClientRequest creates a new OpenVPNClientRequest object based on environment variables.
func NewOpenVPNClientRequest() *OpenVPNClientRequest {
	req := &OpenVPNClientRequest{
		Username: os.Getenv("username"),
		Password: os.Getenv("password"),
		ClientIP: os.Getenv("untrusted_ip"),
	}
	req.Location = getLocation(req.ClientIP)
	return req
}

// getLocation returns the location of the IP address, if known, or "(unknown)" if an error occurs.
func getLocation(ip string) string {
	config := app.Config.Auth
	logger := log.With().
		Str("database", config.GeoIPDBPath).
		Str("ip", ip).
		Logger()

	// no database specified - just return and empty location
	if config.GeoIPDBPath == "" {
		return "(unknown)"
	}

	// open the database
	db, err := geoip2.Open(config.GeoIPDBPath)
	if err != nil {
		e := &errors.GeoIPDatabaseFailure{
			DatabaseFile: config.GeoIPDBPath,
			Err:          err,
		}
		logger.Error().Err(e.InternalError()).Msg(e.Error())
		return "(unknown)"
	}
	defer db.Close()

	// lookup the IP address
	record, err := db.City(net.ParseIP(ip))
	if err != nil {
		e := &errors.GeoIPLookupFailure{
			ClientIP: ip,
			Err:      err,
		}
		logger.Error().Err(e.InternalError()).Msg(e.Error())
		return "(unknown)"
	}

	// build location
	location := []string{}
	if record.City.Names[config.GeoIPLocale] != "" {
		location = append(location, record.City.Names[config.GeoIPLocale])
	}
	for _, s := range record.Subdivisions {
		if s.Names[config.GeoIPLocale] != "" {
			location = append(location, s.Names[config.GeoIPLocale])
		}
	}
	if record.Country.Names[config.GeoIPLocale] != "" {
		location = append(location, record.Country.Names[config.GeoIPLocale])
	}
	return strings.Join(location, ", ")
}
