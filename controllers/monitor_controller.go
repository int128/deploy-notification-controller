/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// MonitorReconciler reconciles a Monitor object
type MonitorReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=deploynotification.int128.github.io,resources=monitors,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=deploynotification.int128.github.io,resources=monitors/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=deploynotification.int128.github.io,resources=monitors/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Monitor object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *MonitorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var rs appsv1.ReplicaSet
	if err := r.Get(ctx, req.NamespacedName, &rs); err != nil {
		logger.Error(err, "unable to fetch ReplicaSet")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("changed status to", "status", rs.Status, "conditions", rs.Status.Conditions)

	if rs.Status.ReadyReplicas == rs.Status.Replicas {
		logger.Info("ready", "time_to_ready", time.Now().Sub(rs.CreationTimestamp.Time))
		return ctrl.Result{}, nil
	}
	if rs.Status.FullyLabeledReplicas == rs.Status.Replicas {
		logger.Info("not ready", "time_to_complete", time.Now().Sub(rs.CreationTimestamp.Time))
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.ReplicaSet{}, builder.WithPredicates(&StatusChangePredicate{})).
		Complete(r)
}

type StatusChangePredicate struct{}

func (l StatusChangePredicate) Create(_ event.CreateEvent) bool {
	return false
}

func (l StatusChangePredicate) Delete(_ event.DeleteEvent) bool {
	return false
}

func (l StatusChangePredicate) Update(e event.UpdateEvent) bool {
	oldReplicaSet, ok := e.ObjectOld.(*appsv1.ReplicaSet)
	if !ok {
		return false
	}
	newReplicaSet, ok := e.ObjectNew.(*appsv1.ReplicaSet)
	if !ok {
		return false
	}

	if oldReplicaSet.Status.ReadyReplicas != newReplicaSet.Status.ReadyReplicas {
		return true
	}
	if oldReplicaSet.Status.FullyLabeledReplicas != newReplicaSet.Status.FullyLabeledReplicas {
		return true
	}
	return false
}

func (l StatusChangePredicate) Generic(_ event.GenericEvent) bool {
	return false
}
