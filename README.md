# github-operator
GitHub Operator is a [Kubernetes operator] built to manage GitHub resources in a declarative manner.

## Description
By representing GitHub resources as Kubernetes resources, GitHub Operator lets users leverage existing Kubernetes resource management tools like [Kustomize] to easily manage GitHub resources. Users also enjoy the usual suite of [Kubernetes API features], including protection against state drift via the Kubernetes reconciliation loop.

### Supported GitHub resources:

| Resource               | Create | Update | Delete |
| ---------------------- | ------ | ------ | ------ |
| Repository             | ✅      | ✅      | ✅      |
| Branch Protection Rule | ✅      | ✅      | ✅      |
| Team                   | ✅      | ✅      | ✅      |
| Organization           | ❌      | ✅      | ❌      |

If you would like a new resource to be supported, please open an issue.

## Usage
TODO

## License

Copyright 2024 Evan Czyzycki.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

[Kubernetes API features]: https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#common-features
[Kubernetes operator]: https://kubernetes.io/docs/concepts/extend-kubernetes/operator/
[Kustomize]: https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/