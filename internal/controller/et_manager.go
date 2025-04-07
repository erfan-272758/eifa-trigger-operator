package controller

import (
	"context"
	"fmt"
	"strconv"

	triggerv1 "github.com/erfan-272758/eif-trigger-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	AFTER_CREATE                   = "after-create"
	ON_UNOBSERVED_UPDATE           = "on-unobserved-update"
	ON_OBSERVED_UPDATE             = "on-observed-update"
	BEFORE_DELETE                  = "before-delete"
	AFTER_DELETE                   = "after-delete"
	UNKNOWN                        = "unknown"
	Annotation_Observed_Generation = "eifa-trigger-operator-manager/observed-generation"
)

func (r *EifaTriggerReconciler) Fetch(ctx context.Context, req ctrl.Request) (*triggerv1.EifaTrigger, string, error) {
	eifaTrigger := &triggerv1.EifaTrigger{}
	if err := r.Get(ctx, req.NamespacedName, eifaTrigger); err != nil {
		if apierrors.IsNotFound(err) {
			// deleted before
			return nil, AFTER_DELETE, nil
		}
		// error to get
		return nil, UNKNOWN, err

	}

	// marks as deleted
	if eifaTrigger.GetDeletionTimestamp() != nil {
		return eifaTrigger, BEFORE_DELETE, nil
	}

	// export Observed Generation
	var og int64
	og_raw := eifaTrigger.GetAnnotations()[Annotation_Observed_Generation]
	if og_raw == "" {
		og = 0
	} else {
		rog, err := strconv.ParseInt(og_raw, 10, 64)
		if err != nil {
			return eifaTrigger, UNKNOWN, err
		}
		og = rog
	}

	// just created
	if og == 0 {
		return eifaTrigger, AFTER_CREATE, nil
	}

	// on observed update
	if og == eifaTrigger.Generation {
		return eifaTrigger, ON_OBSERVED_UPDATE, nil
	}

	// on unobserved update
	return eifaTrigger, ON_UNOBSERVED_UPDATE, nil

}

func (r *EifaTriggerReconciler) Modify(ctx context.Context, et *triggerv1.EifaTrigger) error {
	// set observed generation
	if et.Annotations == nil {
		et.Annotations = make(map[string]string, 1)
	}
	et.Annotations[Annotation_Observed_Generation] = string(et.Generation)

	return r.Update(ctx, et)
}
func (r *EifaTriggerReconciler) FetchWUList(ctx context.Context, et *triggerv1.EifaTrigger) ([]client.Object, []client.Object, error) {

	wList, err := r.FetchWList(ctx, et)
	if err != nil {
		return nil, nil, err
	}

	uList, err := r.FetchUList(ctx, et)
	if err != nil {
		return nil, nil, err
	}

	return wList, uList, nil

}
func (r *EifaTriggerReconciler) FetchWList(ctx context.Context, et *triggerv1.EifaTrigger) ([]client.Object, error) {
	watchLS, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: et.Spec.Watch.LabelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("invalid watch label selector: %w", err)
	}

	if et.Spec.Watch.Kind == "Deployment" {
		var watchList appsv1.DeploymentList
		if err := r.List(ctx, &watchList, client.InNamespace(et.Namespace), client.MatchingLabelsSelector{Selector: watchLS}); err != nil {
			return nil, fmt.Errorf("failed to list watch objects: %w", err)
		}
		watchObjs := make([]client.Object, 0, watchList.Size())
		for i := range watchList.Items {
			watchObjs = append(watchObjs, &watchList.Items[i])
		}

		return watchObjs, nil

	}
	if et.Spec.Watch.Kind == "DaemonSet" {
		var watchList appsv1.DaemonSetList
		if err := r.List(ctx, &watchList, client.InNamespace(et.Namespace), client.MatchingLabelsSelector{Selector: watchLS}); err != nil {
			return nil, fmt.Errorf("failed to list watch objects: %w", err)
		}
		watchObjs := make([]client.Object, 0, watchList.Size())
		for i := range watchList.Items {
			watchObjs = append(watchObjs, &watchList.Items[i])
		}

		return watchObjs, nil

	}

	return nil, fmt.Errorf("invalid .Spec.Watch.Kind, %s", et.Spec.Watch.Kind)
}
func (r *EifaTriggerReconciler) FetchUList(ctx context.Context, et *triggerv1.EifaTrigger) ([]client.Object, error) {
	updateLS, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: et.Spec.Update.LabelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("invalid update label selector: %w", err)
	}

	if et.Spec.Update.Kind == "ConfigMap" {
		var updateList corev1.ConfigMapList
		if err := r.List(ctx, &updateList, client.InNamespace(et.Namespace), client.MatchingLabelsSelector{Selector: updateLS}); err != nil {
			return nil, fmt.Errorf("failed to list update objects: %w", err)
		}
		updateObjs := make([]client.Object, 0, updateList.Size())
		for i := range updateList.Items {
			updateObjs = append(updateObjs, &updateList.Items[i])
		}

		return updateObjs, nil

	}
	if et.Spec.Update.Kind == "Secret" {
		var updateList corev1.SecretList
		if err := r.List(ctx, &updateList, client.InNamespace(et.Namespace), client.MatchingLabelsSelector{Selector: updateLS}); err != nil {
			return nil, fmt.Errorf("failed to list update objects: %w", err)
		}
		updateObjs := make([]client.Object, 0, updateList.Size())
		for i := range updateList.Items {
			updateObjs = append(updateObjs, &updateList.Items[i])
		}

		return updateObjs, nil
	}

	return nil, fmt.Errorf("invalid .Spec.Update.Kind, %s", et.Spec.Update.Kind)
}
