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
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	githubv1alpha1 "github.com/eczy/github-operator/api/v1alpha1" //+kubebuilder:scaffold:imports
	gh "github.com/eczy/github-operator/internal/github"
	"github.com/eczy/github-operator/internal/utils"

	testutils "github.com/eczy/github-operator/test/utils"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg       *rest.Config
	k8sClient client.Client
	testEnv   *envtest.Environment

	ghClient             GitHubRequester
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

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,

		// The BinaryAssetsDirectory is only required if you want to run the tests directly
		// without call the makefile target test. If not informed it will look for the
		// default path defined in controller-runtime which is /usr/local/kubebuilder/.
		// Note that you must have the required binaries setup under the bin directory to perform
		// the tests directly. When we run make test it will be setup and used automatically.
		BinaryAssetsDirectory: filepath.Join("..", "..", "bin", "k8s",
			fmt.Sprintf("1.29.0-%s-%s", runtime.GOOS, runtime.GOARCH)),
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = githubv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

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
	mode, ok := os.LookupEnv("GITHUB_OPERATOR_RECORD_MODE")
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
	rec, err := testutils.GitHubRecorderRoundTripper(ctx, base, &recorder.Options{
		CassetteName:       cassetteName,
		Mode:               recorderMode,
		RealTransport:      http.DefaultTransport,
		SkipRequestLatency: true,
	})
	Expect(err).NotTo(HaveOccurred())
	vcrRecorder = rec
	c, err := utils.GitHubClientFromEnv(ctx, vcrRecorder)
	if err != nil {
		if lowerMode == "replay-only" {
			// continue in replay mode
			c, err := gh.NewClient(gh.WithRoundTripper(vcrRecorder))
			Expect(err).NotTo(HaveOccurred())
			ghClient = c
		} else {
			Fail("no GitHub credentials found; pass credentials or run in 'replay-only' mode")
		}
	} else {
		ghClient = c
	}
})

var _ = AfterSuite(func() {
	By("Tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
	err = vcrRecorder.Stop()
	Expect(err).NotTo(HaveOccurred())
})
