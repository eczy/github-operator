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

	Context("When creating a resource of Kind BranchProtection", func() {
		var testRepo *github.Repository
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
			repo, err := ghClient.CreateRepository(ctx, testOrganization, &github.Repository{
				Name: &testRepoName,
			})
			Expect(err).NotTo(HaveOccurred())
			testRepo = repo
		})

		AfterEach(func() {
			resource := &githubv1alpha1.BranchProtection{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance BranchProtection")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			By("Cleanup the test repository")
			err = ghClient.DeleteRepositoryByName(ctx, testOrganization, testRepoName)
			Expect(err).NotTo(HaveOccurred())
		})
		It("should successfully reconcile the resource and create a new rule", func() {
			By("Checking the branch protection rule doesn't exist")
			_, err := ghClient.GetBranchProtectionByOwnerRepoPattern(ctx, testOrganization, testRepoName, "main")
			Expect(err).To(HaveOccurred())

			By("Reconciling the created resource")
			controllerReconciler := &BranchProtectionReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the branch protection rule exists")
			_, err = ghClient.GetBranchProtectionByOwnerRepoPattern(ctx, testOrganization, testRepoName, "main")
			Expect(err).NotTo(HaveOccurred())
		})

		It("should successfully reconcile the resource and associate with an existing rule", func() {
			By("Creating the branch protection rule")
			beforeBpr, err := ghClient.CreateBranchProtection(ctx, &githubv4.CreateBranchProtectionRuleInput{
				RepositoryID: testRepo.GetNodeID(),
				Pattern:      "main",
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the branch protection rule already exists")
			_, err = ghClient.GetBranchProtectionByOwnerRepoPattern(ctx, testOrganization, testRepoName, "main")
			Expect(err).NotTo(HaveOccurred())

			By("Reconciling the created resource")
			controllerReconciler := &BranchProtectionReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the branch protection is now associated with the resource")
			afterBpr, err := ghClient.GetBranchProtectionByOwnerRepoPattern(ctx, testOrganization, testRepoName, "main")
			Expect(err).NotTo(HaveOccurred())
			Expect(beforeBpr.Id).To(Equal(afterBpr.Id))
		})
	})
	Context("When updating a resource of Kind BranchProtection", func() {
		var testRepo *github.Repository
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
			repo, err := ghClient.CreateRepository(ctx, testOrganization, &github.Repository{
				Name: &testRepoName,
			})
			Expect(err).NotTo(HaveOccurred())
			testRepo = repo
		})

		AfterEach(func() {
			resource := &githubv1alpha1.BranchProtection{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance BranchProtection")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			By("Cleanup the test repository")
			err = ghClient.DeleteRepositoryByName(ctx, testOrganization, testRepoName)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should successfully reconcile the resource and update an existing rule's pattern", func() {
			By("Reconciling the resource (1/2)")
			controllerReconciler := &BranchProtectionReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking external resource properties")
			bp, err := ghClient.GetBranchProtectionByOwnerRepoPattern(ctx, testOrganization, testRepoName, "main")
			Expect(err).NotTo(HaveOccurred())
			Expect(bp.Pattern).To(Equal("main"))

			By("Updating the resource")
			resource := &githubv1alpha1.BranchProtection{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, resource)).To(Succeed())
			resource.Spec.Pattern = "master"
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("Reconciling the resource (2/2)")
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking external resource properties")
			bp, err = ghClient.GetBranchProtection(ctx, bp.Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(bp.Pattern).To(Equal("master"))
		})

		It("should successfully reconcile the resource and associate with an existing rule", func() {
			By("Creating the branch protection rule")
			beforeBpr, err := ghClient.CreateBranchProtection(ctx, &githubv4.CreateBranchProtectionRuleInput{
				RepositoryID: testRepo.GetNodeID(),
				Pattern:      "main",
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the branch protection rule already exists")
			_, err = ghClient.GetBranchProtectionByOwnerRepoPattern(ctx, testOrganization, testRepoName, "main")
			Expect(err).NotTo(HaveOccurred())

			By("Reconciling the created resource")
			controllerReconciler := &BranchProtectionReconciler{
				Client:       k8sClient,
				Scheme:       k8sClient.Scheme(),
				GitHubClient: ghClient,
			}

			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the branch protection is now associated with the resource")
			afterBpr, err := ghClient.GetBranchProtectionByOwnerRepoPattern(ctx, testOrganization, testRepoName, "main")
			Expect(err).NotTo(HaveOccurred())
			Expect(beforeBpr.Id).To(Equal(afterBpr.Id))
		})
	})
	Context("When deleting a resource of Kind BranchProtection", func() {
		var testRepo *github.Repository
		var testBranchProtection *gh.BranchProtection
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
			repo, err := ghClient.CreateRepository(ctx, testOrganization, &github.Repository{
				Name: &testRepoName,
			})
			Expect(err).NotTo(HaveOccurred())
			testRepo = repo

			By("Creating an external branch protection rule")
			bp, err := ghClient.CreateBranchProtection(ctx, &githubv4.CreateBranchProtectionRuleInput{
				RepositoryID: testRepo.GetNodeID(),
				Pattern:      "main",
			})
			Expect(err).NotTo(HaveOccurred())
			testBranchProtection = bp
		})

		AfterEach(func() {
			By("Cleanup the test repository")
			err := ghClient.DeleteRepositoryByName(ctx, testOrganization, testRepoName)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should successfully reconcile the resource and delete an external branch protection", func() {
			By("Reconciling the resource (1/2)")
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
			resource := &githubv1alpha1.BranchProtection{}
			err = k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())
			controllerutil.AddFinalizer(resource, branchProtectionFinalizerName)
			Expect(k8sClient.Update(ctx, resource)).To(Succeed())

			By("Reconciling the resource (2/2)")
			Expect(k8sClient.Delete(ctx, resource, &client.DeleteOptions{
				GracePeriodSeconds: &deletionGracePeriod,
			})).To(Succeed())
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking the external resource doesn't exist")
			_, err = ghClient.GetBranchProtection(ctx, testBranchProtection.Id)
			Expect(err).To(HaveOccurred())
		})
	})
})
