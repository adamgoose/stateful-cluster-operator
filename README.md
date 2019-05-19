# StatefulCluster Operator

The Stateful Cluster Operator is a custom Kubernetes controller that implements the v1.StatefulSet interface, but implements slightly different logic when it comes to replicating your pods.

## Installation

> Note: Never blindly deploy hosted manifests to your cluster. Always review all manifests before deploying!

```bash
# Create stateful-cluster-operator namespace
kubectl create namespace stateful-cluster-operator

# Create ServiceAccount, Role, and RoleBinding
kubectl apply --namespace stateful-cluster-operator -f https://github.com/adamgoose/stateful-cluster-operator/blob/master/deploy/service_account.yaml
kubectl apply --namespace stateful-cluster-operator -f https://github.com/adamgoose/stateful-cluster-operator/blob/master/deploy/role.yaml
kubectl apply --namespace stateful-cluster-operator -f https://github.com/adamgoose/stateful-cluster-operator/blob/master/deploy/role_binding.yaml

# Deploy the Operator
kubectl apply --namespace stateful-cluster-operator -f https://github.com/adamgoose/stateful-cluster-operator/blob/master/deploy/operator.yaml
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

## Usage

StatefulClusters are Kubernetes Custom Resource Definitions that implement the `v1.StatefulSetSpec` and `v1.StatefulSetCluster` types.

```yaml
apiVersion: enge.me/v1alpha1
kind: StatefulCluster
metadata:
  annotations:
  name: coder
spec:
  replicas: 1
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