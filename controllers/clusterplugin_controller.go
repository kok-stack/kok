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
	v13 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterv1 "github.com/kok-stack/kok/api/v1"
)

const ClusterPluginFinalizerName = "finalizer.clusterplugin.kok.tanx"

// ClusterPluginReconciler reconciles a ClusterPlugin object
type ClusterPluginReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

type pluginModuleContext struct {
	*ClusterPluginReconciler
	context.Context
	logr.Logger
	*clusterv1.ClusterPlugin
}

//TODO:实现
type ClusterPluginModule struct {
	Name                string
	create              func(ctx *pluginModuleContext) (*v13.Pod, error)
	next                func(ctx *pluginModuleContext, p *v13.Pod) bool
	delete              func()
	updateClusterPlugin func(ctx *pluginModuleContext, p *v13.Pod)
}

var modules = []*ClusterPluginModule{install, unInstall, del}

var del = &ClusterPluginModule{
	Name: "delete",
	create: func(ctx *pluginModuleContext) (*v13.Pod, error) {
		//TODO:移除这install,uninstall的pod
		return nil, nil
	},
	next: func(ctx *pluginModuleContext, p *v13.Pod) bool {
		return false
	},
	updateClusterPlugin: func(ctx *pluginModuleContext, p *v13.Pod) {
		if len(ctx.ClusterPlugin.Finalizers) != 0 {
			ctx.ClusterPlugin.Finalizers = []string{}
		}
	},
}

var install = &ClusterPluginModule{
	Name: "install",
	create: getCreateFunc(func(ctx *pluginModuleContext) v13.PodSpec {
		return ctx.Spec.Uninstall
	}),
	next: func(ctx *pluginModuleContext, p *v13.Pod) bool {
		if ctx.ClusterPlugin.ObjectMeta.DeletionTimestamp.IsZero() {
			return true
		}
		return false
	},
	updateClusterPlugin: func(ctx *pluginModuleContext, p *v13.Pod) {
		ctx.ClusterPlugin.Status.InstallStatus.Status = p.Status
		if p.Status.Phase == v13.PodSucceeded {
			ctx.ClusterPlugin.Status.InstallStatus.Ready = true
		}
		if len(ctx.ClusterPlugin.Finalizers) == 0 {
			ctx.ClusterPlugin.Finalizers = []string{ClusterPluginFinalizerName}
		}
	},
}

func getCreateFunc(f func(ctx *pluginModuleContext) v13.PodSpec) func(ctx *pluginModuleContext) (*v13.Pod, error) {
	return func(ctx *pluginModuleContext) (*v13.Pod, error) {
		name := getPodName(ctx.ClusterPlugin, "unInstall")
		p := &v13.Pod{}
		namespace := ctx.ClusterPlugin.Namespace
		err := ctx.Get(ctx, types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		}, p)
		if err != nil && errors.IsNotFound(err) {
			pod := &v13.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
					Labels: map[string]string{
						"cluster": ctx.Spec.ClusterName,
					},
				},
				Spec: f(ctx),
			}
			if err := controllerutil.SetControllerReference(ctx.ClusterPlugin, pod, ctx.ClusterPluginReconciler.Scheme); err != nil {
				ctx.Info("set pod owner error", "error", err)
			}
			if err := ctx.Client.Create(ctx.Context, pod); err != nil {
				ctx.Info("create pod error", "error", err)
				return nil, err
			}
			return pod, nil
		}
		return p, err
	}
}

var unInstall = &ClusterPluginModule{
	Name: "unInstall",
	create: getCreateFunc(func(ctx *pluginModuleContext) v13.PodSpec {
		return ctx.Spec.Uninstall
	}),
	next: func(ctx *pluginModuleContext, p *v13.Pod) bool {
		if ctx.ClusterPlugin.ObjectMeta.DeletionTimestamp.IsZero() {
			return true
		}
		return false
	},
	updateClusterPlugin: func(ctx *pluginModuleContext, p *v13.Pod) {
		ctx.ClusterPlugin.Status.UninstallStatus.Status = p.Status
		if p.Status.Phase == v13.PodSucceeded {
			ctx.ClusterPlugin.Status.UninstallStatus.Ready = true
		}
	},
}

// +kubebuilder:rbac:groups=cluster.kok.tanx,resources=clusterplugins,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.kok.tanx,resources=clusterplugins/status,verbs=get;update;patch

func (r *ClusterPluginReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	namespace := req.Namespace
	log := r.Log.WithValues("clusterplugin", req.NamespacedName)

	cp := &clusterv1.ClusterPlugin{}
	err := r.Client.Get(ctx, req.NamespacedName, cp)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	cluster := &clusterv1.Cluster{}
	if err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      cp.Spec.ClusterName,
	}, cluster); err != nil {
		return ctrl.Result{}, err
	}

	pmCtx := &pluginModuleContext{
		r,
		ctx,
		log,
		cp,
	}
	//install,uninstall,delete
	//create,next,updateCP
	for i, module := range modules {
		fmt.Println(i, module)
		p, err := module.create(pmCtx)
		if err != nil {
			log.Info("create Cluster Plugin pod error", "error", err)
			return ctrl.Result{}, err
		}
		module.updateClusterPlugin(pmCtx, p)
		if !module.next(pmCtx, p) {
			break
		}
	}
	if err := r.Client.Update(ctx, cp); err != nil {
		log.Info("update Cluster Plugin error", "error", err)
		return ctrl.Result{}, err
	}
	log.Info("update Cluster Plugin finish", "name", cp.Name, "namespace", cp.Name, "cluster", cp.Spec.ClusterName)

	return ctrl.Result{}, nil
}

func getPodName(cp *clusterv1.ClusterPlugin, name string) string {
	return fmt.Sprintf("%s-%s", cp.Name, name)
}

func (r *ClusterPluginReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterv1.ClusterPlugin{}).Owns(&v13.Pod{}).
		Complete(r)
}
