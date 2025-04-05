package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	triggerv1 "github.com/erfan-272758/eif-trigger-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EifaTriggerReconciler reconciles a EifaTrigger object
type EifaTriggerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=trigger.eifa.org,resources=eifatriggers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=trigger.eifa.org,resources=eifatriggers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=trigger.eifa.org,resources=eifatriggers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *EifaTriggerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var trigger triggerv1.EifaTrigger
	if err := r.Get(ctx, req.NamespacedName, &trigger); err != nil {
		logger.Error(err, "unable to fetch EifaTrigger")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	watchSelector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{MatchLabels: trigger.Spec.WatchLabelSelector})
	if err != nil {
		logger.Error(err, "invalid watchLabelSelector")
		return ctrl.Result{}, err
	}

	applySelector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{MatchLabels: trigger.Spec.ApplyLabelSelector})
	if err != nil {
		logger.Error(err, "invalid applyLabelSelector")
		return ctrl.Result{}, err
	}

	var pods corev1.PodList
	if err := r.List(ctx, &pods, client.InNamespace(req.Namespace), client.MatchingLabelsSelector{Selector: applySelector}); err != nil {
		logger.Error(err, "failed to list pods with applyLabelSelector")
		return ctrl.Result{}, err
	}

	for _, pod := range pods.Items {
		err := r.Delete(ctx, &pod)
		if err != nil {
			logger.Error(err, "failed to delete pod", "name", pod.Name)
		} else {
			logger.Info("deleted pod to trigger restart", "name", pod.Name)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EifaTriggerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&triggerv1.EifaTrigger{}).
		Watches(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
			OwnerType:    &triggerv1.EifaTrigger{},
			IsController: false,
		}).
		WithEventFilter(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				// You can enhance this to check label changes for more efficiency
				return true
			},
		}).
		Complete(r)
}
