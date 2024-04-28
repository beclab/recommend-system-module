package browser

import (
	"errors"

	"bytetrade.io/web3os/backend-server/http/client"
)

var (
	errRequestFailed    = "Unable to open this link: %v"
	errServerFailure    = "Unable to fetch this resource (Status Code = %d)"
	errEncoding         = "Unable to normalize encoding: %q"
	errEmptyFeed        = "This feed is empty"
	errResourceNotFound = "Resource not found (404), this feed doesn't exist anymore, check the feed URL"
	errNotAuthorized    = "You are not authorized to access this resource (invalid username/password)"
)

// Exec executes a HTTP request and handles errors.
func Exec(request *client.Client) (*client.Response, error) {
	response, err := request.Get()
	if err != nil {
		return nil, errors.New(errRequestFailed)
	}

	if response.IsNotFound() {
		return nil, errors.New(errResourceNotFound)
	}

	if response.IsNotAuthorized() {
		return nil, errors.New(errNotAuthorized)
	}

	if response.HasServerFailure() {
		return nil, errors.New(errServerFailure)
	}

	if response.StatusCode != 304 {
		// Content-Length = -1 when no Content-Length header is sent.
		if response.ContentLength == 0 {
			return nil, errors.New(errEmptyFeed)
		}

		if err := response.EnsureUnicodeBody(); err != nil {
			return nil, errors.New(errEncoding)
		}
	}

	return response, nil
}
