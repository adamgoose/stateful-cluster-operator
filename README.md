# StatefulCluster Operator

[![Docker Pulls](https://img.shields.io/docker/pulls/adamgoose/stateful-cluster-operator.svg?style=for-the-badge)](https://hub.docker.com/r/adamgoose/stateful-cluster-operator)

The Stateful Cluster Operator is a custom Kubernetes controller that implements the `core/v1.StatefulSet` interface, but implements slightly different logic when it comes to replicating your pods.

## Installation

The operator only operates on its own namespace. Future plans include the ability to run a single operator instance. The following example deploys into the `stateful-cluster-operator` namespace; feel free to change this to fit your needs.

> Note: Never blindly deploy hosted manifests to your cluster. Always review all manifests before deploying!

```bash
export OPERATOR_NAMESPACE=stateful-cluster-operator
# Create stateful-cluster-operator namespace
kubectl create namespace $OPERATOR_NAMESPACE

# Create ServiceAccount, Role, and RoleBinding
kubectl apply --namespace $OPERATOR_NAMESPACE -f https://raw.githubusercontent.com/adamgoose/stateful-cluster-operator/master/deploy/service_account.yaml
kubectl apply --namespace $OPERATOR_NAMESPACE -f https://raw.githubusercontent.com/adamgoose/stateful-cluster-operator/master/deploy/role.yaml
kubectl apply --namespace $OPERATOR_NAMESPACE -f https://raw.githubusercontent.com/adamgoose/stateful-cluster-operator/master/deploy/role_binding.yaml

# Deploy the Operator
kubectl apply --namespace $OPERATOR_NAMESPACE -f https://raw.githubusercontent.com/adamgoose/stateful-cluster-operator/master/deploy/crds/enge_v1alpha1_statefulcluster_crd.yaml
kubectl apply --namespace $OPERATOR_NAMESPACE -f https://raw.githubusercontent.com/adamgoose/stateful-cluster-operator/master/deploy/operator.yaml
```

## Background

Kubernetes offers a StatefulSet. Its design incorporates sequentiality.

From the K8s documentation:

- For a StatefulSet with N replicas, when Pods are being deployed, they are created sequentially, in order from {0..N-1}.
- Before a Pod is terminated, all of its successors must be completely shutdown.

This has nice properties as long as there are no nodes or pods that fail. However, failure can create a hole in our sequence:

- Pod 0: Good
- Pod 1: Bad
- Pod 2: Good

In the StatefulSet design, Pod 1 should eventually recover. However, this is not always possible. For example, if the node is bad, the pod should be re-scheduled to a new node. However, if the pod has a claim to a PersistentVolume on the bad node, it cannot be re-scheduled. If we delete the pod in a bad state, the StatefulSet will create a new pod with the same identity. However, re-using an identity for a pod that no longer has the state of the failed pod is problemtatic. Reference: [​https://kubernetes.io/docs/tasks/run-application/force-delete-stateful-set-pod/](https://kubernetes.io/docs/tasks/run-application/force-delete-stateful-set-pod/)

## Differences from StatefulSets

- StatefulClusters have no sequential properties
- The identity of a pod is never re-used by new pods (except for in-places updates, not yet supported)
- The following `core/v1.StatefulSet.Spec` properties are not supported:
  - `ServiceName`
  - `PodManagementPolicy`
  - `UpdateStrategy`
  - `RevisionHistoryLimit`

## Usage

StatefulClusters are Kubernetes Custom Resource Definitions that implement the `v1.StatefulSetSpec` and `v1.StatefulSetCluster` types.

```yaml
apiVersion: enge.me/v1alpha1
kind: StatefulCluster
metadata:
  name: coder
spec:
  replicas: 3
  selector:
    matchLabels:
      app: coder
      enge.me/statefulcluster: coder
  template:
    spec:
      containers:
      - args:
        - --allow-http
        - --no-auth
        - --host
        - 0.0.0.0
        image: codercom/code-server
        name: coder-stateful
        ports:
        - containerPort: 8443
          name: https
        volumeMounts:
        - mountPath: /home/coder/project
          name: data
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes:
      - ReadWriteOnce
      resources:
        requests:
          storage: 1Gi
```

## Feature Roadmap

- Proper Liveness evaluation
- In-Place Upgrades
- Custom PVC Deletion Policy
- Support for writing to the `enge.me/v1alpha1.StatefulCluster.Status`
- Implementation of the following `core/v1.StatefulSet.Spec` properties:
  - `ServiceName`
  - `PodManagementPolicy`
  - `UpdateStrategy`
  - `RevisionHistoryLimit`