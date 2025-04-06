package controller

import (
	"context"
	"strconv"

	triggerv1 "github.com/erfan-272758/eif-trigger-operator/api/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
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
