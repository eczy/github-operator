package controller

type GitHubRequester interface {
	TeamRequester
	RepositoryRequester
	OrganizationRequester
}

// ptrNonNilAndNotEqualTo returns true if a is not nil and its underlying value does not equal b.
// This is convenient for determining if an optional field needs to be updated or not compared to a
// non-pointer variable.
func ptrNonNilAndNotEqualTo[T comparable](a *T, b T) bool {
	if a == nil {
		return false
	}
	return *a != b
}
