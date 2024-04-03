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

package utils

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
	"os/exec"
	"path"
	"reflect"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint:golint,revive
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

const (
	prometheusOperatorVersion = "v0.68.0"
	prometheusOperatorURL     = "https://github.com/prometheus-operator/prometheus-operator/" +
		"releases/download/%s/bundle.yaml"

	certmanagerVersion = "v1.5.3"
	certmanagerURLTmpl = "https://github.com/jetstack/cert-manager/releases/download/%s/cert-manager.yaml"
)

func warnError(err error) {
	fmt.Fprintf(GinkgoWriter, "warning: %v\n", err)
}

// InstallPrometheusOperator installs the prometheus Operator to be used to export the enabled metrics.
func InstallPrometheusOperator() error {
	url := fmt.Sprintf(prometheusOperatorURL, prometheusOperatorVersion)
	cmd := exec.Command("kubectl", "create", "-f", url)
	_, err := Run(cmd)
	return err
}

// Run executes the provided command within this context
func Run(cmd *exec.Cmd) ([]byte, error) {
	dir, _ := GetProjectDir()
	cmd.Dir = dir

	if err := os.Chdir(cmd.Dir); err != nil {
		fmt.Fprintf(GinkgoWriter, "chdir dir: %s\n", err)
	}

	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	command := strings.Join(cmd.Args, " ")
	fmt.Fprintf(GinkgoWriter, "running: %s\n", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("%s failed with error: (%v) %s", command, err, string(output))
	}

	return output, nil
}

// UninstallPrometheusOperator uninstalls the prometheus
func UninstallPrometheusOperator() {
	url := fmt.Sprintf(prometheusOperatorURL, prometheusOperatorVersion)
	cmd := exec.Command("kubectl", "delete", "-f", url)
	if _, err := Run(cmd); err != nil {
		warnError(err)
	}
}

// UninstallCertManager uninstalls the cert manager
func UninstallCertManager() {
	url := fmt.Sprintf(certmanagerURLTmpl, certmanagerVersion)
	cmd := exec.Command("kubectl", "delete", "-f", url)
	if _, err := Run(cmd); err != nil {
		warnError(err)
	}
}

// InstallCertManager installs the cert manager bundle.
func InstallCertManager() error {
	url := fmt.Sprintf(certmanagerURLTmpl, certmanagerVersion)
	cmd := exec.Command("kubectl", "apply", "-f", url)
	if _, err := Run(cmd); err != nil {
		return err
	}
	// Wait for cert-manager-webhook to be ready, which can take time if cert-manager
	// was re-installed after uninstalling on a cluster.
	cmd = exec.Command("kubectl", "wait", "deployment.apps/cert-manager-webhook",
		"--for", "condition=Available",
		"--namespace", "cert-manager",
		"--timeout", "5m",
	)

	_, err := Run(cmd)
	return err
}

// LoadImageToKindCluster loads a local docker image to the kind cluster
func LoadImageToKindClusterWithName(name string) error {
	cluster := "kind"
	if v, ok := os.LookupEnv("KIND_CLUSTER"); ok {
		cluster = v
	}
	kindOptions := []string{"load", "docker-image", name, "--name", cluster}
	cmd := exec.Command("kind", kindOptions...)
	_, err := Run(cmd)
	return err
}

// LoadImageToKindCluster loads a local docker image to the kind cluster
func LoadPodmanImageToKindClusterWithName(name string) (err error) {
	imgPath := ".e2e-test-image.tar"
	podmanOptions := []string{"save", name, "-o", imgPath}
	cmd := exec.Command("podman", podmanOptions...)
	_, err = Run(cmd)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, os.Remove(imgPath))
	}()

	cluster := "kind"
	if v, ok := os.LookupEnv("KIND_CLUSTER"); ok {
		cluster = v
	}
	kindOptions := []string{"load", "image-archive", imgPath, "--name", cluster}
	cmd = exec.Command("kind", kindOptions...)
	_, err = Run(cmd)
	return err
}

// GetNonEmptyLines converts given command output string into individual objects
// according to line breakers, and ignores the empty elements in it.
func GetNonEmptyLines(output string) []string {
	var res []string
	elements := strings.Split(output, "\n")
	for _, element := range elements {
		if element != "" {
			res = append(res, element)
		}
	}

	return res
}

