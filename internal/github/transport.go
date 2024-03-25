package github

import (
	"context"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/gofri/go-github-ratelimit/github_ratelimit"
	"golang.org/x/oauth2"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

func AuthRoundTripperFromToken(ctx context.Context, base http.RoundTripper, token string) (http.RoundTripper, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tr := oauth2.Transport{
		Source: ts,
		Base:   base,
	}
	return &tr, nil
}

func AuthRoundTripperFromAppCredentials(ctx context.Context, base http.RoundTripper, appId, installationId int64, pKey []byte) (http.RoundTripper, error) {
	tr, err := ghinstallation.New(base, appId, installationId, pKey)
	if err != nil {
		return nil, err
	}
	return tr, nil
}

func RateLimitRoundTripper(ctx context.Context, base http.RoundTripper, opts ...github_ratelimit.Option) (http.RoundTripper, error) {
	tr, err := github_ratelimit.NewRateLimitWaiter(base, opts...)
	if err != nil {
		return nil, err
	}
	return tr, nil
}

func RecorderRoundTripper(ctx context.Context, base http.RoundTripper, opts *recorder.Options) (*recorder.Recorder, error) {
	r, err := recorder.NewWithOptions(opts)
	if err != nil {
		return nil, err
	}
	return r, nil
}
