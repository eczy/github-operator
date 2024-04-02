package utils

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	gh "github.com/eczy/github-operator/internal/github"
)

func LookupEnvVarsError(names ...string) (map[string]string, error) {
	found := map[string]string{}
	missing := []string{}
	for _, name := range names {
		if v, ok := os.LookupEnv(name); ok {
			found[name] = v
		} else {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("expected env vars not set: %v", missing)
	}
	return found, nil
}

func GitHubClientFromEnv(ctx context.Context, base http.RoundTripper) (*gh.Client, error) {
	appCreds, appErr := LookupEnvVarsError("GITHUB_APP_ID", "GTIHUB_INSTALLATION_ID", "GITHUB_PRIVATE_KEY")
	oauthCreds, oauthErr := LookupEnvVarsError("GITHUB_TOKEN")
	if appErr == nil {
		appId, err := strconv.ParseInt(appCreds["GITHUB_APP_ID"], 10, 64)
		if err != nil {
			return nil, err
		}
		instId, err := strconv.ParseInt(appCreds["GITHUB_INSTALLATION_ID"], 10, 64)
		if err != nil {
			return nil, err
		}
		tr, err := gh.AuthRoundTripperFromAppCredentials(ctx, base, appId, instId, []byte(appCreds["GITHUB_PRIVATE_KEY"]))
		if err != nil {
			return nil, err
		}
		tr, err = gh.RateLimitRoundTripper(ctx, tr)
		if err != nil {
			return nil, err
		}
		return gh.NewClient(gh.WithRoundTripper(tr))
	} else if oauthErr != nil {
		tr, err := gh.AuthRoundTripperFromToken(ctx, base, oauthCreds["GITHUB_TOKEN"])
		if err != nil {
			return nil, err
		}
		tr, err = gh.RateLimitRoundTripper(ctx, tr)
		if err != nil {
			return nil, err
		}
		return gh.NewClient(gh.WithRoundTripper(tr))
	} else {
		return nil, errors.Join(appErr, oauthErr)
	}
}
