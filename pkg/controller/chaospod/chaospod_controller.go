package chaospod

import (
	"context"
	"reflect"

	"strings"

	"github.com/go-logr/logr"
	chaosv1alpha1 "github.com/wtfjoke/ordered-chaos-monkey-operator/pkg/apis/chaos/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_chaospod")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new ChaosPod Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileChaosPod{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("chaospod-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ChaosPod
	err = c.Watch(&source.Kind{Type: &chaosv1alpha1.ChaosPod{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileChaosPod implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileChaosPod{}

// ReconcileChaosPod reconciles a ChaosPod object
type ReconcileChaosPod struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ChaosPod object and makes changes based on the state read
// and what is in the ChaosPod.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileChaosPod) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ChaosPod")

	// Fetch the ChaosPod instance
	instance := &chaosv1alpha1.ChaosPod{}
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

	if len(instance.Spec.PrefixToKill) == 0 {
		reqLogger.Info("No prefix to kill defined in chaospod " + instance.Name + " do nothing")
		// As long prefix is empty, do not requeue
		return reconcile.Result{}, nil
	}

	podListFound := &corev1.PodList{}
	err = r.client.List(context.TODO(), client.InNamespace(request.Namespace), podListFound)
	if err != nil {
		reqLogger.Error(err, "Havent found any pods in "+request.Namespace)
		return reconcile.Result{}, err
	}

	var killedPodNames = killPods(r, instance, podListFound.Items, reqLogger)

	// Update the chaospod status with the killed pod names if needed
	if len(killedPodNames) > 0 {
		// append existing killed pod names to new killed pod names
		for k, v := range instance.Status.KilledPodNames {
			killedPodNames[k] = v
		}
		if !reflect.DeepEqual(killedPodNames, instance.Status.KilledPodNames) {
			instance.Status.KilledPodNames = killedPodNames
			err := r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				reqLogger.Error(err, "Failed to update chaospod status")
				return reconcile.Result{}, err
			}
		}
	}

	// Define a new Pod object
	pod := newPodForCR(instance)

	// Set ChaosPod instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Pod already exists - don't requeue
	reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *chaosv1alpha1.ChaosPod) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}

func killPods(r *ReconcileChaosPod, chaosPod *chaosv1alpha1.ChaosPod, existingPods []corev1.Pod, reqLogger logr.Logger) map[string]string {
	var killedPodNames = make(map[string]string, len(existingPods))
	prefixToKill := chaosPod.Spec.PrefixToKill

	reqLogger.Info("Searching for pods with prefix " + prefixToKill)
	for _, pod := range existingPods {
		if strings.HasPrefix(pod.Name, prefixToKill) {
			podName := pod.Name
			podUID := string(pod.UID)
			alreadyKilledPods := chaosPod.Status.KilledPodNames
			if _, ok := alreadyKilledPods[podUID]; ok {
				reqLogger.Info("😥 Attempted to kill previously killed pod, assume its beeing terminated.", "Pod.Namespace", pod.Namespace, "Pod.Name", podName)
				continue
			}

			reqLogger.Info("🎉 Yay! Found pod to kill!", "Pod.Namespace", pod.Namespace, "Pod.Name", podName)
			err := r.client.Delete(context.TODO(), &pod)

			if err != nil {
				logDeletePodError(reqLogger, err, podName)
			} else {
				killedPodNames[podUID] = podName
				reqLogger.Info("💀 Killed/Deleted pod!", "Pod.Namespace", pod.Namespace, "Pod.Name", podName)
			}
		}
	}
	return killedPodNames
}

func logDeletePodError(reqLogger logr.Logger, err error, podName string) {
	if errors.IsNotFound(err) {
		reqLogger.Info("🤷 Pod '" + podName + "' not found for deletion/killing, assume is already beeing killed")
	} else {
		reqLogger.Error(err, "💥 Problem while killing/deleting pod '"+podName+"'")
	}
}
