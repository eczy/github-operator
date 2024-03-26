package github

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
)

// Branch Protection

// TODO: consistent API - either direct access or getters
type BranchProtection struct {
	AllowsDeletions             bool
	AllowsForcePushes           bool
	BlocksCreations             bool
	BypassForcePushAllowances   BypassForcePushAllowances   `graphql:"bypassForcePushAllowances(first: 100)"`
	BypassPullRequestAllowances BypassPullRequestAllowances `graphql:"bypassPullRequestAllowances(first: 100)"`
	DismissesStaleReviews       bool
	Id                          string
	IsAdminEnforced             bool
	LockAllowsFetchAndMerge     bool
	LockBranch                  bool
	// TODO: future feature
	// MatchingRefs
	Pattern        string
	PushAllowances PushAllowances `graphql:"pushAllowances(first: 100)"`
	Repository     struct {
		Id         string
		DatabaseId int64
		Name       string
		Owner      struct {
			Login string
			Id    string
		}
	}
	RequireLastPushApproval        bool
	RequiredApprovingReviewCount   int64
	RequiredDeploymentEnvironments []string
	RequiredStatusCheckContexts    []string
	RequiredStatusChecks           []RequiredStatusCheckDescription
	RequiresApprovingReviews       bool
	RequiresCodeOwnerReviews       bool
	RequiresCommitSignatures       bool
	RequiresConversationResolution bool
	RequiresDeployments            bool
	RequiresLinearHistory          bool
	RequiresStatusChecks           bool
	RequiresStrictStatusChecks     bool
	RestrictsPushes                bool
	RestrictsReviewDismissals      bool
	ReviewDismissalAllowances      ReviewDismissalAllowances `graphql:"reviewDismissalAllowances(first: 100)"`
}

// TODO: these are distinct in GitHub, but can probably reduce redundancy
type BranchActorAllowanceActors struct {
	Users []User
	Apps  []App
	Teams []Team
}

type PushAllowanceActors struct {
	Users []User
	Apps  []App
	Teams []Team
}

type ReviewDismissalAllowanceActors struct {
	Users []User
	Apps  []App
	Teams []Team
}

func (bp *BranchProtection) GetBypassForcePushAllowances() *BranchActorAllowanceActors {
	users := []User{}
	apps := []App{}
	teams := []Team{}
	for _, node := range bp.BypassForcePushAllowances.Nodes {
		if node.Actor.App.Id != "" {
			apps = append(apps, node.Actor.App)
		} else if node.Actor.Team.Id != "" {
			teams = append(teams, node.Actor.Team)
		} else if node.Actor.User.Id != "" {
			users = append(users, node.Actor.User)
		}
	}
	return &BranchActorAllowanceActors{
		Users: users,
		Apps:  apps,
		Teams: teams,
	}
}

func (bp *BranchProtection) GetBypassPullRequestAllowances() *BranchActorAllowanceActors {
	users := []User{}
	apps := []App{}
	teams := []Team{}
	for _, node := range bp.BypassPullRequestAllowances.Nodes {
		if node.Actor.App.Id != "" {
			apps = append(apps, node.Actor.App)
		} else if node.Actor.Team.Id != "" {
			teams = append(teams, node.Actor.Team)
		} else if node.Actor.User.Id != "" {
			users = append(users, node.Actor.User)
		}
	}
	return &BranchActorAllowanceActors{
		Users: users,
		Apps:  apps,
		Teams: teams,
	}
}

func (bp *BranchProtection) GetPushAllowances() *PushAllowanceActors {
	users := []User{}
	apps := []App{}
	teams := []Team{}
	for _, node := range bp.PushAllowances.Nodes {
		if node.Actor.App.Id != "" {
			apps = append(apps, node.Actor.App)
		} else if node.Actor.Team.Id != "" {
			teams = append(teams, node.Actor.Team)
		} else if node.Actor.User.Id != "" {
			users = append(users, node.Actor.User)
		}
	}
	return &PushAllowanceActors{
		Users: users,
		Apps:  apps,
		Teams: teams,
	}
}

