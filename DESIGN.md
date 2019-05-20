# Operator Design

The Stateful Cluster Operator was built using CoreOS's [Operator SDK](https://github.com/operator-framework/operator-sdk). This document contains some reference information regarding the Operator SDK, as well as implementation-specific notes.

## Project Scaffolding Layout

> Taken from the [Operator SDK Documentation](https://github.com/operator-framework/operator-sdk/blob/master/doc/project_layout.md).

The `operator-sdk` CLI generates a number of packages for each project. The following table describes a basic rundown of each generated file/directory.


| File/Folders   | Purpose                           |
| :---           | :--- |
| cmd       | Contains `manager/main.go` which is the main program of the operator. This instantiates a new manager which registers all custom resource definitions under `pkg/apis/...` and starts all controllers under `pkg/controllers/...`  . |
| pkg/apis | Contains the directory tree that defines the APIs of the Custom Resource Definitions(CRD). Users are expected to edit the `pkg/apis/<group>/<version>/<kind>_types.go` files to define the API for each resource type and import these packages in their controllers to watch for these resource types.|
| pkg/controller | This pkg contains the controller implementations. Users are expected to edit the `pkg/controller/<kind>/<kind>_controller.go` to define the controller's reconcile logic for handling a resource type of the specified `kind`. |
| build | Contains the `Dockerfile` and build scripts used to build the operator. |
| deploy | Contains various YAML manifests for registering CRDs, setting up [RBAC][RBAC], and deploying the operator as a Deployment.
| (Gopkg.toml Gopkg.lock) or (go.mod go.sum) | The [Go mod][go_mod] or [Go Dep][dep] manifests that describe the external dependencies of this operator, depending on the dependency manager chosen when initializing or migrating a project. |
| vendor | The golang [vendor][Vendor] folder that contains the local copies of the external dependencies that satisfy the imports of this project. [Go Dep][dep]/[Go modules][go_mod] manages the vendor directly. |

[RBAC]: https://kubernetes.io/docs/reference/access-authn-authz/rbac/
[Vendor]: https://golang.org/cmd/go/#hdr-Vendor_Directories
[go_mod]: https://github.com/golang/go/wiki/Modules
[dep]: https://github.com/golang/dep

## StatefulCluster Spec

The StatefulCuster spec is identical to `corev1.StatefulSetSpec`, thus you can create a StatefulCluster in the same way you would create a StatefulSet.

The StatefulCluster Status is identical to `corev1.StatefulSetStatus`, however updating the status is not yet implemented by this operator.

## Reconciliation Strategy

Upon receiving a reconciliation request (triggered by changes to StatefulClusters and Pods), the following reconciliation strategy is evaluated:

1. If too many replicas exist, remove one pod and return.
2. For each existing pod:
    - If Pod Pending, return and requeue after 5 seconds
    - If Pod Unhealthy, remove it and return
3. If not enough replicas exist, create one by:
    - Generate a new random replica name
    - Create replica-specific PVCs
    - Patch the pod spec, adding the created PVCs
    - Create replic-specific Pod