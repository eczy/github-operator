/*
Copyright 2025.

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
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	githuboperatoreczyiov1 "github.com/eczy/github-operator/api/v1"
	gh "github.com/eczy/github-operator/internal/github"
	"github.com/eczy/github-operator/internal/utils"
	testutils "github.com/eczy/github-operator/test/utils"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	ctx       context.Context
	cancel    context.CancelFunc
	testEnv   *envtest.Environment
	cfg       *rest.Config
	k8sClient client.Client

	gitHubClient         GitHubRequester
	deletionGracePeriod  int64  = 5
	testOrganization     string = "testorg"
	testUser             string = "testuser"
	ghTestResourcePrefix string = "github-operator-test-"
	vcrRecorder          *recorder.Recorder
)

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	var err error
	err = githuboperatoreczyiov1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	// Retrieve the first found binary directory to allow running tests from IDEs
	if getFirstFoundEnvTestBinaryDir() != "" {
		testEnv.BinaryAssetsDirectory = getFirstFoundEnvTestBinaryDir()
	}

	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	By("Setting up the GitHub client")
	org, ok := os.LookupEnv("GITHUB_OPERATOR_TEST_ORG")
	if ok {
		testOrganization = org
	}

	user, ok := os.LookupEnv("GITHUB_OPERATOR_TEST_USER")
	if ok {
		testUser = user
	}

	recorderMode := recorder.ModeRecordOnce
	mode, ok := os.LookupEnv("GITHUB_OPERATOR_RECORDER_MODE")
	lowerMode := strings.ToLower(mode)
	if ok {
		switch lowerMode {
		case "record-only":
			recorderMode = recorder.ModeRecordOnly
		case "replay-only":
			recorderMode = recorder.ModeReplayOnly
		case "replay-with-new-episodes":
			recorderMode = recorder.ModeReplayWithNewEpisodes
		case "record-once":
			recorderMode = recorder.ModeRecordOnce
		case "mode-passthrough":
			recorderMode = recorder.ModePassthrough
		default:
			err := fmt.Errorf("invalid value for recorder mode: %s", mode)
			Expect(err).NotTo(HaveOccurred())
		}
	}

	cassetteName := "fixtures/test-github-operator-controller"
	cassettePath := path.Dir(cassetteName)
	if _, err := os.Stat(cassettePath); os.IsNotExist(err) {
		err := os.MkdirAll(cassettePath, 0755)
		Expect(err).NotTo(HaveOccurred())
	}

	ctx := context.Background()
	base := http.DefaultTransport
	rec, err := testutils.GitHubRecorderRoundTripper(ctx, base, cassetteName, []recorder.Option{
		recorder.WithMode(recorderMode),
		recorder.WithRealTransport(http.DefaultTransport),
		recorder.WithSkipRequestLatency(true),
	})
	Expect(err).NotTo(HaveOccurred())
	vcrRecorder = rec
	c, err := utils.GitHubClientFromEnv(ctx, vcrRecorder)
	if err != nil {
		if lowerMode == "replay-only" {
			// continue in replay mode
			c, err := gh.NewClient(gh.WithRoundTripper(vcrRecorder))
			Expect(err).NotTo(HaveOccurred())
			gitHubClient = c
		} else {
			Fail("no GitHub credentials found; pass credentials or run in 'replay-only' mode")
		}
	} else {
		gitHubClient = c
	}
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	cancel()
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

// getFirstFoundEnvTestBinaryDir locates the first binary in the specified path.
// ENVTEST-based tests depend on specific binaries, usually located in paths set by
// controller-runtime. When running tests directly (e.g., via an IDE) without using
// Makefile targets, the 'BinaryAssetsDirectory' must be explicitly configured.
//
// This function streamlines the process by finding the required binaries, similar to
// setting the 'KUBEBUILDER_ASSETS' environment variable. To ensure the binaries are
// properly set up, run 'make setup-envtest' beforehand.
func getFirstFoundEnvTestBinaryDir() string {
	basePath := filepath.Join("..", "..", "bin", "k8s")
	entries, err := os.ReadDir(basePath)
	if err != nil {
		logf.Log.Error(err, "Failed to read directory", "path", basePath)
		return ""
	}
	for _, entry := range entries {
		if entry.IsDir() {
			return filepath.Join(basePath, entry.Name())
		}
	}
	return ""
}
