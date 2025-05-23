package controller

import (
	"context"

	triggerv1 "github.com/erfan-272758/eifa-trigger-operator/api/v1"
	corev1 "k8s.io/api/core/v1"

	goruntime "runtime"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// EifaTriggerReconciler reconciles a EifaTrigger object
type EifaTriggerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;update
// +kubebuilder:rbac:groups=apps,resources=daemonsets,verbs=get;list;watch;update
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups=trigger.eifa.org,resources=eifatriggers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=trigger.eifa.org,resources=eifatriggers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=trigger.eifa.org,resources=eifatriggers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *EifaTriggerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	// Fetch EifaTrigger object
	eifaTrigger, event, err := r.Fetch(ctx, req)
	if err != nil {
		log.Error(err, "After Fetch")
		return reconcile.Result{}, err
	}

	switch event {
	case AFTER_CREATE:
		err = r.AfterCreate(ctx, req, eifaTrigger)
	case ON_OBSERVED_UPDATE:
		// my update
		return reconcile.Result{}, nil
	case ON_UNOBSERVED_UPDATE:
		err = r.OnUpdate(ctx, req, eifaTrigger)
	case BEFORE_DELETE:
		err = r.BeforeDelete(ctx, req, eifaTrigger)
	case AFTER_DELETE:
		// after remove kind
		return reconcile.Result{}, nil
	}

	if err != nil {
		log.Error(err, "Error from Handler")
		r.UpdateStatus(ctx, eifaTrigger, &metav1.Condition{
			Type:               triggerv1.FAILED,
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             "ReconcileHandlerError",
			Message:            err.Error(),
		})
		return reconcile.Result{}, err
	}

	// not reconcile anymore
	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EifaTriggerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	watchHandler := &WatchHandler{
		Client: r.Client,
		Scheme: r.Scheme,
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&triggerv1.EifaTrigger{}).
		Watches(&corev1.ConfigMap{}, watchHandler, builder.WithPredicates(predicate.NewPredicateFuncs(WatchPredicateFunc))).
		Watches(&corev1.Secret{}, watchHandler, builder.WithPredicates(predicate.NewPredicateFuncs(WatchPredicateFunc))).
		WithOptions(controller.Options{MaxConcurrentReconciles: goruntime.NumCPU()}).
		Complete(r)
}
