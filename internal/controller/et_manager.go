package controller

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	triggerv1 "github.com/erfan-272758/eifa-trigger-operator/api/v1"
	"github.com/erfan-272758/eifa-trigger-operator/internal/utils"
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
	Annotation_Observer_UID        = "eifa-trigger-operator-manager/observer-uid"
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
	var (
		og          int64
		is_observer bool
	)
	// if observer annotation match to my id
	is_observer = eifaTrigger.GetAnnotations()[Annotation_Observer_UID] == utils.GetId()

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
	if og == eifaTrigger.Generation && is_observer {
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
	et.Annotations[Annotation_Observed_Generation] = fmt.Sprint(et.Generation)
	et.Annotations[Annotation_Observer_UID] = utils.GetId()

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
func (r *EifaTriggerReconciler) FetchUList(ctx context.Context, et *triggerv1.EifaTrigger) ([]client.Object, error) {

	var (
		updateListSpec []triggerv1.UpdateSelector
		updateObjList  []client.Object
	)

	if len(et.Spec.UpdateList) == 0 {
		updateListSpec = []triggerv1.UpdateSelector{et.Spec.Update}
	} else {
		updateListSpec = et.Spec.UpdateList
	}

	if len(updateListSpec) == 0 {
		return nil, errors.New("one of the .Spec.Update or .Spec.UpdateList should be filled")
	}

	for _, updateSpec := range updateListSpec {

		updateLS, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
			MatchLabels: updateSpec.LabelSelector,
		})
		if err != nil {
			return nil, fmt.Errorf("invalid update label selector: %w", err)
		}

		if updateSpec.Kind == "Deployment" {
			var updateList appsv1.DeploymentList
			if err := r.List(ctx, &updateList, client.InNamespace(et.Namespace), client.MatchingLabelsSelector{Selector: updateLS}); err != nil {
				return nil, fmt.Errorf("failed to list update objects: %w", err)
			}
			for i := range updateList.Items {
				updateObjList = append(updateObjList, &updateList.Items[i])
			}
			continue
		}
		if updateSpec.Kind == "DaemonSet" {
			var updateList appsv1.DaemonSetList
			if err := r.List(ctx, &updateList, client.InNamespace(et.Namespace), client.MatchingLabelsSelector{Selector: updateLS}); err != nil {
				return nil, fmt.Errorf("failed to list update objects: %w", err)
			}
			for i := range updateList.Items {
				updateObjList = append(updateObjList, &updateList.Items[i])
			}
			continue
		}

		return nil, fmt.Errorf("invalid .Spec.Update.Kind or .Spec.UpdateList.Kind, %s", updateSpec.Kind)
	}
	return updateObjList, nil

}
func (r *EifaTriggerReconciler) FetchWList(ctx context.Context, et *triggerv1.EifaTrigger) ([]client.Object, error) {

	var (
		watchListSpec []triggerv1.WatchSelector
		watchObjList  []client.Object
	)

	if len(et.Spec.WatchList) == 0 {
		watchListSpec = []triggerv1.WatchSelector{et.Spec.Watch}
	} else {
		watchListSpec = et.Spec.WatchList
	}

	if len(watchListSpec) == 0 {
		return nil, errors.New("one of the .Spec.Watch or .Spec.WatchList should be filled")
	}
	for _, watchSpec := range watchListSpec {

		watchLS, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
			MatchLabels: watchSpec.LabelSelector,
		})
		if err != nil {
			return nil, fmt.Errorf("invalid watch label selector: %w", err)
		}

		if watchSpec.Kind == "ConfigMap" {
			var watchList corev1.ConfigMapList
			if err := r.List(ctx, &watchList, client.InNamespace(et.Namespace), client.MatchingLabelsSelector{Selector: watchLS}); err != nil {
				return nil, fmt.Errorf("failed to list watch objects: %w", err)
			}
			for i := range watchList.Items {
				watchObjList = append(watchObjList, &watchList.Items[i])
			}

			continue
		}
		if watchSpec.Kind == "Secret" {
			var watchList corev1.SecretList
			if err := r.List(ctx, &watchList, client.InNamespace(et.Namespace), client.MatchingLabelsSelector{Selector: watchLS}); err != nil {
				return nil, fmt.Errorf("failed to list watch objects: %w", err)
			}
			for i := range watchList.Items {
				watchObjList = append(watchObjList, &watchList.Items[i])
			}
			continue
		}

		return nil, fmt.Errorf("invalid .Spec.Watch.Kind or .Spec.WatchList.Kind, %s", watchSpec.Kind)
	}
	return watchObjList, nil
}

func (r *EifaTriggerReconciler) UpdateStatus(ctx context.Context, et *triggerv1.EifaTrigger, cond *metav1.Condition) error {
	return utils.UpdateStatus(ctx, r.Client, et, cond)
}
