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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	githubv1alpha1 "github.com/eczy/github-operator/api/v1alpha1"
	gh "github.com/eczy/github-operator/internal/github" //+kubebuilder:scaffold:imports
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

	instCreds, err0 := GitHubInstallationCredentialsFromEnv()
	oauthCreds, err1 := GitHubOauthCredentialsFromEnv()
	base := http.DefaultTransport
	rec, err := gh.RecorderRoundTripper(ctx, base, &recorder.Options{
		CassetteName:       cassetteName,
		Mode:               recorderMode,
		RealTransport:      http.DefaultTransport,
		SkipRequestLatency: true,
	})
	Expect(err).NotTo(HaveOccurred())
	rmSecretsHook := func(i *cassette.Interaction) error {
		delete(i.Request.Headers, "Authorization")
		return nil
	}
	rec.AddHook(rmSecretsHook, recorder.AfterCaptureHook)
	rec.SetMatcher(func(r1 *http.Request, r2 cassette.Request) bool {
		// Method
		if r1.Method != r2.Method {
			return false
		}
		if r1.Method == "" && r2.Method != "GET" {
			// for client requests, "" means GET
			return false
		}
		// URL
		if r1.URL.String() != r2.URL {
			return false
		}
		// Body (as JSON)
		if r1.Body == nil && r2.Body == "" {
			return true
		}
		if r1.Body != nil && r2.Body == "" {
			return false
		}
		if r1.Body == nil && r2.Body != "" {
			return false
		}
		r1BodyBytes, err := io.ReadAll(r1.Body)
		if err != nil {
			log.Fatal(err)
		}
		err = r1.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
		r1.Body = io.NopCloser(bytes.NewBuffer(r1BodyBytes))
		if len(r1BodyBytes) == 0 && len(r2.Body) == 0 {
			return true
		}
		if len(r1BodyBytes) == 0 && len(r2.Body) != 0 {
			return false
		}
		if len(r1BodyBytes) != 0 && len(r2.Body) == 0 {
			return false
		}
		// at this point both r1 and r2 have bodies
		var r1Body interface{}
		err = json.Unmarshal(r1BodyBytes, &r1Body)
		if err != nil {
			log.Fatal(err)
		}
		r2BodyBytes := []byte(r2.Body)
		var r2Body interface{}
		err = json.Unmarshal(r2BodyBytes, &r2Body)
		if err != nil {
			log.Fatal(err)
		}
		return reflect.DeepEqual(r1Body, r2Body)
	})
	vcrRecorder = rec
	if err0 == nil {
		c, err := NewGitHubClientFromInstallationCredentials(ctx, *instCreds, vcrRecorder)
		Expect(err).NotTo(HaveOccurred())
		ghClient = c
	} else if err1 == nil {
		c, err := NewGitHubClientFromOauthCredentials(ctx, *oauthCreds, vcrRecorder)
		Expect(err).NotTo(HaveOccurred())
		ghClient = c
	} else if lowerMode == "replay-only" {
		fmt.Fprintf(os.Stderr, "no GitHub credentials found; continuing in 'replay-only' mode\n")
		c, err := gh.NewClient(gh.WithRoundTripper(vcrRecorder))
		Expect(err).NotTo(HaveOccurred())
		ghClient = c
	} else {
		fmt.Fprintf(os.Stderr, "no GitHub credentials found; set credential env vars or run in 'replay-only' mode\n")
		log.Fatal(errors.Join(err0, err1))
	}
})

var _ = AfterSuite(func() {
	By("Tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
	err = vcrRecorder.Stop()
	Expect(err).NotTo(HaveOccurred())
})