func (bp *BranchProtection) GetReviewDismissalAllowances() *ReviewDismissalAllowanceActors {
	users := []User{}
	apps := []App{}
	teams := []Team{}
	for _, node := range bp.ReviewDismissalAllowances.Nodes {
		if node.Actor.App.Id != "" {
			apps = append(apps, node.Actor.App)
		} else if node.Actor.Team.Id != "" {
			teams = append(teams, node.Actor.Team)
		} else if node.Actor.User.Id != "" {
			users = append(users, node.Actor.User)
		}
	}
	return &ReviewDismissalAllowanceActors{
		Users: users,
		Apps:  apps,
		Teams: teams,
	}
}

type PageInfo struct {
	EndCursor   string
	HasNextPage bool
}

type App struct {
	Id         string
	DatabaseId int64
	Slug       string
}

type Team struct {
	Id   string
	Slug string
}

type User struct {
	Id    string
	Login string
}

// Used for:
// - BypassPullRequestAllowance
// - BypassForcePushAllowance
type BranchActorAllowanceActor struct {
	App  App  `graphql:"... on App"`
	Team Team `graphql:"... on Team"`
	User User `graphql:"... on User"`
}

// Used for:
// - PushAllowances
type PushAllowanceActor struct {
	App  App  `graphql:"... on App"`
	Team Team `graphql:"... on Team"`
	User User `graphql:"... on User"`
}

// Used for:
// - ReviewDismissalAllowances
type ReviewDismissalAllowanceActor struct {
	App  App  `graphql:"... on App"`
	Team Team `graphql:"... on Team"`
	User User `graphql:"... on User"`
}

type Conflict struct {
	ConflictingPattern string
	ConflictingRefName string
}

type BypassForcePushAllowances struct {
	Nodes []struct {
		Actor BranchActorAllowanceActor
	}
	PageInfo PageInfo
}

type BypassPullRequestAllowances struct {
	Nodes []struct {
		Actor BranchActorAllowanceActor
	}
	PageInfo PageInfo
}

type PushAllowances struct {
	Nodes []struct {
		Actor PushAllowanceActor
	}
	PageInfo PageInfo
}

type RequiredStatusCheckDescription struct {
	App     App
	Context string
}

type ReviewDismissalAllowances struct {
	Nodes []struct {
		Actor ReviewDismissalAllowanceActor
	}
	PageInfo PageInfo
}

