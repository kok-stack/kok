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
		//删除
		return Delete(ctx, cs, rl, r, r.Scheme)
	}

	return Reconcile(ctx, cs, rl, r, r.Scheme)
}

func Delete(ctx context.Context, c *clusterv1.Cluster, rl logr.Logger, r client.Client, scheme *runtime.Scheme) (ctrl.Result, error) {
	modules, ok := VersionsModules[c.Spec.ClusterVersion]
	if !ok {
		return ctrl.Result{}, fmt.Errorf("not support version %s", c.Spec.ClusterVersion)
	}

	for _, module := range modules {
		if err := module.Init(ctx, c, r, rl, scheme); err != nil {
			return ctrl.Result{}, err
		}
		//TODO:会丢status信息吗?
		if err := module.RemoveFinalizer(); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func Reconcile(ctx context.Context, c *clusterv1.Cluster, rl logr.Logger, r client.Client, scheme *runtime.Scheme) (ctrl.Result, error) {
	modules, ok := VersionsModules[c.Spec.ClusterVersion]
	if !ok {
		return ctrl.Result{}, fmt.Errorf("not support version %s", c.Spec.ClusterVersion)
	}
	fmt.Println("-------------------")
	for _, module := range modules {
		moduleInst := module.copy()
		if err := moduleInst.Init(ctx, c, r, rl, scheme); err != nil {
			return ctrl.Result{}, err
		}
		exist, err := moduleInst.Exist()
		if err != nil {
			return ctrl.Result{}, err
		}
		fmt.Println("exist:", exist)
		if !exist {
			if err = moduleInst.Create(); err != nil {
				fmt.Println("moduleInst.Create():", err)
				return ctrl.Result{}, err
			}
		}
		if err = moduleInst.StatusUpdate(); err != nil {
			fmt.Println("moduleInst.StatusUpdate():", err)
			return ctrl.Result{}, err
		}
		fmt.Println("for循环结束")
	}
	fmt.Println("----3333333333333---------------")
	if err := r.Update(ctx, c); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	//TODO:监听owner为Cluster的
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterv1.Cluster{}).Owns(&clusterv1.Cluster{}).
		Complete(r)
}
