package cesta

import (
	"net/url"
	"regexp"
)

func IsValidUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func IsValidHostname(hostname string) bool {
	var hostnameRegex = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)*([a-zA-Z]{2,})$`)
	return hostnameRegex.MatchString(hostname)
}
