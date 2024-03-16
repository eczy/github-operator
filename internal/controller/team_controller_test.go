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

	githubv1alpha1 "github.com/eczy/github-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Team Controller", func() {
	const resourceName = "test-resource"

	ctx := context.Background()

	typeNamespacedName := types.NamespacedName{
		Name:      resourceName,
		Namespace: "default", // TODO(user):Modify as needed
	}
	team := &githubv1alpha1.Team{}

	testGitHubResourcePrefix := "github-operator-test-"
	testOrganization := "testorg"
	testTeamName := testGitHubResourcePrefix + "team0"
	// TODO: set this from env
	mock := true
	var ghClient *TestGitHubClient
	if mock {
		ghClient = NewTestGitHubClient(WithTestOrganization(*NewTestOrganization(testOrganization, 0)))
	} else {
		// TODO: real github client with creds
		ghClient = NewTestGitHubClient()
	}
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
			Expect(len(ghClient.OrgsBySlug[testOrganization].TeamBySlug)).To(Equal(1))
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
			Expect(resource.Status.LastUpdateTimestamp).NotTo(Equal(nil))
			Expect(resource.Status.Description).To(Equal("foo"))

			By("Checking the external resource")
			ghTeam, err := ghClient.GetTeamBySlug(ctx, testOrganization, testTeamName)
			Expect(err).NotTo(HaveOccurred())
			Expect(ghTeam.Description).NotTo(BeNil())
			Expect(*ghTeam.Description).To(Equal("foo"))
		})

		It("should successfully reconcile an updated resource's name", func() {
			By("Reconciling the resource (1/2)")
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

			newName := testGitHubResourcePrefix + "team1"
			resource.Spec.Name = newName
			firstUpdateTimestamp := resource.Status.LastUpdateTimestamp
			Expect(firstUpdateTimestamp).ToNot(BeNil())

			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("Reconciling the resource (2/2)")
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the resource Status")
			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())
			Expect(resource.Status.LastUpdateTimestamp.After(firstUpdateTimestamp.Time)).To(BeTrue())
			Expect(resource.Status.Name).To(Equal(newName))

			By("Checking the external resource")
			ghTeam, err := ghClient.GetTeamBySlug(ctx, testOrganization, *team.Status.Slug)
			Expect(err).NotTo(HaveOccurred())
			Expect(ghTeam.Name).To(Equal(newName))
		})
		// TODO: other fields
	})

	Context("When deleting a resource", func() {
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
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}

			By("creating the corresponding external resource")
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
		})

		It("should delete a resource and the associated external resource", func() {
			// when associated before deletion
			By("associating the resource with the external resource")
			controllerReconciler := &TeamReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("scheduling the resource for deletion")
			resource := &githubv1alpha1.Team{}
			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())
			now := metav1.Now()
			resource.SetDeletionTimestamp(&now)
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("deleting the resource")
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the external resource does not exist")
			_, err = ghClient.GetTeamBySlug(ctx, testOrganization, testTeamName)
			Expect(err).To(HaveOccurred())
		})

		It("should delete a resource and the previously unassociated external resource", func() {
			// when NOT associated before deletion
			By("scheduling the resource for deletion")
			resource := &githubv1alpha1.Team{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())
			now := metav1.Now()
			resource.SetDeletionTimestamp(&now)
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("deleting the resource")
			controllerReconciler := &TeamReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the external resource does not exist")
			_, err = ghClient.GetTeamBySlug(ctx, testOrganization, testTeamName)
			Expect(err).To(HaveOccurred())
		})

		It("should delete a resource when there is no external resource", func() {
			By("scheduling the resource for deletion")
			resource := &githubv1alpha1.Team{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())
			now := metav1.Now()
			resource.SetDeletionTimestamp(&now)
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("deleting the resource")
			controllerReconciler := &TeamReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
