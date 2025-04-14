package utils

import (
	"context"
	"os"

	triggerv1 "github.com/erfan-272758/eifa-trigger-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetId() string {
	if uid, ok := os.LookupEnv("POD_UID"); ok {
		return uid
	}
	hostname, err := os.Hostname()
	if err == nil {
		return hostname
	}
	return "eifa-trigger-controller-manager"

}

func UpdateStatus(ctx context.Context, c client.Client, et *triggerv1.EifaTrigger, cond *metav1.Condition) error {
	if cond == nil {
		// nothing to do
		return nil
	}
	// append
	et.Status.Conditions = append(et.Status.Conditions, *cond)

	// store only last 10 conditions
	if len(et.Status.Conditions) > 10 {
		et.Status.Conditions = et.Status.Conditions[len(et.Status.Conditions)-10:]
	}
	return c.Status().Update(ctx, et)
}
