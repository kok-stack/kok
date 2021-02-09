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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterv1 "github.com/kok-stack/kok/api/v1"
)

// MultiClusterPluginReconciler reconciles a MultiClusterPlugin object
type MultiClusterPluginReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=cluster.kok.tanx,resources=multiclusterplugins,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.kok.tanx,resources=multiclusterplugins/status,verbs=get;update;patch

func (r *MultiClusterPluginReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	namespace := req.Namespace
	log := r.Log.WithValues("multi_cluster_plugin", req.NamespacedName)

	cp := &clusterv1.MultiClusterPlugin{}
	err := r.Client.Get(ctx, req.NamespacedName, cp)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	clusters := make([]*clusterv1.Cluster, len(cp.Spec.Clusters))
	for i, name := range cp.Spec.Clusters {
		cluster := &clusterv1.Cluster{}
		if err := r.Client.Get(ctx, types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		}, cluster); err != nil {
			return ctrl.Result{}, err
		}
		clusters[i] = cluster
	}

	pmCtx := &PluginModuleContext{
		Client:           r.Client,
		Logger:           log,
		Scheme:           r.Scheme,
		EventRecorder:    r.Recorder,
		Context:          ctx,
		ClusterPluginObj: cp,
		Clusters:         clusters,
	}
	return reconcile(pmCtx, cp)
}

func (r *MultiClusterPluginReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterv1.MultiClusterPlugin{}).
		Complete(r)
}
