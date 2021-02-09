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

type PluginModuleContext struct {
	client.Client
	logr.Logger
	*runtime.Scheme
	record.EventRecorder

	context.Context
	clusterv1.ClusterPluginObj
	*clusterv1.Cluster
	Clusters   []*clusterv1.Cluster
	AddVolumes func(*PluginModuleContext, v13.PodSpec) v13.PodSpec
}

type ClusterPluginModule struct {
	Name                string
	create              func(ctx *PluginModuleContext) (*v13.Pod, error)
	next                func(ctx *PluginModuleContext, p *v13.Pod) bool
	updateClusterPlugin func(ctx *PluginModuleContext, p *v13.Pod)
}

var modules = []*ClusterPluginModule{install, unInstall, del}

var del = &ClusterPluginModule{
	Name: "delete",
	create: func(ctx *PluginModuleContext) (*v13.Pod, error) {
		podNames := []string{ctx.ClusterPluginObj.GetStatus().InstallStatus.PodName, ctx.ClusterPluginObj.GetStatus().UninstallStatus.PodName}
		for _, name := range podNames {
			if err := ctx.Client.Delete(ctx.Context, &v13.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: ctx.ClusterPluginObj.GetNamespace(),
				},
			}); err != nil {
				return nil, err
			}
		}
		return nil, nil
	},
	next: func(ctx *PluginModuleContext, p *v13.Pod) bool {
		return false
	},
	updateClusterPlugin: func(ctx *PluginModuleContext, p *v13.Pod) {
		if len(ctx.ClusterPluginObj.GetFinalizers()) != 0 {
			ctx.ClusterPluginObj.SetFinalizers([]string{})
		}
	},
}

var install = &ClusterPluginModule{
	Name: "install",
	create: getCreateFunc(func(ctx *PluginModuleContext) clusterv1.ClusterPluginPodSpec {
		return ctx.ClusterPluginObj.GetSpec().Install
	}, "install"),
	next: func(ctx *PluginModuleContext, p *v13.Pod) bool {
		if !ctx.ClusterPluginObj.GetDeletionTimestamp().IsZero() &&
			ctx.ClusterPluginObj.GetStatus().InstallStatus.Status != v13.PodPending &&
			ctx.ClusterPluginObj.GetStatus().InstallStatus.Status != v13.PodRunning {
			return true
		}
		return false
	},
	updateClusterPlugin: func(ctx *PluginModuleContext, p *v13.Pod) {
		status := ctx.ClusterPluginObj.GetStatus()
		status.InstallStatus.PodName = p.Name
		status.InstallStatus.Status = p.Status.Phase
		ctx.ClusterPluginObj.UpdateStatus(status)
		if len(ctx.ClusterPluginObj.GetFinalizers()) == 0 {
			ctx.ClusterPluginObj.SetFinalizers([]string{ClusterPluginFinalizerName})
		}
	},
}

const MountPath = "/etc/cluster/"