// GetProjectDir will return the directory where the project is
func GetProjectDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return wd, err
	}
	wd = strings.Replace(wd, "/test/e2e", "", -1)
	return wd, nil
}

func VerifyPodUp(namespace, name string) error {
	cmd := exec.Command("kubectl", "get",
		"pods", "-l", fmt.Sprintf("control-plane=%s", name),
		"-o", "go-template={{ range .items }}"+
			"{{ if not .metadata.deletionTimestamp }}"+
			"{{ .metadata.name }}"+
			"{{ \"\\n\" }}{{ end }}{{ end }}",
		"-n", namespace,
	)

	podOutput, err := Run(cmd)
	if err != nil {
		return err
	}
	podNames := GetNonEmptyLines(string(podOutput))
	if len(podNames) != 1 {
		return fmt.Errorf("expect 1 controller pods running, but got %d", len(podNames))
	}
	controllerPodName := podNames[0]
	if !strings.Contains(controllerPodName, name) {
		return fmt.Errorf("'%s' should contain '%s'", controllerPodName, name)
	}

	// Validate pod status
	cmd = exec.Command("kubectl", "get",
		"pods", controllerPodName, "-o", "jsonpath={.status.phase}",
		"-n", namespace,
	)
	status, err := Run(cmd)
	if err != nil {
		return err
	}
	if string(status) != "Running" {
		return fmt.Errorf("controller pod in %s status", status)
	}
	return nil
}

func GetField(namespace, kind, name, jsonPath string) (string, error) {
	cmd := exec.Command("kubectl", "get", "-n", namespace, kind, name, fmt.Sprintf("-o=jsonpath=%s", jsonPath))
	output, err := Run(cmd)
	return string(output), err
}

func Patch(namespace, kind, name, patch string) error {
	cmd := exec.Command("kubectl", "patch", "-n", namespace, kind, name, "--type=json", "-p", patch)
	_, err := Run(cmd)
	return err
}

func Apply(namespace string, object map[string]interface{}) error {
	cmd := exec.Command("kubectl", "apply", "-n", namespace, "-f", "-")
	objJson, err := json.Marshal(object)
	if err != nil {
		return err
	}
	cmd.Stdin = bytes.NewBuffer(objJson)
	if _, err := Run(cmd); err != nil {
		return err
	}
	return nil
}

func Delete(namespace string, object map[string]interface{}) error {
	cmd := exec.Command("kubectl", "delete", "-n", namespace, "-f", "-")
	objJson, err := json.Marshal(object)
	if err != nil {
		return err
	}
	cmd.Stdin = bytes.NewBuffer(objJson)
	if _, err := Run(cmd); err != nil {
		return err
	}
	return nil
}

func LoadEnvVarsError(names ...string) (map[string]string, error) {
	varMap := map[string]string{}
	notSet := []string{}
	for _, name := range names {
		if v, ok := os.LookupEnv(name); ok {
			varMap[name] = v
		} else {
			notSet = append(notSet, name)
		}
	}
	if len(notSet) > 0 {
		return nil, fmt.Errorf("expected env vars to be set: %v", notSet)
	}
	return varMap, nil
}

// nolint
func GitHubRecorderRoundTripper(ctx context.Context, base http.RoundTripper, opts *recorder.Options) (*recorder.Recorder, error) {
	r, err := recorder.NewWithOptions(opts)
	if err != nil {
		return nil, err
	}
	rmSecretsHook := func(i *cassette.Interaction) error {
		delete(i.Request.Headers, "Authorization")
		if path.Base(i.Request.URL) == "access_tokens" {
			var body map[string]interface{}
			err := json.Unmarshal([]byte(i.Response.Body), &body)
			if err != nil {
				log.Fatal(err)
			}
			body["token"] = ""
			body["expires_at"] = time.Now().Add(time.Hour * 24 * 365 * 200) // 200 years from now
			modified, err := json.Marshal(body)
			if err != nil {
				log.Fatal(err)
			}
			i.Response.Body = string(modified)
		}
		return nil
	}
	r.AddHook(rmSecretsHook, recorder.BeforeSaveHook)
	r.SetMatcher(func(r1 *http.Request, r2 cassette.Request) bool {
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
	return r, nil
}
