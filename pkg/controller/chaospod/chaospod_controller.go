package chaospod

import (
	"context"

	"strings"

	"github.com/go-logr/logr"
	chaosv1alpha1 "github.com/wtfjoke/ordered-chaos-monkey-operator/pkg/apis/chaos/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_chaospod")

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

// PodDeletionResult object identifiyng a just deleted pod
type PodDeletionResult struct {
	podUID            string
	podName           string
	deletionSucessful bool
	reconcile         bool
	err               error
}

// Reconcile reads that state of the cluster for a ChaosPod object and makes changes based on the state read
// and what is in the ChaosPod.Spec
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
	err = r.client.List(context.TODO(), podListFound, client.InNamespace(request.Namespace))
	if err != nil {
		reqLogger.Error(err, "Havent found any pods in "+request.Namespace)
		return reconcile.Result{}, err
	}

	var killedPodResult = killPod(r, instance, podListFound.Items, reqLogger)

	// Update the chaospod status with the killed pod names if needed
	if killedPodResult.deletionSucessful {
		if instance.Status.KilledPodNames == nil {
			instance.Status.KilledPodNames = make(map[string]string, len(podListFound.Items))
		}
		if _, ok := instance.Status.KilledPodNames[killedPodResult.podUID]; !ok {
			instance.Status.KilledPodNames[killedPodResult.podUID] = killedPodResult.podName
			err := r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				reqLogger.Error(err, "Failed to update chaospod status")
				return reconcile.Result{}, err
			}

			// requeue to check for other pods to terminate
			return reconcile.Result{Requeue: true}, nil
		}
	} else if killedPodResult.reconcile {
		// deletion failed, requeue and include error
		return reconcile.Result{Requeue: true}, killedPodResult.err
	}

	reqLogger.Info("Skip reconcile")
	return reconcile.Result{}, nil
}

func killPod(r *ReconcileChaosPod, chaosPod *chaosv1alpha1.ChaosPod, existingPods []corev1.Pod, reqLogger logr.Logger) PodDeletionResult {
	killedPodResult := PodDeletionResult{deletionSucessful: false, reconcile: false}
	prefixToKill := chaosPod.Spec.PrefixToKill

	reqLogger.Info("Searching for pods with prefix " + prefixToKill)
	for _, pod := range existingPods {
		isAlreadyBeeingTerminated := pod.DeletionTimestamp != nil
		if strings.HasPrefix(pod.Name, prefixToKill) && !isAlreadyBeeingTerminated {
			podName := pod.Name

			reqLogger.Info("🎉 Yay! Found pod to kill!", "Pod.Namespace", pod.Namespace, "Pod.Name", podName)
			err := r.client.Delete(context.TODO(), &pod)

			if err != nil {
				logDeletePodError(reqLogger, err, podName)
				killedPodResult.reconcile = true
				killedPodResult.err = err
			} else {
				killedPodResult.podName = podName
				killedPodResult.podUID = string(pod.UID)
				killedPodResult.deletionSucessful = true
				reqLogger.Info("💀 Killed pod!", "Pod.Namespace", pod.Namespace, "Pod.Name", podName)
			}
			break // just delete one pod at a time
		}
	}
	return killedPodResult
}

func logDeletePodError(reqLogger logr.Logger, err error, podName string) {
	if errors.IsNotFound(err) {
		reqLogger.Info("🤷 Pod '" + podName + "' not found for deletion/killing, assume is already beeing killed")
	} else {
		reqLogger.Error(err, "💥 Problem while killing/deleting pod '"+podName+"'")
	}
}