func convertSpec(spec clusterv1.ClusterPluginPodSpec) v13.PodSpec {
	initContainers := make([]v13.Container, len(spec.InitContainers))
	for i, container := range spec.InitContainers {
		initContainers[i] = v13.Container{
			Name:       container.Name,
			Image:      container.Image,
			Command:    container.Command,
			Args:       container.Args,
			WorkingDir: container.WorkingDir,
			//Ports:                  ,
			EnvFrom:        container.EnvFrom,
			Env:            container.Env,
			Resources:      container.Resources,
			VolumeMounts:   container.VolumeMounts,
			VolumeDevices:  container.VolumeDevices,
			LivenessProbe:  container.LivenessProbe,
			ReadinessProbe: container.ReadinessProbe,
			StartupProbe:   container.StartupProbe,
			Lifecycle:      container.Lifecycle,
			//TerminationMessagePath:   ,
			//TerminationMessagePolicy: "",
			ImagePullPolicy: container.ImagePullPolicy,
			SecurityContext: container.SecurityContext,
			//Stdin:                    false,
			//StdinOnce:                false,
			//TTY:                      false,
		}
	}
	containers := make([]v13.Container, len(spec.Containers))
	for i, container := range spec.Containers {
		containers[i] = v13.Container{
			Name:       container.Name,
			Image:      container.Image,
			Command:    container.Command,
			Args:       container.Args,
			WorkingDir: container.WorkingDir,
			//Ports:                  ,
			EnvFrom:        container.EnvFrom,
			Env:            container.Env,
			Resources:      container.Resources,
			VolumeMounts:   container.VolumeMounts,
			VolumeDevices:  container.VolumeDevices,
			LivenessProbe:  container.LivenessProbe,
			ReadinessProbe: container.ReadinessProbe,
			StartupProbe:   container.StartupProbe,
			Lifecycle:      container.Lifecycle,
			//TerminationMessagePath:   ,
			//TerminationMessagePolicy: "",
			ImagePullPolicy: container.ImagePullPolicy,
			SecurityContext: container.SecurityContext,
			//Stdin:                    false,
			//StdinOnce:                false,
			//TTY:                      false,
		}
	}
	return v13.PodSpec{
		Volumes:        spec.Volumes,
		InitContainers: initContainers,
		Containers:     containers,
		//EphemeralContainers:           nil,
		RestartPolicy: v13.RestartPolicyNever,
		//TerminationGracePeriodSeconds: ,
		//ActiveDeadlineSeconds:         nil,
		//DNSPolicy:                     "",
		//NodeSelector:                  nil,
		ServiceAccountName: spec.ServiceAccountName,
		//DeprecatedServiceAccount:      "",
		//AutomountServiceAccountToken:  nil,
		//NodeName:                      "",
		//HostNetwork:                   false,
		//HostPID:                       false,
		//HostIPC:                       false,
		//ShareProcessNamespace:         nil,
		//SecurityContext:               spec.,
		ImagePullSecrets: spec.ImagePullSecrets,
		Hostname:         spec.Hostname,
		//Subdomain:                     "",
		//Affinity:                      nil,
		//SchedulerName:                 "",
		//Tolerations:                   nil,
		//HostAliases:                   ,
		//PriorityClassName:             "",
		//Priority:                      nil,
		//DNSConfig:                     nil,
		//ReadinessGates:                nil,
		RuntimeClassName: spec.RuntimeClassName,
		//EnableServiceLinks:            nil,
		//PreemptionPolicy:              nil,
		//Overhead:                      nil,
		//TopologySpreadConstraints:     nil,
	}
}

func getCreateFunc(f func(ctx *PluginModuleContext) clusterv1.ClusterPluginPodSpec, moduleName string) func(ctx *PluginModuleContext) (*v13.Pod, error) {
	return func(ctx *PluginModuleContext) (*v13.Pod, error) {
		name := getPodName(ctx.ClusterPluginObj, moduleName)
		p := &v13.Pod{}
		namespace := ctx.ClusterPluginObj.GetNamespace()
		err := ctx.Get(ctx, types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		}, p)
		if err != nil && errors.IsNotFound(err) {
			spec := convertSpec(f(ctx))
			spec = ctx.AddVolumes(ctx, spec)
			pod := &v13.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
					Labels: map[string]string{
						"cluster": ctx.ClusterPluginObj.GetClusterNames(),
					},
				},
				Spec: spec,
			}
			if err := controllerutil.SetControllerReference(ctx.ClusterPluginObj, pod, ctx.Scheme); err != nil {
				ctx.Info("set pod owner error", "error", err)
			}
			if err := ctx.Client.Create(ctx.Context, pod); err != nil {
				ctx.Info("create pod error", "error", err)
				return nil, err
			}
			ctx.Event(ctx.ClusterPluginObj, v13.EventTypeNormal, "CreatePod", fmt.Sprintf("Pod %s created", name))
			return pod, nil
		}
		return p, err
	}
}

//TODO:覆盖卸载失败的场景

