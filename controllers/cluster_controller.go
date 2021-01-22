/*


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
	"fmt"
	"github.com/coreos/etcd-operator/pkg/apis/etcd/v1beta2"
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/api/batch/v1"
	v13 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterv1 "github.com/tangxusc/kok/api/v1"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=cluster.kok.tanx,resources=clusters,verbs=get;list;watch;create;update;patch;Del
// +kubebuilder:rbac:groups=cluster.kok.tanx,resources=clusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=cluster.kok.tanx,resources=clusters/events,verbs=get;list;watch;create;update;patch
func (r *ClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	rl := r.Log.WithValues("cluster", req.NamespacedName)

	cs := &clusterv1.Cluster{}
	err := r.Client.Get(ctx, req.NamespacedName, cs)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	moduleContext := NewModuleContext(ctx, cs, rl, r)
	if !cs.ObjectMeta.DeletionTimestamp.IsZero() {
		return deleteCluster(moduleContext)
	}

	return ReconcileCluster(moduleContext)
}

func deleteCluster(ctx *ModuleContext) (ctrl.Result, error) {
	version := ctx.Spec.ClusterVersion
	modules, ok := VersionsModules[version]
	if !ok {
		return ctrl.Result{}, fmt.Errorf("not support version %s", version)
	}
	total := len(modules)
	for index, module := range modules {
		moduleName := module.Name
		moduleString := fmt.Sprintf("[%v/%v]%s ", index+1, total, moduleName)
		if err := module.Delete(ctx); err != nil {
			ctx.Recorder.Event(ctx, v13.EventTypeWarning, "ModuleDeleteError", fmt.Sprintf("[%s] Error:%v", moduleName, err))
			ctx.Info(moduleString+"module Del error", "error", err)
			return ctrl.Result{}, err
		}
		ctx.Info(moduleString + "module Del success")
	}

	ctx.SetFinalizers([]string{})
	ctx.Info("remove Finalizers...")
	if err := ctx.Update(ctx, ctx.Cluster); err != nil {
		ctx.Info("remove Finalizers error", "error", err)
		return ctrl.Result{}, err
	}
	ctx.Info("Del cluster finish", "name", ctx.Name, "namespace", ctx.Namespace)

	return ctrl.Result{}, nil
}

var retryDuration = time.Second * 1

func ReconcileCluster(ctx *ModuleContext) (ctrl.Result, error) {
	version := ctx.Spec.ClusterVersion
	modules, ok := VersionsModules[version]
	if !ok {
		return ctrl.Result{}, fmt.Errorf("not support version %s", version)
	}
	ctx.Info("Begin Cluster Reconcile", "version", version, "name", ctx.Name, "namespace", ctx.Namespace)
	total := len(modules)
	for index, module := range modules {
		moduleName := module.Name
		moduleString := fmt.Sprintf("[%v/%v]%s ", index+1, total, moduleName)
		ctx.Info(moduleString, "version", version, "name", ctx.Name, "namespace", ctx.Namespace)
		if err := module.Reconcile(ctx); err != nil {
			ctx.Info(moduleString+"module not exist,create error", "error", err)
			ctx.Recorder.Event(ctx, v13.EventTypeWarning, "ReconcileError", fmt.Sprintf("[%s] Error:%v", moduleName, err))
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: retryDuration,
			}, err
		}
		if !module.Ready(ctx) {
			break
		}
	}
	ctx.Info("update Cluster(crd) status")
	if len(ctx.GetFinalizers()) == 0 {
		ctx.SetFinalizers([]string{FinalizerName})
	}
	if err := ctx.Update(ctx, ctx.Cluster); err != nil {
		ctx.Info("update Cluster(crd) status error", "error", err)
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: retryDuration,
		}, err
	}
	ctx.Info("End Cluster Reconcile", "version", version)

	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterv1.Cluster{}).Owns(&v1.Deployment{}).Owns(&v12.Job{}).Owns(&v13.Service{}).Owns(&v1beta2.EtcdCluster{}).
		Complete(r)
}
