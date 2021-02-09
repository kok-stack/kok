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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"path"

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
		AddVolumes:       multiClusterPluginAddVolumes,
	}
	return reconcile(pmCtx)
}

func multiClusterPluginAddVolumes(ctx *PluginModuleContext, spec v13.PodSpec) v13.PodSpec {
	plugin := ctx.ClusterPluginObj.(*clusterv1.MultiClusterPlugin)
	cs := ctx.Clusters
	clusters := plugin.Spec.Clusters

	mount := []v13.VolumeMount{
		{
			Name:      "kubeconfig",
			ReadOnly:  true,
			MountPath: MountPath,
		},
	}
	for _, container := range spec.InitContainers {
		if len(container.VolumeMounts) > 0 {
			container.VolumeMounts = append(container.VolumeMounts, mount...)
		} else {
			container.VolumeMounts = mount
		}
	}
	for _, container := range spec.Containers {
		if len(container.VolumeMounts) > 0 {
			container.VolumeMounts = append(container.VolumeMounts, mount...)
		} else {
			container.VolumeMounts = mount
		}
	}

	volumes := make([]v13.Volume, len(clusters))
	for i, cluster := range clusters {
		fmt.Println(cluster)
		volume := v13.Volume{
			Name: "kubeconfig",
			VolumeSource: v13.VolumeSource{
				Secret: &v13.SecretVolumeSource{
					SecretName: cs[i].Status.Init.AdminConfigName,
					Items: []v13.KeyToPath{
						{
							Key:  "admin.config",
							Path: path.Join(cluster, "config"),
						},
					},
				},
			},
		}
		volumes[i] = volume
	}

	if len(spec.Volumes) == 0 {
		spec.Volumes = volumes
	} else {
		spec.Volumes = append(spec.Volumes, volumes...)
	}
	return spec
}

func (r *MultiClusterPluginReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterv1.MultiClusterPlugin{}).
		Complete(r)
}
