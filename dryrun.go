package logroundtripper

import "net/http"

// DryRunRoundTripper - use it if you want to do nothing
type DryRunRoundTripper struct{}

// RoundTrip will do nothing
func (drt *DryRunRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{}, nil
}
