package controller

import (
	"context"
	"fmt"
	"time"

	triggerv1 "github.com/erfan-272758/eifa-trigger-operator/api/v1"
	"github.com/erfan-272758/eifa-trigger-operator/internal/store"
	appsv1 "k8s.io/api/apps/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const etFinalizer = "eifa-trigger.eifa.org/finalizer"

func (r *EifaTriggerReconciler) AfterCreate(ctx context.Context, req ctrl.Request, et *triggerv1.EifaTrigger) error {
	// add finalizer
	controllerutil.AddFinalizer(et, etFinalizer)

	// add to store
	wList, uList, err := r.FetchWUList(ctx, et)
	if err != nil {
		return err
	}
	store.Get().Add(wList, uList)

	// update
	return r.Modify(ctx, et)
}
func (r *EifaTriggerReconciler) BeforeDelete(ctx context.Context, req ctrl.Request, et *triggerv1.EifaTrigger) error {
	if controllerutil.ContainsFinalizer(et, etFinalizer) {
		// run finalizer logic
		wList, err := r.FetchWList(ctx, et)
		if err != nil {
			return err
		}
		store.Get().Delete(wList)
		controllerutil.RemoveFinalizer(et, etFinalizer)

		// update
		return r.Modify(ctx, et)
	}
	return nil
}
func (r *EifaTriggerReconciler) OnUpdate(ctx context.Context, req ctrl.Request, et *triggerv1.EifaTrigger) error {
	log := log.FromContext(ctx)
	// update store
	wList, uList, err := r.FetchWUList(ctx, et)
	if err != nil {
		return err
	}
	log.Info("i got, ", fmt.Sprintf("watch: %v, update: %v", wList, uList))

	store.Get().Update(wList, uList)

	return r.Modify(ctx, et)
}

func OnChange(c client.Client, watchObj client.Object) {
	ctx := context.Background()
	log := log.FromContext(ctx)
	log.Info(fmt.Sprintf("onChange call, %v", watchObj))
	updateList := store.Get().GetUpdateList(watchObj)
	if len(updateList) == 0 {
		return
	}

	for _, updateObj := range updateList {

		err := c.Get(ctx, client.ObjectKey{Name: updateObj.GetName(), Namespace: updateObj.GetNamespace()}, updateObj)
		if err != nil {
			log.Error(err, "Try to get update obj")
			return
		}

		if deploy, ok := updateObj.(*appsv1.Deployment); ok {

			if deploy.Spec.Template.Annotations == nil {
				deploy.Spec.Template.Annotations = make(map[string]string, 1)
			}
			deploy.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)

		} else if deamon, ok := updateObj.(*appsv1.DaemonSet); ok {
			if deamon.Spec.Template.Annotations == nil {
				deamon.Spec.Template.Annotations = make(map[string]string, 1)
			}
			deamon.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
		}
		err = c.Update(ctx, updateObj)
		if err != nil {
			log.Error(err, "From Update UpdateObject")
		}
	}

}
