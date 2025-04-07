package controller

import (
	"context"

	"github.com/erfan-272758/eif-trigger-operator/internal/store"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type WatchHandler struct {
	Client client.Client
	Scheme *runtime.Scheme
}

func (e *WatchHandler) Create(ctx context.Context, evt event.TypedCreateEvent[client.Object], q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
}
func (e *WatchHandler) Delete(ctx context.Context, evt event.TypedDeleteEvent[client.Object], q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
}
func (e *WatchHandler) Generic(ctx context.Context, evt event.TypedGenericEvent[client.Object], q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
}
func (e *WatchHandler) Update(ctx context.Context, evt event.TypedUpdateEvent[client.Object], q workqueue.TypedRateLimitingInterface[reconcile.Request]) {
	OnChange(e.Client, evt.ObjectNew)
}

func WatchPredicateFunc(obj client.Object) bool {
	return store.Get().IsInWatchList(obj)
}
