package controller

import (
	"context"

	triggerv1 "github.com/erfan-272758/eif-trigger-operator/api/v1"
	"github.com/erfan-272758/eif-trigger-operator/internal/store"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const etFinalizer = "eifa-trigger.eifa.org/finalizer"

func (r *EifaTriggerReconciler) AfterCreate(ctx context.Context, req ctrl.Request, et *triggerv1.EifaTrigger) error {
	// add finalizer
	controllerutil.AddFinalizer(et, etFinalizer)

	// add to store
	store.Get().Add(et)

	// update
	return r.Modify(ctx, et)
}
func (r *EifaTriggerReconciler) BeforeDelete(ctx context.Context, req ctrl.Request, et *triggerv1.EifaTrigger) error {
	if controllerutil.ContainsFinalizer(et, etFinalizer) {
		// run finalizer logic
		store.Get().Delete(et)
		controllerutil.RemoveFinalizer(et, etFinalizer)

		// update
		return r.Modify(ctx, et)
	}
	return nil
}
func (r *EifaTriggerReconciler) OnUpdate(ctx context.Context, req ctrl.Request, et *triggerv1.EifaTrigger) error {
	// update store
	store.Get().Update(et)

	return r.Modify(ctx, et)
}
