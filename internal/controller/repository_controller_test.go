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
	testRepositoryName := ghTestResourcePrefix + "testrepo"

	Context("When creating a Repository resource", func() {
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
		})

		AfterEach(func() {
			resource := &githubv1alpha1.Repository{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())

			By("Cleanup the specific resource instance Repository")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should create a new Repository resource and a new GitHub repository", func() {
			resource := &githubv1alpha1.Repository{}

			By("Checking the GitHub repository doesn't exist")
			_, err := ghClient.GetRepositoryByName(ctx, testOrganization, testRepositoryName)
			Expect(err).To(HaveOccurred())

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

			By("Checking the Repository Status")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			Expect(resource.Status.NodeId).NotTo(Equal(nil))

			By("Checking the GitHub Repository exists")
			_, err = ghClient.GetRepositoryByName(ctx, testOrganization, testRepositoryName)
			Expect(err).NotTo(HaveOccurred())

			By("Cleaning up the GitHub repository")
			Expect(ghClient.DeleteRepositoryByName(ctx, testOrganization, testRepositoryName)).To(Succeed())
		})

		It("should create a new Repository resource managing an existing GitHub repository", func() {
			resource := &githubv1alpha1.Repository{}

			By("Creating a matching GitHub repository")
			ghRepo, err := ghClient.CreateRepository(ctx, testOrganization, &github.Repository{
				Name: github.String(testRepositoryName),
			})
			Expect(err).NotTo(HaveOccurred())

			defer func() {
				By("Cleaning up the GitHub repository")
				Expect(ghClient.DeleteRepositoryByName(ctx, testOrganization, testRepositoryName)).To(Succeed())
			}()

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

			By("Checking the Repository Status")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			Expect(resource.Status.NodeId).NotTo(BeNil())
			Expect(*resource.Status.NodeId).To(Equal(ghRepo.GetNodeID()))
		})
	})

	Context("When updating a Repository resource", func() {
		var ghRepository *github.Repository // temporarily store the created GitHub reference for each test
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

			By("Creating a matching GitHub repository")
			r, err := ghClient.CreateRepository(ctx, testOrganization, &github.Repository{
				Name: &testRepositoryName,
			})
			Expect(err).NotTo(HaveOccurred())
			ghRepository = r

			By("Associating the Repository and GitHub repository")
			controllerReconciler := &RepositoryReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			resource := &githubv1alpha1.Repository{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Repository")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			By("Cleanup the matching GitHub repository")
			Expect(ghClient.DeleteRepositoryByName(ctx, ghRepository.GetOrganization().GetLogin(), ghRepository.GetName())).To(Succeed())
			ghRepository = nil
		})

		It("should successfully reconcile an updated Repository description", func() {
			resource := &githubv1alpha1.Repository{}
			By("Updating the Repository resource Spec description")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			resource.Spec.Description = github.String("foobar")
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("Reconciling the resource")
			controllerReconciler := &RepositoryReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the Repository resource Status")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			Expect(resource.Status.Description).NotTo(BeNil())
			Expect(*resource.Status.Description).To(Equal("foobar"))

			By("Checking the GitHub repository")
			ghRepository, err := ghClient.GetRepositoryByNodeId(ctx, ghRepository.GetNodeID())
			Expect(err).NotTo(HaveOccurred())
			Expect(ghRepository.Description).NotTo(BeNil())
			Expect(*ghRepository.Description).To(Equal("foobar"))
		})

		It("should successfully reconcile an updated Repository name", func() {
			resource := &githubv1alpha1.Repository{}
			newName := ghTestResourcePrefix + "foo"

			By("Updating the Repository resource Spec name")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			resource.Spec.Name = newName
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("Reconciling the resource")
			controllerReconciler := &RepositoryReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the Repository resource Status")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			Expect(resource.Status.Name).NotTo(BeNil())
			Expect(*resource.Status.Name).To(Equal(newName))

			By("Checking the GitHub repository")
			ghRepository, err := ghClient.GetRepositoryByNodeId(ctx, ghRepository.GetNodeID())
			Expect(err).NotTo(HaveOccurred())
			Expect(ghRepository.GetName()).To(Equal(newName))
		})
		// TODO more fields
	})

	Context("When deleting a resource", func() {
		var ghRepository *github.Repository // temporarily store the created GitHub reference for each test
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

			By("Creating a matching GitHub repository")
			r, err := ghClient.CreateRepository(ctx, testOrganization, &github.Repository{
				Name: &testRepositoryName,
			})
			Expect(err).NotTo(HaveOccurred())
			ghRepository = r
		})

		AfterEach(func() {
			By("Check the specific resource instance Repository is deleted")
			resource := &githubv1alpha1.Repository{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).NotTo(Succeed())
		})

		It("should delete a Repository resource and a managed GitHub repository", func() {
			// when managed before deletion
			resource := &githubv1alpha1.Repository{}

			By("Associating the Repository resource with the GitHub repository")
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
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
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

			By("Checking the GitHub repository does not exist")
			_, err = ghClient.GetRepositoryByNodeId(ctx, ghRepository.GetNodeID())
			Expect(err).To(HaveOccurred())
		})

		It("should delete a Repository resource without affecting an unmanaged external repository", func() {
			// when not managed before deletion
			resource := &githubv1alpha1.Repository{}

			By("Scheduling the resource for deletion")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			// manually add finalizer
			controllerutil.AddFinalizer(resource, repositoryFinalizerName)
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("Deleting the resource")
			Expect(k8sClient.Delete(ctx, resource, &client.DeleteOptions{
				GracePeriodSeconds: &deletionGracePeriod,
			})).To(Succeed())
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

			By("Checking the GitHub repository still exists")
			_, err = ghClient.GetRepositoryByNodeId(ctx, ghRepository.GetNodeID())
			Expect(err).NotTo(HaveOccurred())

			By("Cleaning up the GitHub repository")
			Expect(ghClient.DeleteRepositoryByName(ctx, testOrganization, testRepositoryName)).To(Succeed(), "this may change if BeforeEach is modified")
		})

		It("should delete a Repository resource when there is no matching GitHub repository", func() {
			resource := &githubv1alpha1.Repository{}

			By("Checking there is no matching GitHub repository")
			Expect(ghClient.DeleteRepositoryByName(ctx, ghRepository.Organization.GetLogin(), ghRepository.GetName())).To(Succeed(), "this may change if BeforeEach is modified")

			By("Scheduling the resource for deletion")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			// manually add finalizer
			controllerutil.AddFinalizer(resource, repositoryFinalizerName)
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("Deleting the resource")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			Expect(k8sClient.Delete(ctx, resource, &client.DeleteOptions{
				GracePeriodSeconds: &deletionGracePeriod,
			})).To(Succeed())
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
		})

		It("should delete a Repository resource without deleting the GitHub repository", func() {
			// when DeleteOnResourceDeletion isn't enabled
			resource := &githubv1alpha1.Repository{}

			By("Fetching the current resource")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())

			By("Associating the Repository resource with the GitHub repository")
			controllerReconciler := &RepositoryReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			// don't add the finalizer manually since it's only added to the resource when DeleteOnResourceDeletion is enabled
			// IF THIS BEHAVIOR CHANGES, THIS TEST NEEDS TO BE UPDATED

			By("Deleting the resource")
			Expect(k8sClient.Delete(ctx, resource, &client.DeleteOptions{
				GracePeriodSeconds: &deletionGracePeriod,
			})).To(Succeed())
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the GitHub repository still exists")
			_, err = ghClient.GetRepositoryByNodeId(ctx, ghRepository.GetNodeID())
			Expect(err).NotTo(HaveOccurred())

			By("Cleaning up the GitHub repository")
			Expect(ghClient.DeleteRepositoryByName(ctx, ghRepository.GetOrganization().GetLogin(), ghRepository.GetName())).To(Succeed())
		})
	})
})
