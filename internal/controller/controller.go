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

type GitHubRequester interface {
	TeamRequester
	RepositoryRequester
	OrganizationRequester
	BranchProtectionRequester
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

// returns if set(a) is equivalent to set(b)
// inefficient if the slices are frequently compared since this constructs
// a new set every time it is called
func cmpSlices[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	setA := map[T]struct{}{}
	for _, a := range a {
		setA[a] = struct{}{}
	}
	for _, b := range b {
		if _, ok := setA[b]; !ok {
			return false
		}
	}
	return true
}
