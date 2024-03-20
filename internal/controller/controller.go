package controller

import (
	"context"
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
}
type GitHubCredentials struct {
	AppId          *int64
	InstallationId *int64
	PrivateKey     []byte
	OAuthToken     *string
}

func GitHubCredentialsFromEnv() (*GitHubCredentials, error) {
	creds := &GitHubCredentials{}
	appId, ok := os.LookupEnv("GITHUB_APP_ID")
	if ok {
		value, err := strconv.ParseInt(appId, 10, 64)
		if err != nil {
			return nil, err
		}
		creds.AppId = &value
	}
	instId, ok := os.LookupEnv("GITHUB_INSTALLATION_ID")
	if ok {
		value, err := strconv.ParseInt(instId, 10, 64)
		if err != nil {
			return nil, err
		}
		creds.InstallationId = &value
	}
	pKey, ok := os.LookupEnv("GITHUB_PRIVATE_KEY")
	if ok {
		creds.PrivateKey = []byte(pKey)
	}
	token, ok := os.LookupEnv("GITHUB_TOKEN")
	if ok {
		creds.OAuthToken = &token
	}
	return creds, nil
}

func (ghc *GitHubCredentials) HasCredentials() bool {
	return ghc.HasInstallationCredentials() || ghc.HasOauthCredentials()
}

func (ghc *GitHubCredentials) HasInstallationCredentials() bool {
	return ghc.AppId != nil && ghc.InstallationId != nil && len(ghc.PrivateKey) > 0
}

func (ghc *GitHubCredentials) InstallationCredentials() (int64, int64, []byte, bool) {
	if ghc.HasInstallationCredentials() {
		return *ghc.AppId, *ghc.InstallationId, ghc.PrivateKey, true
	}
	return 0, 0, []byte{}, false
}

func (ghc *GitHubCredentials) HasOauthCredentials() bool {
	return ghc.AppId != nil && ghc.InstallationId != nil && len(ghc.PrivateKey) > 0
}

func (ghc *GitHubCredentials) OauthCredentials() (string, bool) {
	if ghc.HasOauthCredentials() {
		return *ghc.OAuthToken, true
	}
	return "", false
}

func NewGitHubClientFromCredentials(ctx context.Context, creds *GitHubCredentials) (*gh.Client, error) {
	base := http.DefaultTransport
	if appId, instId, pKey, ok := creds.InstallationCredentials(); ok {
		tr, err := gh.AuthRoundTripperFromAppCredentials(ctx, base, appId, instId, pKey)
		if err != nil {
			return nil, err
		}
		tr, err = gh.RateLimitRoundTripper(ctx, tr)
		if err != nil {
			return nil, err
		}
		return gh.NewClient(gh.WithRoundTripper(tr))
	} else if token, ok := creds.OauthCredentials(); ok {
		tr, err := gh.AuthRoundTripperFromToken(ctx, base, token)
		if err != nil {
			return nil, err
		}
		tr, err = gh.RateLimitRoundTripper(ctx, tr)
		if err != nil {
			return nil, err
		}
		return gh.NewClient(gh.WithRoundTripper(tr))
	} else {
		return nil, fmt.Errorf("unable to obtain installation credentials or oauth credentials")
	}
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