func (c *Client) GetBranchProtection(ctx context.Context, nodeId string) (*BranchProtection, error) {
	var q struct {
		Node struct {
			BranchProtectionRule BranchProtection `graphql:"... on BranchProtectionRule"`
		} `graphql:"node(id: $nodeId)"`
	}

	variables := map[string]interface{}{
		"nodeId": nodeId,
	}

	err := c.graphql.Query(ctx, &q, variables)
	if err != nil {
		return nil, err
	}

	branchProtection := q.Node.BranchProtectionRule

	// check paginated functions
	if q.Node.BranchProtectionRule.BypassForcePushAllowances.PageInfo.HasNextPage {
		var q struct {
			Node struct {
				BranchProtectionRule struct {
					BypassForcePushAllowances BypassForcePushAllowances `graphql:"bypassForcePushAllowances(first: 100, after: $cursor)"`
				} `graphql:"... on BranchProtectionRule"`
			} `graphql:"node(id: $nodeId)"`
		}

		variables := map[string]interface{}{
			"nodeId": nodeId,
			"cursor": q.Node.BranchProtectionRule.BypassForcePushAllowances.PageInfo.EndCursor,
		}

		for {
			err := c.graphql.Query(ctx, &q, variables)
			if err != nil {
				return nil, err
			}
			branchProtection.BypassForcePushAllowances.Nodes = append(branchProtection.BypassForcePushAllowances.Nodes, q.Node.BranchProtectionRule.BypassForcePushAllowances.Nodes...)
			if !q.Node.BranchProtectionRule.BypassForcePushAllowances.PageInfo.HasNextPage {
				break
			}
			variables["cursor"] = q.Node.BranchProtectionRule.BypassForcePushAllowances.PageInfo.EndCursor
		}
	}
	if q.Node.BranchProtectionRule.PushAllowances.PageInfo.HasNextPage {
		var q struct {
			Node struct {
				BranchProtectionRule struct {
					PushAllowances PushAllowances `graphql:"pushAllowances(first: 100, after: $cursor)"`
				} `graphql:"... on BranchProtectionRule"`
			} `graphql:"node(id: $nodeId)"`
		}

		variables := map[string]interface{}{
			"nodeId": nodeId,
			"cursor": q.Node.BranchProtectionRule.PushAllowances.PageInfo.EndCursor,
		}

		for {
			err := c.graphql.Query(ctx, &q, variables)
			if err != nil {
				return nil, err
			}
			branchProtection.PushAllowances.Nodes = append(branchProtection.PushAllowances.Nodes, q.Node.BranchProtectionRule.PushAllowances.Nodes...)
			if !q.Node.BranchProtectionRule.PushAllowances.PageInfo.HasNextPage {
				break
			}
			variables["cursor"] = q.Node.BranchProtectionRule.PushAllowances.PageInfo.EndCursor
		}
	}

	if q.Node.BranchProtectionRule.BypassPullRequestAllowances.PageInfo.HasNextPage {
		var q struct {
			Node struct {
				BranchProtectionRule struct {
					BypassPullRequestAllowances BypassPullRequestAllowances `graphql:"bypassPullRequestAllowances(first: 100, after: $cursor)"`
				} `graphql:"... on BranchProtectionRule"`
			} `graphql:"node(id: $nodeId)"`
		}

		variables := map[string]interface{}{
			"nodeId": nodeId,
			"cursor": q.Node.BranchProtectionRule.BypassPullRequestAllowances.PageInfo.EndCursor,
		}

		for {
			err := c.graphql.Query(ctx, &q, variables)
			if err != nil {
				return nil, err
			}
			branchProtection.BypassForcePushAllowances.Nodes = append(branchProtection.BypassForcePushAllowances.Nodes, q.Node.BranchProtectionRule.BypassPullRequestAllowances.Nodes...)
			if !q.Node.BranchProtectionRule.BypassPullRequestAllowances.PageInfo.HasNextPage {
				break
			}
			variables["cursor"] = q.Node.BranchProtectionRule.BypassPullRequestAllowances.PageInfo.EndCursor
		}
	}

	if q.Node.BranchProtectionRule.BypassForcePushAllowances.PageInfo.HasNextPage {
		var q struct {
			Node struct {
				BranchProtectionRule struct {
					BypassForcePushAllowances BypassForcePushAllowances `graphql:"bypassForcePushAllowances(first: 100, after: $cursor)"`
				} `graphql:"... on BranchProtectionRule"`
			} `graphql:"node(id: $nodeId)"`
		}

		variables := map[string]interface{}{
			"nodeId": nodeId,
			"cursor": q.Node.BranchProtectionRule.BypassForcePushAllowances.PageInfo.EndCursor,
		}

		for {
			err := c.graphql.Query(ctx, &q, variables)
			if err != nil {
				return nil, err
			}
			branchProtection.BypassForcePushAllowances.Nodes = append(branchProtection.BypassForcePushAllowances.Nodes, q.Node.BranchProtectionRule.BypassForcePushAllowances.Nodes...)
			if !q.Node.BranchProtectionRule.BypassForcePushAllowances.PageInfo.HasNextPage {
				break
			}
			variables["cursor"] = q.Node.BranchProtectionRule.BypassForcePushAllowances.PageInfo.EndCursor
		}
	}

	if q.Node.BranchProtectionRule.ReviewDismissalAllowances.PageInfo.HasNextPage {
		var q struct {
			Node struct {
				BranchProtectionRule struct {
					ReviewDismissalAllowances ReviewDismissalAllowances `graphql:"reviewDismissalAllowances(first: 100, after: $cursor)"`
				} `graphql:"... on BranchProtectionRule"`
			} `graphql:"node(id: $nodeId)"`
		}

		variables := map[string]interface{}{
			"nodeId": nodeId,
			"cursor": q.Node.BranchProtectionRule.ReviewDismissalAllowances.PageInfo.EndCursor,
		}

		for {
			err := c.graphql.Query(ctx, &q, variables)
			if err != nil {
				return nil, err
			}
			branchProtection.ReviewDismissalAllowances.Nodes = append(branchProtection.ReviewDismissalAllowances.Nodes, q.Node.BranchProtectionRule.ReviewDismissalAllowances.Nodes...)
			if !q.Node.BranchProtectionRule.ReviewDismissalAllowances.PageInfo.HasNextPage {
				break
			}
			variables["cursor"] = q.Node.BranchProtectionRule.ReviewDismissalAllowances.PageInfo.EndCursor
		}
	}

	return &q.Node.BranchProtectionRule, err
}

