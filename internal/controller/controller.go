package controller

type GitHubRequester interface {
	TeamRequester
	RepositoryRequester
	OrganizationRequester
}