var unInstall = &ClusterPluginModule{
	Name: "unInstall",
	create: getCreateFunc(func(ctx *PluginModuleContext) clusterv1.ClusterPluginPodSpec {
		return ctx.ClusterPluginObj.GetSpec().Uninstall
	}, "uninstall"),
	next: func(ctx *PluginModuleContext, p *v13.Pod) bool {
		if !ctx.ClusterPluginObj.GetDeletionTimestamp().IsZero() &&
			ctx.ClusterPluginObj.GetStatus().InstallStatus.Status != v13.PodPending &&
			ctx.ClusterPluginObj.GetStatus().InstallStatus.Status != v13.PodRunning {
			return true
		}
		return false
	},
	updateClusterPlugin: func(ctx *PluginModuleContext, p *v13.Pod) {
		status := ctx.ClusterPluginObj.GetStatus()
		status.UninstallStatus.PodName = p.Name
		status.UninstallStatus.Status = p.Status.Phase
		ctx.ClusterPluginObj.UpdateStatus(status)
	},
}

// +kubebuilder:rbac:groups=cluster.kok.tanx,resources=clusterplugins,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.kok.tanx,resources=clusterplugins/status,verbs=get;update;patch

func (r *ClusterPluginReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	namespace := req.Namespace
	log := r.Log.WithValues("cluster_plugin", req.NamespacedName)

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

	pmCtx := &PluginModuleContext{
		Client:           r.Client,
		Logger:           log,
		Scheme:           r.Scheme,
		EventRecorder:    r.Recorder,
		Context:          ctx,
		ClusterPluginObj: cp,
		Cluster:          cluster,
		Clusters:         nil,
		AddVolumes:       clusterPluginAddVolume,
	}
	return reconcile(pmCtx)
}

func clusterPluginAddVolume(ctx *PluginModuleContext, spec v13.PodSpec) v13.PodSpec {
	mount := []v13.VolumeMount{
		{
			Name:      "kubeconfig",
			ReadOnly:  true,
			MountPath: MountPath,
		},
		{
			Name:      "kubeconfig",
			ReadOnly:  true,
			MountPath: "/root/.kube/",
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

	volume := v13.Volume{
		Name: "kubeconfig",
		VolumeSource: v13.VolumeSource{
			Secret: &v13.SecretVolumeSource{
				SecretName: ctx.Cluster.Status.Init.AdminConfigName,
				Items: []v13.KeyToPath{
					{
						Key:  "admin.config",
						Path: "config",
					},
				},
			},
		},
	}
	if len(spec.Volumes) == 0 {
		spec.Volumes = []v13.Volume{volume}
	} else {
		spec.Volumes = append(spec.Volumes, volume)
	}
	return spec
}

func reconcile(pmCtx *PluginModuleContext) (ctrl.Result, error) {
	//install,uninstall,delete
	//create,next,updateCP
	total := len(modules)
	for i, module := range modules {
		modStr := fmt.Sprintf("[%v/%v]%s ", i+1, total, module.Name)
		p, err := module.create(pmCtx)
		if err != nil {
			pmCtx.Info(modStr+"create Cluster Plugin Pod error", "error", err)
			return ctrl.Result{}, err
		}
		module.updateClusterPlugin(pmCtx, p)
		pmCtx.Info(modStr + "updated Cluster Plugin")
		if !module.next(pmCtx, p) {
			break
		}
	}
	if err := pmCtx.Client.Update(pmCtx.Context, pmCtx.ClusterPluginObj); err != nil {
		pmCtx.Info("update Cluster Plugin error", "error", err)
		return ctrl.Result{}, err
	}
	pmCtx.Info("update Cluster Plugin finish", "name", pmCtx.ClusterPluginObj.GetName(),
		"namespace", pmCtx.ClusterPluginObj.GetNamespace(), "cluster", pmCtx.ClusterPluginObj.GetClusterNames())

	return ctrl.Result{}, nil
}

func getPodName(cp clusterv1.ClusterPluginObj, name string) string {
	return fmt.Sprintf("%s-%s", cp.GetName(), name)
}

func (r *ClusterPluginReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&clusterv1.ClusterPlugin{}).Owns(&v13.Pod{}).
		Complete(r)
}
