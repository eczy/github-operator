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
	"github.com/shurcooL/githubv4"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	githubv1alpha1 "github.com/eczy/github-operator/api/v1alpha1"
	gh "github.com/eczy/github-operator/internal/github"
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
	testBranchProtectionPattern := "master"

	var testRepository *github.Repository

	Context("When creating a BranchProtection resource", func() {
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
						Pattern:         testBranchProtectionPattern,
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}

			By("Creating a test repository")
			r, err := ghClient.CreateRepository(ctx, testOrganization, &github.Repository{
				Name:       &testRepoName,
				Visibility: github.String("public"),
			})
			Expect(err).NotTo(HaveOccurred())
			testRepository = r
		})

		AfterEach(func() {
			resource := &githubv1alpha1.BranchProtection{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())

			By("Cleanup the specific resource instance BranchProtection")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			By("Cleanup test repository")
			Expect(ghClient.DeleteRepositoryByName(ctx, testOrganization, testRepoName)).To(Succeed())
			testRepository = nil
		})

		It("should create a new BranchProtection resource and a new GitHub branch protection", func() {
			resource := &githubv1alpha1.BranchProtection{}

			By("Checking the GitHub branch protection doesn't exist")
			_, err := ghClient.GetBranchProtectionByOwnerRepoPattern(ctx, testOrganization, testRepoName, testBranchProtectionPattern)
			Expect(err).To(HaveOccurred())

			By("Reconciling the resource")
			controllerReconciler := &BranchProtectionReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the BranchProtection Status")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			Expect(resource.Status.NodeId).NotTo(Equal(nil))

			By("Checking the GitHub branch protection rule exists")
			_, err = ghClient.GetBranchProtectionByOwnerRepoPattern(ctx, testOrganization, testRepoName, testBranchProtectionPattern)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should create a new BranchProtection resource managing an existing GitHub branch protecetion", func() {
			resource := &githubv1alpha1.BranchProtection{}

			By("Creating a matching GitHub branch protection")
			ghBp, err := ghClient.CreateBranchProtection(ctx, &githubv4.CreateBranchProtectionRuleInput{
				RepositoryID: testRepository.GetNodeID(),
				Pattern:      githubv4.String(testBranchProtectionPattern),
			})
			Expect(err).NotTo(HaveOccurred())

			By("Reconciling the resource")
			controllerReconciler := &BranchProtectionReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the BranchProtection Status")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			Expect(resource.Status.NodeId).NotTo(BeNil())
			Expect(*resource.Status.NodeId).To(Equal(ghBp.Id))
		})
	})

	Context("When updating a BranchProtection resource", func() {
		var ghBp *gh.BranchProtection // temporarily store the created GitHub reference for each test
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
						Pattern:         testBranchProtectionPattern,
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}

			By("Creating a test repository")
			r, err := ghClient.CreateRepository(ctx, testOrganization, &github.Repository{
				Name:       &testRepoName,
				Visibility: github.String("public"),
			})
			Expect(err).NotTo(HaveOccurred())
			testRepository = r

			By("Creating a matching GitHub branch protection")
			bp, err := ghClient.CreateBranchProtection(ctx, &githubv4.CreateBranchProtectionRuleInput{
				RepositoryID: r.GetNodeID(),
				Pattern:      githubv4.String(testBranchProtectionPattern),
			})
			Expect(err).NotTo(HaveOccurred())
			ghBp = bp

			By("Associating the BranchProtection and GitHub branch protection")
			controllerReconciler := &BranchProtectionReconciler{
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
			resource := &githubv1alpha1.BranchProtection{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())

			By("Cleanup the specific resource instance BranchProtection")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			By("Cleanup test repository")
			Expect(ghClient.DeleteRepositoryByName(ctx, testOrganization, testRepoName)).To(Succeed())
			testRepository = nil
		})

		It("should successfully reconcile an updated BranchProtection pattern", func() {
			resource := &githubv1alpha1.BranchProtection{}
			By("Updating the BranchProtection resource Spec description")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			resource.Spec.Pattern = "master*"
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("Reconciling the resource")
			controllerReconciler := &BranchProtectionReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the BranchProtection resource Status")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			Expect(resource.Status.Pattern).NotTo(BeNil())
			Expect(*resource.Status.Pattern).To(Equal("master*"))

			By("Checking the GitHub branch protection")
			ghBp, err := ghClient.GetBranchProtection(ctx, ghBp.Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(ghBp.Pattern).To(Equal("master*"))
		})
		// TODO: other fields
	})

	Context("When deleting a BranchProtection resource", func() {
		var ghBp *gh.BranchProtection // temporarily store the created GitHub reference for each test
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
						Pattern:         testBranchProtectionPattern,
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}

			By("Creating a test repository")
			r, err := ghClient.CreateRepository(ctx, testOrganization, &github.Repository{
				Name:       &testRepoName,
				Visibility: github.String("public"),
			})
			Expect(err).NotTo(HaveOccurred())
			testRepository = r

			By("Creating a matching GitHub branch protection")
			bp, err := ghClient.CreateBranchProtection(ctx, &githubv4.CreateBranchProtectionRuleInput{
				RepositoryID: r.GetNodeID(),
				Pattern:      githubv4.String(testBranchProtectionPattern),
			})
			Expect(err).NotTo(HaveOccurred())
			ghBp = bp
		})

		AfterEach(func() {
			By("Check the specific resource instance BranchProtection is deleted")
			resource := &githubv1alpha1.BranchProtection{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).NotTo(Succeed())

			By("Cleanup test repository")
			Expect(ghClient.DeleteRepositoryByName(ctx, testOrganization, testRepoName)).To(Succeed())
			testRepository = nil
		})

		It("should delete a BranchProtection resource and a managed GitHub branch protection", func() {
			// when managed before deletion
			resource := &githubv1alpha1.BranchProtection{}

			By("Associating the BranchProtection resource with the GitHub branch protection")
			controllerReconciler := &BranchProtectionReconciler{
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
			controllerutil.AddFinalizer(resource, branchProtectionFinalizerName)
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("Deleting the resource")
			Expect(k8sClient.Delete(ctx, resource, &client.DeleteOptions{
				GracePeriodSeconds: &deletionGracePeriod,
			})).To(Succeed())
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the GitHub branch protection does not exist")
			_, err = ghClient.GetBranchProtection(ctx, ghBp.Id)
			Expect(err).To(HaveOccurred())
		})

		It("should delete a BranchProtection resource without affecting an unmanaged external resource", func() {
			// when not managed before deletion
			resource := &githubv1alpha1.BranchProtection{}

			By("Scheduling the resource for deletion")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			// manually add finalizer
			controllerutil.AddFinalizer(resource, branchProtectionFinalizerName)
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("Deleting the resource")
			Expect(k8sClient.Delete(ctx, resource, &client.DeleteOptions{
				GracePeriodSeconds: &deletionGracePeriod,
			})).To(Succeed())
			controllerReconciler := &BranchProtectionReconciler{
				Client:                   k8sClient,
				Scheme:                   k8sClient.Scheme(),
				GitHubClient:             ghClient,
				DeleteOnResourceDeletion: true,
			}
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the GitHub branch protection still exists")
			_, err = ghClient.GetBranchProtection(ctx, ghBp.Id)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should delete a BranchProtection resource when there is no matching GitHub branch protection", func() {
			resource := &githubv1alpha1.BranchProtection{}

			By("Checking there is no matching GitHub branch protection")
			Expect(ghClient.DeleteBranchProtection(ctx, &githubv4.DeleteBranchProtectionRuleInput{
				BranchProtectionRuleID: ghBp.Id,
			})).To(Succeed(), "this may change if BeforeEach is modified")

			By("Scheduling the resource for deletion")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			// manually add finalizer
			controllerutil.AddFinalizer(resource, branchProtectionFinalizerName)
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("Deleting the resource")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			Expect(k8sClient.Delete(ctx, resource, &client.DeleteOptions{
				GracePeriodSeconds: &deletionGracePeriod,
			})).To(Succeed())
			controllerReconciler := &BranchProtectionReconciler{
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

		It("should delete a BranchProtection resource without deleting the GitHub branch protection", func() {
			// when DeleteOnResourceDeletion isn't enabled
			resource := &githubv1alpha1.BranchProtection{}

			By("Fetching the current resource")
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())

			By("Associating the BranchProtection resource with the GitHub branch protection")
			controllerReconciler := &BranchProtectionReconciler{
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

			By("Checking the GitHub branch protection still exists")
			_, err = ghClient.GetBranchProtection(ctx, ghBp.Id)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
