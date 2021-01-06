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
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=cluster.kok.tanx,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.kok.tanx,resources=clusters/status,verbs=get;update;patch
func (r *ClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	rl := r.Log.WithValues("cluster", req.NamespacedName)

	cs := &clusterv1.Cluster{}
	err := r.Client.Get(ctx, req.NamespacedName, cs)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	//https://book.kubebuilder.io/reference/using-finalizers.html
	if !cs.ObjectMeta.DeletionTimestamp.IsZero() {
		return Delete(ctx, cs, rl, r, r.Scheme)
	}

	return Reconcile(ctx, cs, rl, r, r.Scheme)
}

func Delete(ctx context.Context, c *clusterv1.Cluster, rl logr.Logger, r client.Client, scheme *runtime.Scheme) (ctrl.Result, error) {
	modules, ok := VersionsModules[c.Spec.ClusterVersion]
	if !ok {
		return ctrl.Result{}, fmt.Errorf("not support version %s", c.Spec.ClusterVersion)
	}
	total := len(modules)
	for index, module := range modules {
		moduleString := fmt.Sprintf("[%v/%v]", index+1, total)
		moduleInst := module.copy()
		rl.Info(moduleString, "moduleInst", moduleInst)
		moduleInst.Init(ctx, c, r, rl, scheme)
		rl.Info(moduleString + "module delete")
		if err := moduleInst.Delete(); err != nil {
			rl.Info(moduleString+"module delete error", "error", err)
			return ctrl.Result{}, err
		}
		rl.Info(moduleString + "module delete success")
	}

	c.SetFinalizers([]string{})
	rl.Info("remove Finalizers...")
	if err := r.Update(ctx, c); err != nil {
		rl.Info("remove Finalizers error", "error", err)
		return ctrl.Result{}, err
	}
	rl.Info("delete cluster finish", "name", c.Name, "namespace", c.Namespace)

	return ctrl.Result{}, nil
}

var retryDuration = time.Second * 1

func Reconcile(ctx context.Context, c *clusterv1.Cluster, rl logr.Logger, r client.Client, scheme *runtime.Scheme) (ctrl.Result, error) {
	modules, ok := VersionsModules[c.Spec.ClusterVersion]
	if !ok {
		return ctrl.Result{}, fmt.Errorf("not support version %s", c.Spec.ClusterVersion)
	}
	rl.Info("Cluster Reconcile", "version", c.Spec.ClusterVersion, "name", c.Name, "namespace", c.Namespace)
	total := len(modules)
	for index, module := range modules {
		moduleInst := module.copy()
		moduleString := fmt.Sprintf("[%v/%v]", index+1, total)
		rl.Info(moduleString, "moduleInst", moduleInst)
		moduleInst.Init(ctx, c, r, rl, scheme)
		exist, err := moduleInst.Exist()
		if err != nil {
			return ctrl.Result{}, err
		}
		rl.Info(moduleString+"check module exist", "exist", exist)
		if !exist {
			rl.Info(moduleString + "module not exist,creating...")
			if err = moduleInst.Create(); err != nil {
				rl.Info(moduleString+"module not exist,create error", "error", err)
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: retryDuration,
				}, err
			}
			return ctrl.Result{
				RequeueAfter: retryDuration,
			}, err
		} else {
			rl.Info(moduleString + "module exist,update status...")
			if err = moduleInst.StatusUpdate(); err != nil {
				rl.Info(moduleString+"module exist,update status error", "error", err)
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: retryDuration,
				}, err
			}
		}
	}
	rl.Info("update Cluster(crd) status")
	if len(c.GetFinalizers()) == 0 {
		c.SetFinalizers([]string{FinalizerName})
	}
	if err := r.Update(ctx, c); err != nil {
		rl.Info("update Cluster(crd) status error", "error", err)
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: retryDuration,
		}, err
	}
	rl.Info("End Cluster Reconcile", "version", c.Spec.ClusterVersion)

	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).Owns(&clusterv1.Cluster{}).
		For(&clusterv1.Cluster{}).
		Complete(r)
}
