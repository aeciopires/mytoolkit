// Package urlencode implements the URL Encode/Decode tool's pure logic.
package urlencode

import (
	"net/url"

	"github.com/aeciopires/mytoolkit/internal/apperr"
)

type Options struct {
	Decode    bool   `json:"decode"`
	Component string `json:"component"` // "query" | "path" | "full", default "query"
}

func Process(input []byte, opts Options) (string, error) {
	component := opts.Component
	if component == "" {
		component = "query"
	}
	if err := apperr.OneOf("component", component, "query", "path", "full"); err != nil {
		return "", err
	}

	text := string(input)
	if len(text) == 0 {
		return "", nil
	}

	if opts.Decode {
		return decode(text, component)
	}
	return encode(text, component)
}

func encode(text, component string) (string, error) {
	switch component {
	case "path":
		return url.PathEscape(text), nil
	case "full":
		u, err := url.Parse(text)
		if err != nil {
			return "", apperr.Newf(400, "INVALID_URL", "%s", err.Error())
		}
		return u.String(), nil
	default:
		return url.QueryEscape(text), nil
	}
}

func decode(text, component string) (string, error) {
	switch component {
	case "path":
		out, err := url.PathUnescape(text)
		if err != nil {
			return "", apperr.Newf(400, "INVALID_ENCODING", "%s", err.Error())
		}
		return out, nil
	case "full":
		out, err := url.QueryUnescape(text)
		if err != nil {
			return "", apperr.Newf(400, "INVALID_ENCODING", "%s", err.Error())
		}
		return out, nil
	default:
		out, err := url.QueryUnescape(text)
		if err != nil {
			return "", apperr.Newf(400, "INVALID_ENCODING", "%s", err.Error())
		}
		return out, nil
	}
}
