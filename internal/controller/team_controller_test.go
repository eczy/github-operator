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

var _ = Describe("Team Controller", func() {
	const resourceName = "test-resource"

	ctx := context.Background()

	typeNamespacedName := types.NamespacedName{
		Name:      resourceName,
		Namespace: "default", // TODO(user):Modify as needed
	}

	team := &githubv1alpha1.Team{}
	testTeamName := ghTestResourcePrefix + "team0"

	Context("When creating a Team resource", func() {
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
		})

		AfterEach(func() {
			resource := &githubv1alpha1.Team{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())

			By("Cleanup the specific resource instance Team")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should create a new Team resource and a new GitHub team", func() {
			resource := &githubv1alpha1.Team{}

			By("Checking the GitHub team doesn't exist")
			_, err := ghClient.GetTeamBySlug(ctx, testOrganization, testTeamName)
			Expect(err).To(HaveOccurred())

			By("Reconciling the resource")
			controllerReconciler := &TeamReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the Team Status")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			Expect(resource.Status.NodeId).NotTo(Equal(nil))

			By("Checking the GitHub team exists")
			_, err = ghClient.GetTeamBySlug(ctx, testOrganization, testTeamName)
			Expect(err).NotTo(HaveOccurred())

			By("Cleaning up the GitHub team")
			Expect(ghClient.DeleteTeamBySlug(ctx, testOrganization, testTeamName)).To(Succeed())
		})

		It("should create a new Team resource managing an existing GitHub team", func() {
			resource := &githubv1alpha1.Team{}

			By("Creating a matching GitHub team")
			ghTeam, err := ghClient.CreateTeam(ctx, testOrganization, github.NewTeam{
				Name: testTeamName,
			})
			Expect(err).NotTo(HaveOccurred())

			defer func() {
				By("Cleaning up the GitHub team")
				Expect(ghClient.DeleteTeamBySlug(ctx, testOrganization, testTeamName)).To(Succeed())
			}()

			By("Reconciling the resource")
			controllerReconciler := &TeamReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the Team Status")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			Expect(resource.Status.NodeId).NotTo(BeNil())
			Expect(*resource.Status.NodeId).To(Equal(ghTeam.GetNodeID()))
		})
	})

	Context("When updating a Team resource", func() {
		var ghTeam *github.Team // temporarily store the created GitHub reference for each test
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
			By("Creating a matching GitHub team")
			t, err := ghClient.CreateTeam(ctx, testOrganization, github.NewTeam{
				Name: testTeamName,
			})
			Expect(err).NotTo(HaveOccurred())
			ghTeam = t

			By("Associating the Team and GitHub team")
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

		AfterEach(func() {
			resource := &githubv1alpha1.Team{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())

			By("Cleanup the specific resource instance Team")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			By("Cleanup the matching GitHub team")
			Expect(ghClient.DeleteTeamById(ctx, ghTeam.GetOrganization().GetID(), ghTeam.GetID())).To(Succeed())
			ghTeam = nil
		})

		It("should successfully reconcile an updated Team description", func() {
			resource := &githubv1alpha1.Team{}
			By("Updating the Team resource Spec description")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			resource.Spec.Description = github.String("foobar")
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

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

			By("Checking the Team resource Status")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			Expect(resource.Status.Description).NotTo(BeNil())
			Expect(*resource.Status.Description).To(Equal("foobar"))

			By("Checking the GitHub team")
			ghTeam, err := ghClient.GetTeamByNodeId(ctx, ghTeam.GetNodeID())
			Expect(err).NotTo(HaveOccurred())
			Expect(ghTeam.Description).NotTo(BeNil())
			Expect(*ghTeam.Description).To(Equal("foobar"))
		})

		It("should successfully reconcile an updated Team name", func() {
			resource := &githubv1alpha1.Team{}
			newName := ghTestResourcePrefix + "foo"

			By("Updating the Team resource Spec name")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			resource.Spec.Name = newName
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

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

			By("Checking the Team resource Status")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			Expect(resource.Status.Name).NotTo(BeNil())
			Expect(*resource.Status.Name).To(Equal(newName))

			By("Checking the GitHub team")
			ghTeam, err := ghClient.GetTeamByNodeId(ctx, ghTeam.GetNodeID())
			Expect(err).NotTo(HaveOccurred())
			Expect(ghTeam.GetName()).To(Equal(newName))
		})

		It("should successfully reconcile an updated Team repository permissions", func() {
			resource := &githubv1alpha1.Team{}

			By("Creating a new repository to test permissions")
			testRepo, err := ghClient.CreateRepository(ctx, testOrganization, &github.Repository{
				Name:       github.String(ghTestResourcePrefix + "team-test"),
				Visibility: github.String("private"),
			})
			Expect(err).NotTo(HaveOccurred())

			defer func() {
				By("Cleaning up the test repository")
				Expect(ghClient.DeleteRepositoryByName(ctx, testRepo.GetOwner().GetLogin(), testRepo.GetName())).To(Succeed())
			}()

			By("Updating the Team resource Spec repositories")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			resource.Spec.Repositories = map[string]githubv1alpha1.RepositoryPermission{
				testRepo.GetName(): "pull",
			}
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("Reconciling the resource")
			controllerReconciler := &TeamReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the Team resource Status")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			Expect(resource.Status.Repositories).Should(HaveKeyWithValue(testRepo.GetName(), (githubv1alpha1.RepositoryPermission)("pull")))

			By("Checking the GitHub team")
			rp, err := ghClient.GetTeamRepositoryPermission(ctx, testOrganization, testTeamName, testRepo.GetName())
			Expect(err).NotTo(HaveOccurred())
			Expect(rp.Permission).To(Equal("pull"))
		})
		// TODO: other fields
	})

	Context("When deleting a Team resource", func() {
		var ghTeam *github.Team // temporarily store the created GitHub reference for each test
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

			By("Creating a matching GitHub team")
			t, err := ghClient.CreateTeam(ctx, testOrganization, github.NewTeam{
				Name: testTeamName,
			})
			Expect(err).NotTo(HaveOccurred())
			ghTeam = t
		})

		AfterEach(func() {
			By("Check the specific resource instance Team is deleted")
			resource := &githubv1alpha1.Team{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).NotTo(Succeed())
		})

		It("should delete a Team resource and a managed GitHub team", func() {
			// when managed before deletion
			resource := &githubv1alpha1.Team{}

			By("Associating the Team resource with the GitHub team")
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
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
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

			By("Checking the GitHub team does not exist")
			_, err = ghClient.GetTeamByNodeId(ctx, ghTeam.GetNodeID())
			Expect(err).To(HaveOccurred())
		})

		It("should delete a Team resource without affecting an unmanaged external resource", func() {
			// when not managed before deletion
			resource := &githubv1alpha1.Team{}

			By("Scheduling the resource for deletion")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
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
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the GitHub team still exists")
			_, err = ghClient.GetTeamByNodeId(ctx, ghTeam.GetNodeID())
			Expect(err).NotTo(HaveOccurred())

			By("Cleaning up the GitHub team")
			Expect(ghClient.DeleteTeamBySlug(ctx, testOrganization, testTeamName)).To(Succeed(), "this may change if BeforeEach is modified")
		})

		It("should delete a Team resource when there is no matching GitHub team", func() {
			resource := &githubv1alpha1.Team{}

			By("Checking there is no matching GitHub team")
			Expect(ghClient.DeleteTeamById(ctx, ghTeam.Organization.GetID(), ghTeam.GetID())).To(Succeed(), "this may change if BeforeEach is modified")

			By("Scheduling the resource for deletion")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			// manually add finalizer
			controllerutil.AddFinalizer(resource, teamFinalizerName)
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("Deleting the resource")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			Expect(k8sClient.Delete(ctx, resource, &client.DeleteOptions{
				GracePeriodSeconds: &deletionGracePeriod,
			})).To(Succeed())
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
		})

		It("should delete a Team resource without deleting the GitHub team", func() {
			// when DeleteOnResourceDeletion isn't enabled
			resource := &githubv1alpha1.Team{}

			By("Fetching the current resource")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())

			By("Associating the Team resource with the GitHub team")
			controllerReconciler := &TeamReconciler{
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

			By("Checking the GitHub team still exists")
			_, err = ghClient.GetTeamByNodeId(ctx, ghTeam.GetNodeID())
			Expect(err).NotTo(HaveOccurred())

			By("Cleaning up the GitHub team")
			Expect(ghClient.DeleteTeamById(ctx, ghTeam.GetOrganization().GetID(), ghTeam.GetID())).To(Succeed())
		})
	})
})
