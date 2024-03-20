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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	githubv1alpha1 "github.com/eczy/github-operator/api/v1alpha1"
)

var _ = Describe("Repository Controller", func() {
	const resourceName = "test-resource"

	ctx := context.Background()

	typeNamespacedName := types.NamespacedName{
		Name:      resourceName,
		Namespace: "default",
	}
	repository := &githubv1alpha1.Repository{}

	testRepositoryName := "testrepo"

	Context("When creating a resource", func() {
		BeforeEach(func() {
			By("creating the custom resource for the Kind Repository")
			err := k8sClient.Get(ctx, typeNamespacedName, repository)
			if err != nil && errors.IsNotFound(err) {
				resource := &githubv1alpha1.Repository{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: githubv1alpha1.RepositorySpec{
						Name:  testRepositoryName,
						Owner: testOrganization,
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &githubv1alpha1.Repository{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Repository")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			By("Cleanup the external resource")
			Expect(ghClient.DeleteRepositoryBySlug(ctx, testOrganization, testRepositoryName))
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &RepositoryReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking resource Status")
			resource := &githubv1alpha1.Repository{}
			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.Status.LastUpdateTimestamp.Time).NotTo(BeNil())
			Expect(resource.Status.Name).NotTo(BeNil())
			Expect(*resource.Status.Name).To(Equal(testRepositoryName))
			Expect(resource.Status.Id).NotTo(BeNil())

			By("Checking external resource")
			ghRepo, err := ghClient.GetRepositoryBySlug(ctx, testOrganization, testRepositoryName)
			Expect(err).NotTo(HaveOccurred())
			Expect(ghRepo.Name).NotTo(BeNil())
			Expect(*ghRepo.Name).To(Equal(testRepositoryName))
		})
	})
	Context("When updating a resource", func() {
		BeforeEach(func() {
			By("Creating the custom resource for the Kind Repository")
			err := k8sClient.Get(ctx, typeNamespacedName, repository)
			if err != nil && errors.IsNotFound(err) {
				resource := &githubv1alpha1.Repository{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: githubv1alpha1.RepositorySpec{
						Name:  testRepositoryName,
						Owner: testOrganization,
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}

			By("Setup the external resource")
			_, err = ghClient.CreateRepository(ctx, testOrganization, &github.Repository{
				Name:        &testRepositoryName,
				Description: github.String("foo"),
			})
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			resource := &githubv1alpha1.Repository{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Repository")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			By("Cleanup the external resource")
			Expect(ghClient.DeleteRepositoryBySlug(ctx, testOrganization, testRepositoryName))
		})

		It("should successfully update the resource Name", func() {
			newName := testRepositoryName + "-foo"

			// need to associate the repo before modifying
			By("Reconciling the resource (1/3)")
			controllerReconciler := &RepositoryReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			By("Updating the resource Spec.Name")
			resource := &githubv1alpha1.Repository{}
			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())
			resource.Spec.Name = newName
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("Reconciling the resource (2/3)")
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			By("Checking resource Status")
			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.Status.Name).NotTo(BeNil())
			Expect(*resource.Status.Name).To(Equal(newName))
			Expect(resource.Status.Id).NotTo(BeNil())

			By("Checking external resource")
			ghRepo, err := ghClient.GetRepositoryBySlug(ctx, testOrganization, newName)
			Expect(err).NotTo(HaveOccurred())
			Expect(ghRepo.Name).NotTo(BeNil())

			By("Updating the resource Spec.Name")
			resource.Spec.Name = testRepositoryName
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("Reconciling the resource (3/3)")
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})

			Expect(err).NotTo(HaveOccurred())
			By("Checking resource Status")
			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.Status.Name).NotTo(BeNil())
			Expect(*resource.Status.Name).To(Equal(testRepositoryName))
			Expect(resource.Status.Id).NotTo(BeNil())

			By("Checking external resource")
			ghRepo, err = ghClient.GetRepositoryBySlug(ctx, testOrganization, testRepositoryName)
			Expect(err).NotTo(HaveOccurred())
			Expect(ghRepo.Name).NotTo(BeNil())
		})

		It("should successfully update the resource Description", func() {
			newDescription := "bar"
			By("Updating the resource Spec.Name")
			resource := &githubv1alpha1.Repository{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())
			resource.Spec.Description = &newDescription
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("Reconciling the resource")
			controllerReconciler := &RepositoryReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			By("Checking resource Status")
			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.Status.Description).NotTo(BeNil())
			Expect(*resource.Status.Description).To(Equal(newDescription))
			Expect(resource.Status.Id).NotTo(BeNil())

			By("Checking external resource")
			ghRepo, err := ghClient.GetRepositoryBySlug(ctx, testOrganization, testRepositoryName)
			Expect(err).NotTo(HaveOccurred())
			Expect(ghRepo.Description).NotTo(BeNil())
			Expect(*ghRepo.Description).To(Equal(newDescription))
		})
	})
	Context("When deleting a resource", func() {
		BeforeEach(func() {
			By("Creating the custom resource for the Kind Repository")
			err := k8sClient.Get(ctx, typeNamespacedName, repository)
			if err != nil && errors.IsNotFound(err) {
				resource := &githubv1alpha1.Repository{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: githubv1alpha1.RepositorySpec{
						Name:  testRepositoryName,
						Owner: testOrganization,
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}

			By("Setup the external resource")
			_, err = ghClient.CreateRepository(ctx, testOrganization, &github.Repository{
				Name: &testRepositoryName,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should successfully reconcile the resource", func() {
			// when associated before deletion
			By("Associating the resource with the external resource")
			controllerReconciler := &RepositoryReconciler{
				Client:                   k8sClient,
				Scheme:                   k8sClient.Scheme(),
				GitHubClient:             ghClient,
				DeleteOnResourceDeletion: true,
			}
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Scheduling the resource for deletion")
			resource := &githubv1alpha1.Repository{}
			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())
			// manually add finalizer
			controllerutil.AddFinalizer(resource, repositoryFinalizerName)
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("Deleting the resource")
			Expect(k8sClient.Delete(ctx, resource, &client.DeleteOptions{
				GracePeriodSeconds: &deletionGracePeriod,
			})).To(Succeed())
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the external resource does not exist")
			_, err = ghClient.GetRepositoryBySlug(ctx, testOrganization, testRepositoryName)
			Expect(err).To(HaveOccurred())
		})
	})
})
