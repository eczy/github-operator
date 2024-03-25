/*
Copyright 2024.

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

	"github.com/google/go-github/v60/github"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	githubv1alpha1 "github.com/eczy/github-operator/api/v1alpha1"
)

var _ = Describe("BranchProtection Controller", func() {
	const resourceName = "test-resource"

	ctx := context.Background()

	typeNamespacedName := types.NamespacedName{
		Name:      resourceName,
		Namespace: "default", // TODO(user):Modify as needed
	}
	branchprotection := &githubv1alpha1.BranchProtection{}
	testRepoName := ghTestResourcePrefix + "branch-protection-test-repo"

	Context("When reconciling a resource", func() {
		BeforeEach(func() {
			By("Creating the custom resource for the Kind BranchProtection")
			err := k8sClient.Get(ctx, typeNamespacedName, branchprotection)
			if err != nil && errors.IsNotFound(err) {
				resource := &githubv1alpha1.BranchProtection{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: githubv1alpha1.BranchProtectionSpec{
						RepositoryOwner: testOrganization,
						RepositoryName:  testRepoName,
						Pattern:         "main",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}

			By("Creating an external repository for testing")
			_, err = ghClient.CreateRepository(ctx, testOrganization, &github.Repository{
				Name: &testRepoName,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &githubv1alpha1.BranchProtection{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance BranchProtection")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			By("Cleanup the test repository")
			err = ghClient.DeleteRepositoryBySlug(ctx, testOrganization, testRepoName)
			Expect(err).NotTo(HaveOccurred())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &BranchProtectionReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})
})
