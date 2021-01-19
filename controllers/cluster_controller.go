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

// +kubebuilder:rbac:groups=cluster.kok.tanx,resources=clusters,verbs=get;list;watch;create;update;patch;delete
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
	if !cs.ObjectMeta.DeletionTimestamp.IsZero() {
		return Delete(ctx, cs, rl, r)
	}

	return Reconcile(ctx, cs, rl, r)
}

func Delete(ctx context.Context, c *clusterv1.Cluster, rl logr.Logger, r *ClusterReconciler) (ctrl.Result, error) {
	modules, ok := VersionsModules[c.Spec.ClusterVersion]
	if !ok {
		return ctrl.Result{}, fmt.Errorf("not support version %s", c.Spec.ClusterVersion)
	}
	total := len(modules)
	for index, module := range modules {
		moduleName := module.Name
		moduleString := fmt.Sprintf("[%v/%v]%s", index+1, total, moduleName)
		moduleInst := module.copy()
		rl.Info(moduleString, "moduleInst", moduleInst)
		moduleInst.Init(ctx, c, r, rl)
		rl.Info(moduleString + "module delete")
		r.Recorder.Event(c, v13.EventTypeNormal, "ModuleDeleting", fmt.Sprintf("[%s] Creating", moduleName))
		if err := moduleInst.Delete(); err != nil {
			r.Recorder.Event(c, v13.EventTypeWarning, "ModuleDeleteError", fmt.Sprintf("[%s] Error:%v", moduleName, err))
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

func Reconcile(ctx context.Context, c *clusterv1.Cluster, rl logr.Logger, r *ClusterReconciler) (ctrl.Result, error) {
	modules, ok := VersionsModules[c.Spec.ClusterVersion]
	if !ok {
		return ctrl.Result{}, fmt.Errorf("not support version %s", c.Spec.ClusterVersion)
	}
	rl.Info("Cluster Reconcile", "version", c.Spec.ClusterVersion, "name", c.Name, "namespace", c.Namespace)
	total := len(modules)
	for index, module := range modules {
		//TODO:判断完成后再进行下一步
		time.Sleep(time.Second * 2)
		moduleName := module.Name
		moduleString := fmt.Sprintf("[%v/%v]%s ", index+1, total, moduleName)
		moduleInst := module.copy()
		rl.Info(moduleString, "moduleInst", moduleInst)
		moduleInst.Init(ctx, c, r, rl)
		exist, err := moduleInst.Exist()
		if err != nil {
			return ctrl.Result{}, err
		}
		rl.Info(moduleString+"check module exist", "exist", exist)
		if !exist {
			rl.Info(moduleString + "module not exist,creating...")
			r.Recorder.Event(c, v13.EventTypeNormal, "ModuleCreating", fmt.Sprintf("[%s] Creating", moduleName))
			if err = moduleInst.Create(); err != nil {
				rl.Info(moduleString+"module not exist,create error", "error", err)
				r.Recorder.Event(c, v13.EventTypeWarning, "ModuleCreateError", fmt.Sprintf("[%s] Error:%v", moduleName, err))
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: retryDuration,
				}, err
			}
			return ctrl.Result{}, nil
		} else {
			r.Recorder.Event(c, v13.EventTypeNormal, "ModuleUpdating", fmt.Sprintf("[%s] Updating", moduleName))
			rl.Info(moduleString + "module exist,update status...")
			if err = moduleInst.StatusUpdate(); err != nil {
				r.Recorder.Event(c, v13.EventTypeWarning, "ModuleUpdateError", fmt.Sprintf("[%s] Error:%v", moduleName, err))
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
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterv1.Cluster{}).Owns(&v1.Deployment{}).Owns(&v12.Job{}).Owns(&v13.Service{}).
		Complete(r)
}
