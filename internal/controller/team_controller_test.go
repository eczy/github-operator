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

	githubv1alpha1 "github.com/eczy/github-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var testTeamName string = ghTestResourcePrefix + "team0"

var _ = Describe("Team Controller", func() {
	const resourceName = "test-resource"

	ctx := context.Background()

	typeNamespacedName := types.NamespacedName{
		Name:      resourceName,
		Namespace: "default", // TODO(user):Modify as needed
	}

	team := &githubv1alpha1.Team{}

	Context("When creating a resource", func() {
		BeforeEach(func() {
			By("creating the custom resource for the Kind Team")
			err := k8sClient.Get(ctx, typeNamespacedName, team)
			if err != nil && errors.IsNotFound(err) {
				resource := &githubv1alpha1.Team{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: githubv1alpha1.TeamSpec{
						Organization: testOrganization,
						Name:         testTeamName,
						Description:  github.String("foo"),
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &githubv1alpha1.Team{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Team")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			By("Cleanup the external resource")
			err = ghClient.DeleteTeamBySlug(ctx, testOrganization, testTeamName)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should create a new resource and external resource", func() {
			By("Checking the external resource doesn't exist")
			_, err := ghClient.GetTeamBySlug(ctx, testOrganization, testTeamName)
			Expect(err).To(HaveOccurred())

			By("Creating the resource")
			controllerReconciler := &TeamReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
			By("Checking the resource Status")
			resource := &githubv1alpha1.Team{}
			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.Status.Id).NotTo(Equal(nil))

			By("Checking the external resource exists")
			_, err = ghClient.GetTeamBySlug(ctx, testOrganization, testTeamName)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should create a new resource managing an external resource", func() {
			By("Creating a matching external resource")
			_, err := ghClient.CreateTeam(ctx, testOrganization, github.NewTeam{
				Name:        testTeamName,
				Description: github.String("foo"),
			})
			Expect(err).NotTo(HaveOccurred())

			By("Creating the resource")
			controllerReconciler := &TeamReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the resource Status")
			resource := &githubv1alpha1.Team{}
			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.Status.Id).NotTo(Equal(nil))

			By("Checking the matching external resource is now managed")
			_, err = ghClient.GetTeamBySlug(ctx, testOrganization, testTeamName)
			Expect(err).NotTo(HaveOccurred())
			// TODO: find a way to test this with  actual GitHub client without list all
			if c, ok := ghClient.(*TestGitHubClient); ok {
				Expect(len(c.OrgsBySlug[testOrganization].TeamBySlug)).To(Equal(1))
			}
		})
	})
	Context("When updating a resource", func() {
		BeforeEach(func() {
			By("creating the custom resource for the Kind Team")
			err := k8sClient.Get(ctx, typeNamespacedName, team)
			if err != nil && errors.IsNotFound(err) {
				resource := &githubv1alpha1.Team{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: githubv1alpha1.TeamSpec{
						Organization: testOrganization,
						Name:         testTeamName,
						Description:  github.String("foo"),
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
			By("creating the corresponding external resource")
			// everything but name is default
			// update in individual tests to test specific updates
			_, err = ghClient.CreateTeam(ctx, testOrganization, github.NewTeam{
				Name: testTeamName,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			resource := &githubv1alpha1.Team{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Team")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			By("Cleanup the external resource")
			err = ghClient.DeleteTeamBySlug(ctx, testOrganization, testTeamName)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should successfully reconcile an updated resource's description", func() {
			By("Reconciling the resource")
			controllerReconciler := &TeamReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the resource Status")
			resource := &githubv1alpha1.Team{}
			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.Status.Description).NotTo(BeNil())
			Expect(*resource.Status.Description).To(Equal("foo"))

			By("Checking the external resource")
			ghTeam, err := ghClient.GetTeamBySlug(ctx, testOrganization, testTeamName)
			Expect(err).NotTo(HaveOccurred())
			Expect(ghTeam.Description).NotTo(BeNil())
			Expect(*ghTeam.Description).To(Equal("foo"))
		})

		It("should successfully reconcile an updated resource's name", func() {
			By("Reconciling the resource (1/3)")
			controllerReconciler := &TeamReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Updating the resource Spec.Name")
			resource := &githubv1alpha1.Team{}
			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			originalName := resource.Spec.Name
			newName := ghTestResourcePrefix + "team1"
			resource.Spec.Name = newName
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("Reconciling the resource (2/3)")
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the external resource")
			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			ghTeam, err := ghClient.GetTeamBySlug(ctx, testOrganization, *resource.Status.Slug)
			Expect(err).NotTo(HaveOccurred())
			Expect(ghTeam.Name).NotTo(BeNil())
			Expect(*ghTeam.Name).To(Equal(newName))

			By("Restoring the resource Spec.Name")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			resource.Spec.Name = originalName
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("Reconciling the resource (3/3)")
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the resource Status")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())

			By("Checking the external resource")
			ghTeam, err = ghClient.GetTeamBySlug(ctx, testOrganization, *resource.Status.Slug)
			Expect(err).NotTo(HaveOccurred())
			Expect(ghTeam.Name).NotTo(BeNil())
			Expect(*ghTeam.Name).To(Equal(originalName))
		})
		// TODO: other fields
	})

	// TODO: test deletion

	Context("When deleting a resource", func() {
		BeforeEach(func() {
			By("Creating the custom resource for the Kind Team")
			err := k8sClient.Get(ctx, typeNamespacedName, team)
			if err != nil && errors.IsNotFound(err) {
				resource := &githubv1alpha1.Team{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: githubv1alpha1.TeamSpec{
						Organization: testOrganization,
						Name:         testTeamName,
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}

			By("Creating the corresponding external resource")
			_, err = ghClient.CreateTeam(ctx, testOrganization, github.NewTeam{
				Name: testTeamName,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			By("Check the specific resource instance Team is deleted")
			resource := &githubv1alpha1.Team{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).To(HaveOccurred())
		})

		It("should delete a resource and the associated external resource", func() {
			// when associated before deletion
			By("Associating the resource with the external resource")
			controllerReconciler := &TeamReconciler{
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
			resource := &githubv1alpha1.Team{}
			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())
			// manually add finalizer
			controllerutil.AddFinalizer(resource, teamFinalizerName)
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
			_, err = ghClient.GetTeamBySlug(ctx, testOrganization, testTeamName)
			Expect(err).To(HaveOccurred())
		})

		It("should delete a resource without affecting an unassociated external resource", func() {
			// i.e. if we are deleting on the first reconciliation of a resource, don't touch
			// external state
			By("Scheduling the resource for deletion")
			resource := &githubv1alpha1.Team{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())
			// manually add finalizer
			controllerutil.AddFinalizer(resource, teamFinalizerName)
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("Deleting the resource")
			Expect(k8sClient.Delete(ctx, resource, &client.DeleteOptions{
				GracePeriodSeconds: &deletionGracePeriod,
			})).To(Succeed())
			controllerReconciler := &TeamReconciler{
				Client:                   k8sClient,
				Scheme:                   k8sClient.Scheme(),
				GitHubClient:             ghClient,
				DeleteOnResourceDeletion: true,
			}
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the external resource still exists")
			_, err = ghClient.GetTeamBySlug(ctx, testOrganization, testTeamName)
			Expect(err).NotTo(HaveOccurred())

			By("Cleaning up the external resource")
			err = ghClient.DeleteTeamBySlug(ctx, testOrganization, testTeamName)
			Expect(err).NotTo(HaveOccurred(), "this may change if BeforeEach is modified")
		})

		It("should delete a resource when there is no external resource", func() {
			By("Checking there is no matching external resource")
			err := ghClient.DeleteTeamBySlug(ctx, testOrganization, testTeamName)
			Expect(err).NotTo(HaveOccurred(), "this may change if BeforeEach is modified")

			By("Deleting the resource")
			resource := &githubv1alpha1.Team{}
			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())
			Expect(k8sClient.Delete(ctx, resource, &client.DeleteOptions{
				GracePeriodSeconds: &deletionGracePeriod,
			})).To(Succeed())
			controllerReconciler := &TeamReconciler{
				Client:                   k8sClient,
				Scheme:                   k8sClient.Scheme(),
				GitHubClient:             ghClient,
				DeleteOnResourceDeletion: true,
			}
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
