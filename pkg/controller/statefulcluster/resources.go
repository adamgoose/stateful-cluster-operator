package statefulcluster

import (
	"context"

	engev1alpha1 "github.com/adamgoose/stateful-cluster-operator/pkg/apis/enge/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *ReconcileStatefulCluster) getStatefulCluster(name types.NamespacedName) (*engev1alpha1.StatefulCluster, error) {
	instance := &engev1alpha1.StatefulCluster{}
	err := r.client.Get(context.TODO(), name, instance)
	return instance, err
}

func (r *ReconcileStatefulCluster) getStatefulClusterPods(name types.NamespacedName, labels map[string]string) (*corev1.PodList, error) {
	pods := &corev1.PodList{}
	listOpts := &client.ListOptions{Namespace: name.Namespace}
	listOpts.MatchingLabels(labels)
	err := r.client.List(context.TODO(), listOpts, pods)
	return pods, err
}
