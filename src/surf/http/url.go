package http

import (
	"os"
)

const (
	URL_BaseProduction        = ""
	URL_BaseDevelopment       = ""
	URL_BaseLocal             = ""
	URL_Login                 = ""
	URL_Logout                = ""
	URL_ValidateTunnelToken   = ""
	URL_ValidateTunnelRequest = ""
)

func GetLoginURL() string {
	return getBaseURL(os.Getenv("OCEAN_SERVER")) + URL_Login
}

func GetLogoutURL() string {
	return getBaseURL(os.Getenv("OCEAN_SERVER")) + URL_Logout
}

func GetValidateTunnelTokenURL() string {
	return getBaseURL(os.Getenv("OCEAN_SERVER")) + URL_ValidateTunnelToken
}

func GetValidateTunnelRequestURL() string {
	return getBaseURL(os.Getenv("OCEAN_SERVER")) + URL_ValidateTunnelRequest
}

func getBaseURL(server string) string {
	if server == "production" {
		return URL_BaseProduction
	} else if server == "development" {
		return URL_BaseDevelopment
	} else {
		return URL_BaseLocal
	}
}
