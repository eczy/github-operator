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

package e2e

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/eczy/github-operator/internal/controller"
	"github.com/eczy/github-operator/test/utils"
)

// TODO: these tests could be cleaned up considerably if we move common functionality
// to the utils package

const namespace = "github-operator-system"

var _ = Describe("controller", Ordered, func() {
	BeforeAll(func() {
		By("installing prometheus operator")
		Expect(utils.InstallPrometheusOperator()).To(Succeed())

		By("installing the cert-manager")
		Expect(utils.InstallCertManager()).To(Succeed())

		By("creating manager namespace")
		cmd := exec.Command("kubectl", "create", "ns", namespace)
		_, _ = utils.Run(cmd)
	})

	AfterAll(func() {
		By("uninstalling the Prometheus manager bundle")
		utils.UninstallPrometheusOperator()

		By("uninstalling the cert-manager bundle")
		utils.UninstallCertManager()

		By("removing manager namespace")
		cmd := exec.Command("kubectl", "delete", "ns", namespace)
		_, _ = utils.Run(cmd)
	})

	Context("Operator", func() {
		It("should run successfully", func() {
			var controllerPodName string
			var err error

			containerTool := "podman"
			if v, ok := os.LookupEnv("CONAINER_TOOL"); ok {
				containerTool = v
			}

			// projectimage stores the name of the image used in the example
			var projectimage = "example.com/github-operator:v0.0.1"

			By("building the manager(Operator) image")
			cmd := exec.Command("make", "docker-build", fmt.Sprintf("IMG=%s", projectimage))
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("loading the the manager(Operator) image on Kind")
			if containerTool == "podman" {
				err = utils.LoadPodmanImageToKindClusterWithName(projectimage)
				ExpectWithOffset(1, err).NotTo(HaveOccurred())
			} else {
				err = utils.LoadImageToKindClusterWithName(projectimage)
				ExpectWithOffset(1, err).NotTo(HaveOccurred())
			}

			By("installing CRDs")
			cmd = exec.Command("make", "install")
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("deploying the controller-manager")
			cmd = exec.Command("make", "deploy", fmt.Sprintf("IMG=%s", projectimage))
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("validating that the controller-manager pod is running as expected")
			verifyControllerUp := func() error {
				// Get pod name

				cmd = exec.Command("kubectl", "get",
					"pods", "-l", "control-plane=controller-manager",
					"-o", "go-template={{ range .items }}"+
						"{{ if not .metadata.deletionTimestamp }}"+
						"{{ .metadata.name }}"+
						"{{ \"\\n\" }}{{ end }}{{ end }}",
					"-n", namespace,
				)

				podOutput, err := utils.Run(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				podNames := utils.GetNonEmptyLines(string(podOutput))
				if len(podNames) != 1 {
					return fmt.Errorf("expect 1 controller pods running, but got %d", len(podNames))
				}
				controllerPodName = podNames[0]
				ExpectWithOffset(2, controllerPodName).Should(ContainSubstring("controller-manager"))

				// Validate pod status
				cmd = exec.Command("kubectl", "get",
					"pods", controllerPodName, "-o", "jsonpath={.status.phase}",
					"-n", namespace,
				)
				status, err := utils.Run(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				if string(status) != "Running" {
					return fmt.Errorf("controller pod in %s status", status)
				}
				return nil
			}
			EventuallyWithOffset(1, verifyControllerUp, time.Minute, time.Second).Should(Succeed())
		})
		It("should manage the full lifecycle of a Team", func() {
			var controllerPodName string
			var err error

			containerTool := "podman"
			if v, ok := os.LookupEnv("CONAINER_TOOL"); ok {
				containerTool = v
			}

			testOrg, ok := os.LookupEnv("GITHUB_OPERATOR_TEST_ORG")
			Expect(ok).To(BeTrue(), "this text expects GITHUB_OPERATOR_TEST_ORG to be set")

			var credEnvVars []string
			instCreds, err0 := controller.GitHubInstallationCredentialsFromEnv()
			oauthCreds, err1 := controller.GitHubOauthCredentialsFromEnv()
			if err0 == nil {
				credEnvVars = []string{
					fmt.Sprintf("GITHUB_APP_ID=%d", instCreds.AppId),
					fmt.Sprintf("GITHUB_INSTALLATION_ID=%d", instCreds.InstallationId),
					fmt.Sprintf("GITHUB_PRIVATE_KEY=%s", instCreds.PrivateKey),
				}
			} else if err1 == nil {
				credEnvVars = []string{fmt.Sprintf("GITHUB_TOKEN=%s", oauthCreds.OAuthToken)}
			} else {
				Expect(errors.Join(err0, err1)).To(BeNil(), "valid GitHub credentials required for this test")
			}

			// projectimage stores the name of the image used in the example
			var projectimage = "example.com/github-operator:v0.0.1"

			By("building the manager(Operator) image")
			cmd := exec.Command("make", "docker-build", fmt.Sprintf("IMG=%s", projectimage))
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("loading the the manager(Operator) image on Kind")
			if containerTool == "podman" {
				err = utils.LoadPodmanImageToKindClusterWithName(projectimage)
				ExpectWithOffset(1, err).NotTo(HaveOccurred())
			} else {
				err = utils.LoadImageToKindClusterWithName(projectimage)
				ExpectWithOffset(1, err).NotTo(HaveOccurred())
			}

			By("installing CRDs")
			cmd = exec.Command("make", "install")
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("deploying the controller-manager")
			cmd = exec.Command("make", "deploy", fmt.Sprintf("IMG=%s", projectimage))
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			// TODO: configure the manager properly the first time instead of patching here
			By("setting GitHub credentials for the manager")
			kubectlArgs := []string{
				"set", "env", "-n", namespace, "deployment/github-operator-controller-manager",
			}
			kubectlArgs = append(kubectlArgs, credEnvVars...)
			cmd = exec.Command("kubectl", kubectlArgs...)
			err = cmd.Run() // use raw 'Run' since we have a secret in args
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			err = utils.Patch(namespace, "deployment", "github-operator-controller-manager", `
[{"op": "add", "path": "/spec/template/spec/containers/1/args/-", "value": "--delete-on-resource-deletion"}]
`)

			By("validating that the controller-manager pod is running as expected")
			verifyControllerUp := func() error {
				// Get pod name

				cmd = exec.Command("kubectl", "get",
					"pods", "-l", "control-plane=controller-manager",
					"-o", "go-template={{ range .items }}"+
						"{{ if not .metadata.deletionTimestamp }}"+
						"{{ .metadata.name }}"+
						"{{ \"\\n\" }}{{ end }}{{ end }}",
					"-n", namespace,
				)

				podOutput, err := utils.Run(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				podNames := utils.GetNonEmptyLines(string(podOutput))
				if len(podNames) != 1 {
					return fmt.Errorf("expect 1 controller pods running, but got %d", len(podNames))
				}
				controllerPodName = podNames[0]
				ExpectWithOffset(2, controllerPodName).Should(ContainSubstring("controller-manager"))
				return nil
			}
			EventuallyWithOffset(1, verifyControllerUp, time.Minute, time.Second).Should(Succeed())
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("creating a Team resource")
			err = utils.Apply(namespace, map[string]interface{}{
				"apiVersion": "github.github-operator.eczy.io/v1alpha1",
				"kind":       "Team",
				"metadata": map[string]interface{}{
					"name": "test-team",
				},
				"spec": map[string]interface{}{
					"organization": testOrg,
					"name":         "test-team",
				},
			})
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			EventuallyWithOffset(1, func() error {
				v, err := utils.GetField(namespace, "Team", "test-team", "{.status.name}")
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				if v != "test-team" {
					return fmt.Errorf("'%s' should equal 'test-team'", v)
				}
				return nil
			}, time.Minute, time.Second).Should(Succeed())
			By("updating a Team resource")
			err = utils.Apply(namespace, map[string]interface{}{
				"apiVersion": "github.github-operator.eczy.io/v1alpha1",
				"kind":       "Team",
				"metadata": map[string]interface{}{
					"name": "test-team",
				},
				"spec": map[string]interface{}{
					"organization": testOrg,
					"name":         "test-team",
					"description":  "foo",
				},
			})
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			EventuallyWithOffset(1, func() error {
				v, err := utils.GetField(namespace, "Team", "test-team", "{.status.description}")
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				if v != "foo" {
					return fmt.Errorf("'%s' should equal 'foo'", v)
				}
				return nil
			}, time.Minute, time.Second).Should(Succeed())
			By("deleting a Team resource")
			err = utils.Delete(namespace, map[string]interface{}{
				"apiVersion": "github.github-operator.eczy.io/v1alpha1",
				"kind":       "Team",
				"metadata": map[string]interface{}{
					"name": "test-team",
				},
				"spec": map[string]interface{}{
					"organization": testOrg,
					"name":         "test-team",
					"description":  "foo",
				},
			})
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
		})
		It("should manage the full lifecycle of a Repository", func() {
			var controllerPodName string
			var err error

			containerTool := "podman"
			if v, ok := os.LookupEnv("CONAINER_TOOL"); ok {
				containerTool = v
			}

			testOrg, ok := os.LookupEnv("GITHUB_OPERATOR_TEST_ORG")
			Expect(ok).To(BeTrue(), "this text expects GITHUB_OPERATOR_TEST_ORG to be set")

			var credEnvVars []string
			instCreds, err0 := controller.GitHubInstallationCredentialsFromEnv()
			oauthCreds, err1 := controller.GitHubOauthCredentialsFromEnv()
			if err0 == nil {
				credEnvVars = []string{
					fmt.Sprintf("GITHUB_APP_ID=%d", instCreds.AppId),
					fmt.Sprintf("GITHUB_INSTALLATION_ID=%d", instCreds.InstallationId),
					fmt.Sprintf("GITHUB_PRIVATE_KEY=%s", instCreds.PrivateKey),
				}
			} else if err1 == nil {
				credEnvVars = []string{fmt.Sprintf("GITHUB_TOKEN=%s", oauthCreds.OAuthToken)}
			} else {
				Expect(errors.Join(err0, err1)).To(BeNil(), "valid GitHub credentials required for this test")
			}

			// projectimage stores the name of the image used in the example
			var projectimage = "example.com/github-operator:v0.0.1"

			By("building the manager(Operator) image")
			cmd := exec.Command("make", "docker-build", fmt.Sprintf("IMG=%s", projectimage))
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("loading the the manager(Operator) image on Kind")
			if containerTool == "podman" {
				err = utils.LoadPodmanImageToKindClusterWithName(projectimage)
				ExpectWithOffset(1, err).NotTo(HaveOccurred())
			} else {
				err = utils.LoadImageToKindClusterWithName(projectimage)
				ExpectWithOffset(1, err).NotTo(HaveOccurred())
			}

			By("installing CRDs")
			cmd = exec.Command("make", "install")
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("deploying the controller-manager")
			cmd = exec.Command("make", "deploy", fmt.Sprintf("IMG=%s", projectimage))
			_, err = utils.Run(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			// TODO: configure the manager properly the first time instead of patching here
			By("setting GitHub credentials for the manager")
			kubectlArgs := []string{
				"set", "env", "-n", namespace, "deployment/github-operator-controller-manager",
			}
			kubectlArgs = append(kubectlArgs, credEnvVars...)
			cmd = exec.Command("kubectl", kubectlArgs...)
			err = cmd.Run() // use raw 'Run' since we have a secret in args
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			err = utils.Patch(namespace, "deployment", "github-operator-controller-manager", `
[{"op": "add", "path": "/spec/template/spec/containers/1/args/-", "value": "--delete-on-resource-deletion"}]
`)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("validating that the controller-manager pod is running as expected")
			verifyControllerUp := func() error {
				// Get pod name

				cmd = exec.Command("kubectl", "get",
					"pods", "-l", "control-plane=controller-manager",
					"-o", "go-template={{ range .items }}"+
						"{{ if not .metadata.deletionTimestamp }}"+
						"{{ .metadata.name }}"+
						"{{ \"\\n\" }}{{ end }}{{ end }}",
					"-n", namespace,
				)

				podOutput, err := utils.Run(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				podNames := utils.GetNonEmptyLines(string(podOutput))
				if len(podNames) != 1 {
					return fmt.Errorf("expect 1 controller pods running, but got %d", len(podNames))
				}
				controllerPodName = podNames[0]
				ExpectWithOffset(2, controllerPodName).Should(ContainSubstring("controller-manager"))
				return nil
			}
			EventuallyWithOffset(1, verifyControllerUp, time.Minute, time.Second).Should(Succeed())

			By("creating a Repository resource")
			err = utils.Apply(namespace, map[string]interface{}{
				"apiVersion": "github.github-operator.eczy.io/v1alpha1",
				"kind":       "Repository",
				"metadata": map[string]interface{}{
					"name": "test-repo",
				},
				"spec": map[string]interface{}{
					"owner": testOrg,
					"name":  "test-repo",
				},
			})
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			EventuallyWithOffset(1, func() error {
				v, err := utils.GetField(namespace, "Repository", "test-repo", "{.status.name}")
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				if v != "test-repo" {
					return fmt.Errorf("'%s' should equal 'test-repo'", v)
				}
				return nil
			}, time.Minute, time.Second).Should(Succeed())
			By("updating a Repository resource")
			err = utils.Apply(namespace, map[string]interface{}{
				"apiVersion": "github.github-operator.eczy.io/v1alpha1",
				"kind":       "Repository",
				"metadata": map[string]interface{}{
					"name": "test-repo",
				},
				"spec": map[string]interface{}{
					"owner":       testOrg,
					"name":        "test-repo",
					"description": "foo",
				},
			})
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			EventuallyWithOffset(1, func() error {
				v, err := utils.GetField(namespace, "Repository", "test-repo", "{.status.description}")
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				if v != "foo" {
					return fmt.Errorf("'%s' should equal 'foo'", v)
				}
				return nil
			}, time.Minute, time.Second).Should(Succeed())
			By("deleting a Repository resource")
			err = utils.Delete(namespace, map[string]interface{}{
				"apiVersion": "github.github-operator.eczy.io/v1alpha1",
				"kind":       "Repository",
				"metadata": map[string]interface{}{
					"name": "test-repo",
				},
				"spec": map[string]interface{}{
					"organization": testOrg,
					"name":         "test-repo",
					"description":  "foo",
				},
			})
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
		})
	})
})
