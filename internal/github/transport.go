/*
Copyright 2024 Evan Czyzycki

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
