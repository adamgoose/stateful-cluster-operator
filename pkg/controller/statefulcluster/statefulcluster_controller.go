package statefulcluster

import (
	"context"
	"fmt"
	"time"

	engev1alpha1 "github.com/adamgoose/stateful-cluster-operator/pkg/apis/enge/v1alpha1"

	"github.com/thanhpk/randstr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_statefulcluster")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new StatefulCluster Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileStatefulCluster{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("statefulcluster-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource StatefulCluster
	err = c.Watch(&source.Kind{Type: &engev1alpha1.StatefulCluster{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner StatefulCluster
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &engev1alpha1.StatefulCluster{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileStatefulCluster{}

// ReconcileStatefulCluster reconciles a StatefulCluster object
type ReconcileStatefulCluster struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a StatefulCluster object and makes changes based on the state read
// and what is in the StatefulCluster.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileStatefulCluster) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	// Reconcile Strategy
	// - Fetch CRD
	// - Fetch Pods, foreach
	//   - If unhealthy, delete one
	//   - If creating, return and wait
	// - If too many, delete one
	// - If not enough, create one

	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling StatefulCluster")

	// Fetch the StatefulCluster instance
	instance := &engev1alpha1.StatefulCluster{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Fetch owned Pod instances
	pods := &corev1.PodList{}
	listOpts := &client.ListOptions{Namespace: request.Namespace}
	listOpts.MatchingLabels(instance.Spec.Selector.MatchLabels)
	err = r.client.List(context.TODO(), listOpts, pods)
	if err != nil {
		return reconcile.Result{}, err
	}

	// TODO:
	//   - If unhealthy, delete one
	//   - If creating, return and wait

	desiredReplicas := 1
	if instance.Spec.Replicas != nil {
		desiredReplicas = int(*instance.Spec.Replicas)
	}
	if len(pods.Items) < desiredReplicas {
		// TODO: Create volumes too
		pod, pvcs := newPodForCR(instance)

		for _, pvc := range pvcs {
			// Set StatefulCluster instance as the owner and controller
			if err := controllerutil.SetControllerReference(instance, pvc, r.scheme); err != nil {
				return reconcile.Result{}, err
			}

			reqLogger.Info("Creating a new PVC", "PVC.Namespace", pvc.Namespace, "PVC.Name", pvc.Name)
			err = r.client.Create(context.TODO(), pvc)
			if err != nil {
				return reconcile.Result{}, err
			}
		}

		// Set StatefulCluster instance as the owner and controller
		if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
			return reconcile.Result{}, err
		}

		reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return reconcile.Result{}, err
		}

		time.Sleep(5 * time.Second)
		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, nil

	// // Define a new Pod object
	// pod := newPodForCR(instance)

	// // Set StatefulCluster instance as the owner and controller
	// if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
	// 	return reconcile.Result{}, err
	// }

	// // Check if this Pod already exists
	// found := &corev1.Pod{}
	// err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	// if err != nil && errors.IsNotFound(err) {
	// 	reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
	// 	err = r.client.Create(context.TODO(), pod)
	// 	if err != nil {
	// 		return reconcile.Result{}, err
	// 	}

	// 	// Pod created successfully - don't requeue
	// 	return reconcile.Result{}, nil
	// } else if err != nil {
	// 	return reconcile.Result{}, err
	// }

	// // Pod already exists - don't requeue
	// reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	// return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *engev1alpha1.StatefulCluster) (*corev1.Pod, map[string]*corev1.PersistentVolumeClaim) {
	iteration := randstr.Hex(8)
	labels := cr.Spec.Selector.MatchLabels

	podVols := []corev1.Volume{}
	pvcs := map[string]*corev1.PersistentVolumeClaim{}
	for _, pvct := range cr.Spec.VolumeClaimTemplates {
		pvcs[pvct.Name] = &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%s-%s", cr.Name, pvct.Name, iteration),
				Namespace: cr.Namespace,
				Labels:    labels,
			},
			Spec: pvct.Spec,
		}
		podVols = append(podVols, corev1.Volume{
			Name: pvct.Name,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: fmt.Sprintf("%s-%s-%s", cr.Name, pvct.Name, iteration),
					ReadOnly:  false,
				},
			},
		})
	}

	cr.Spec.Template.Spec.Volumes = podVols
	// for _, container := range cr.Spec.Template.Spec.Containers {
	// 	for _, vm := range container.VolumeMounts {
	// 		vm.Name = pvcs[vm.Name].Name
	// 	}
	// }

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", cr.Name, iteration),
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: cr.Spec.Template.Spec,
	}, pvcs
}
