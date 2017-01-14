package util

import (
	"net/url"
	"net"
	"fmt"
)

func ParseUrl(sourceUrl string) (u string) {
	URL, err := url.Parse(sourceUrl)
	if err != nil {
		panic(err)
	}

	host, _, err := net.SplitHostPort(URL.Host)
	if err != nil {
		panic(err)
	}

	u = fmt.Sprintf("%s://%s", URL.Scheme, host)
	return
}
