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

var _ = Describe("Organization Controller", func() {
	const resourceName = "test-resource"

	ctx := context.Background()

	typeNamespacedName := types.NamespacedName{
		Name:      resourceName,
		Namespace: "default",
	}

	organization := &githubv1alpha1.Organization{}
	beforeState := &github.Organization{}

	Context("When updating an Organization resource", func() {
		BeforeEach(func() {
			By("Creating the custom resource for the Kind Organization")
			err := k8sClient.Get(ctx, typeNamespacedName, organization)
			if err != nil && errors.IsNotFound(err) {
				resource := &githubv1alpha1.Organization{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: githubv1alpha1.OrganizationSpec{
						Login:        testOrganization,
						Name:         testOrganization,
						BillingEmail: "fakeemailfoobar@gmail.com",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
			By("Associating the Organization and GitHub organization")
			controllerReconciler := &OrganizationReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Storing the preexisting organization state")
			o, err := ghClient.GetOrganization(ctx, testOrganization)
			Expect(err).NotTo(HaveOccurred())
			o.MembersCanCreateInternalRepos = nil
			beforeState = o
		})

		AfterEach(func() {
			resource := &githubv1alpha1.Organization{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())

			By("Cleanup the specific resource instance Organization")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			By("Restoring the preexisting organization state")
			_, err := ghClient.UpdateOrganization(ctx, testOrganization, beforeState)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should successfully reconcile an updated Organization description", func() {
			resource := &githubv1alpha1.Organization{}
			By("Updating the Organization resource Spec description")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			resource.Spec.Description = github.String("foobar")
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("Reconciling the resource")
			controllerReconciler := &OrganizationReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the Organization resource Status")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			Expect(resource.Status.Description).NotTo(BeNil())
			Expect(*resource.Status.Description).To(Equal("foobar"))

			By("Checking the GitHub organization")
			ghOrg, err := ghClient.GetOrganization(ctx, testOrganization)
			Expect(err).NotTo(HaveOccurred())
			Expect(ghOrg.Description).NotTo(BeNil())
			Expect(*ghOrg.Description).To(Equal("foobar"))
		})

		It("should successfully reconcile an updated Organization name", func() {
			resource := &githubv1alpha1.Organization{}
			newName := ghTestResourcePrefix + "foo"

			By("Updating the Organization resource Spec name")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			resource.Spec.Name = newName
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("Reconciling the resource")
			controllerReconciler := &OrganizationReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the Organization resource Status")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			Expect(resource.Status.Name).To(Equal(newName))

			By("Checking the GitHub organization")
			ghOrg, err := ghClient.GetOrganization(ctx, testOrganization)
			Expect(err).NotTo(HaveOccurred())
			Expect(ghOrg.GetName()).To(Equal(newName))
		})
		// TODO: other fields
	})
})