// this is inefficient and should ideally not be used
func (c *Client) GetBranchProtectionByOwnerRepoPattern(ctx context.Context, repositoryOwner, repositoryName, pattern string) (*BranchProtection, error) {
	var q struct {
		Repository struct {
			BranchProtectionRuleConnection struct {
				Nodes []struct {
					// Branch protection rule
					Id      string
					Pattern string
				}
				PageInfo struct {
					EndCursor   string
					HasNextPage bool
				}
			} `graphql:"branchProtectionRules(first: 100, after: $branchProtectionRuleCursor)"`
		} `graphql:"repository(owner: $repositoryOwner, name: $repositoryName)"`
	}

	variables := map[string]interface{}{
		"repositoryOwner":            githubv4.String(repositoryOwner),
		"repositoryName":             githubv4.String(repositoryName),
		"branchProtectionRuleCursor": (*githubv4.String)(nil),
	}

	for {
		err := c.graphql.Query(ctx, &q, variables)
		if err != nil {
			return nil, err
		}
		for _, node := range q.Repository.BranchProtectionRuleConnection.Nodes {
			if node.Pattern == pattern {
				return c.GetBranchProtection(ctx, node.Id)
			}
		}
		if !q.Repository.BranchProtectionRuleConnection.PageInfo.HasNextPage {
			break
		}
		variables["branchProtectionRuleCursor"] = q.Repository.BranchProtectionRuleConnection.PageInfo.EndCursor
	}
	return nil, fmt.Errorf("no branch protection rule with pattern '%s' found for repo '%s' with owner '%s'", pattern, repositoryName, repositoryOwner)
}

func (c *Client) CreateBranchProtection(ctx context.Context, input *githubv4.CreateBranchProtectionRuleInput) (*BranchProtection, error) {
	var m struct {
		CreateBranchProtectionRule struct {
			BranchProtectionRule BranchProtection
		} `graphql:"createBranchProtectionRule(input: $input)"`
	}
	err := c.graphql.Mutate(ctx, &m, *input, nil)
	if err != nil {
		return nil, err
	}
	return &m.CreateBranchProtectionRule.BranchProtectionRule, nil
}

func (c *Client) UpdateBranchProtection(ctx context.Context, input *githubv4.UpdateBranchProtectionRuleInput) (*BranchProtection, error) {
	var m struct {
		UpdateBranchProtection struct {
			BranchProtectionRule BranchProtection
		} `graphql:"updateBranchProtectionRule(input: $input)"`
	}
	err := c.graphql.Mutate(ctx, &m, *input, nil)
	if err != nil {
		return nil, err
	}
	return &m.UpdateBranchProtection.BranchProtectionRule, nil
}

func (c *Client) DeleteBranchProtection(ctx context.Context, input *githubv4.DeleteBranchProtectionRuleInput) error {
	var m struct {
		DeleteBranchProtectionRule struct {
			ClientMutationId string // can't return nothing
		} `graphql:"deleteBranchProtectionRule(input: $input)"`
	}
	return c.graphql.Mutate(ctx, &m, *input, nil)
}
