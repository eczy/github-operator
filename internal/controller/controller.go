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

package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	gh "github.com/eczy/github-operator/internal/github"
)

type GitHubRequester interface {
	TeamRequester
	RepositoryRequester
	OrganizationRequester
	BranchProtectionRequester
}

type GitHubInstallationCredentials struct {
	AppId          int64
	InstallationId int64
	PrivateKey     []byte
}

type GitHubOauthCredentials struct {
	OAuthToken string
}

func GitHubInstallationCredentialsFromEnv() (*GitHubInstallationCredentials, error) {
	creds := &GitHubInstallationCredentials{}
	errs := []error{}
	appId, ok := os.LookupEnv("GITHUB_APP_ID")
	if ok {
		value, err := strconv.ParseInt(appId, 10, 64)
		if err != nil {
			return nil, err
		}
		creds.AppId = value
	} else {
		errs = append(errs, fmt.Errorf("env var GITHUB_APP_ID not found"))
	}
	instId, ok := os.LookupEnv("GITHUB_INSTALLATION_ID")
	if ok {
		value, err := strconv.ParseInt(instId, 10, 64)
		if err != nil {
			return nil, err
		}
		creds.InstallationId = value
	} else {
		errs = append(errs, fmt.Errorf("env var GITHUB_INSTALLATION_ID not found"))
	}
	pKey, ok := os.LookupEnv("GITHUB_PRIVATE_KEY")
	if ok {
		creds.PrivateKey = []byte(pKey)
	} else {
		errs = append(errs, fmt.Errorf("env var GITHUB_PRIVATE_KEY not found"))
	}
	return creds, errors.Join(errs...)
}

func GitHubOauthCredentialsFromEnv() (*GitHubOauthCredentials, error) {
	creds := &GitHubOauthCredentials{}

	token, ok := os.LookupEnv("GITHUB_TOKEN")
	if ok {
		creds.OAuthToken = token
	} else {
		return nil, fmt.Errorf("env var GITHUB_TOKEN not found")
	}
	return creds, nil
}

func NewGitHubClientFromInstallationCredentials(ctx context.Context, creds GitHubInstallationCredentials, base http.RoundTripper) (*gh.Client, error) {
	tr, err := gh.AuthRoundTripperFromAppCredentials(ctx, base, creds.AppId, creds.InstallationId, creds.PrivateKey)
	if err != nil {
		return nil, err
	}
	tr, err = gh.RateLimitRoundTripper(ctx, tr)
	if err != nil {
		return nil, err
	}
	return gh.NewClient(gh.WithRoundTripper(tr))
}

func NewGitHubClientFromOauthCredentials(ctx context.Context, creds GitHubOauthCredentials, base http.RoundTripper) (*gh.Client, error) {
	tr, err := gh.AuthRoundTripperFromToken(ctx, base, creds.OAuthToken)
	if err != nil {
		return nil, err
	}
	tr, err = gh.RateLimitRoundTripper(ctx, tr)
	if err != nil {
		return nil, err
	}
	return gh.NewClient(gh.WithRoundTripper(tr))
}

// ptrNonNilAndNotEqualTo returns true if a is not nil and its underlying value does not equal b.
// This is convenient for determining if an optional field needs to be updated or not compared to a
// non-pointer variable.
func ptrNonNilAndNotEqualTo[T comparable](a *T, b T) bool {
	if a == nil {
		return false
	}
	return *a != b
}

// returns if set(a) is equivalent to set(b)
// inefficient if the slices are frequently compared since this constructs
// a new set every time it is called
func cmpSlices[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	setA := map[T]struct{}{}
	for _, a := range a {
		setA[a] = struct{}{}
	}
	for _, b := range b {
		if _, ok := setA[b]; !ok {
			return false
		}
	}
	return true
}
