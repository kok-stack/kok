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

package v1

import (
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var clusterlog = logf.Log.WithName("cluster-resource")

func (r *Cluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-cluster-kok-tanx-v1-cluster,mutating=true,failurePolicy=fail,groups=cluster.kok.tanx,resources=clusters,verbs=create;update,versions=v1,name=mcluster.kb.io

var _ webhook.Defaulter = &Cluster{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Cluster) Default() {
	clusterlog.Info("default", "name", r.Name)

	if r.Spec.ClusterDomain == "" {
		r.Spec.ClusterDomain = "cluster.local"
	}
	if r.Spec.ClusterVersion == "" {
		r.Spec.ClusterVersion = "1.18.4"
	}
	if r.Spec.ClusterCIDR == "" {
		r.Spec.ClusterCIDR = "10.0.0.0/8"
	}
	if r.Spec.ServiceClusterIpRange == "" {
		r.Spec.ServiceClusterIpRange = "10.96.0.0/12"
	}
	if len(r.Spec.RegistryMirrors) == 0 {
		r.Spec.RegistryMirrors = []string{"https://registry.docker-cn.com"}
	}
	if r.Spec.InitSpec.Image == "" {
		r.Spec.InitSpec.Image = "ccr.ccs.tencentyun.com/k8sonk8s/init:v1"
	}
	if r.Spec.EtcdSpec.Image == "" {
		r.Spec.EtcdSpec.Image = "registry.aliyuncs.com/google_containers/etcd:3.3.10"
	}
	if r.Spec.EtcdSpec.Count == 0 {
		r.Spec.EtcdSpec.Count = 3
	}
	if r.Spec.ApiServerSpec.Image == "" {
		r.Spec.ApiServerSpec.Image = "registry.aliyuncs.com/google_containers/kube-apiserver:v1.18.4"
	}
	if r.Spec.ApiServerSpec.Count == 0 {
		r.Spec.ApiServerSpec.Count = 3
	}
	if r.Spec.ControllerManagerSpec.Image == "" {
		r.Spec.ControllerManagerSpec.Image = "registry.aliyuncs.com/google_containers/kube-controller-manager:v1.18.4"
	}
	if r.Spec.ControllerManagerSpec.Count == 0 {
		r.Spec.ControllerManagerSpec.Count = 1
	}
	if r.Spec.SchedulerSpec.Image == "" {
		r.Spec.SchedulerSpec.Image = "registry.aliyuncs.com/google_containers/kube-scheduler:v1.18.4"
	}
	if r.Spec.SchedulerSpec.Count == 0 {
		r.Spec.SchedulerSpec.Count = 1
	}
	if r.Spec.ClientSpec.Image == "" {
		r.Spec.ClientSpec.Image = "ccr.ccs.tencentyun.com/k8sonk8s/init:v1"
	}
	if r.Spec.KubeletSpec.PodInfraContainerImage == "" {
		r.Spec.KubeletSpec.PodInfraContainerImage = "registry.aliyuncs.com/google_containers/pause:3.1"
	}
	if r.Spec.KubeProxySpec.BindAddress == "" {
		r.Spec.KubeProxySpec.BindAddress = "0.0.0.0"
	}
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-cluster-kok-tanx-v1-cluster,mutating=false,failurePolicy=fail,groups=cluster.kok.tanx,resources=clusters,versions=v1,name=vcluster.kb.io

var _ webhook.Validator = &Cluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Cluster) ValidateCreate() error {
	clusterlog.Info("validate create", "name", r.Name)
	var allErrs field.ErrorList

	if r.Spec.ClusterDomain == "" {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.clusterDomain"), r.Spec.ClusterDomain, "不能为空"))
	}
	if r.Spec.ClusterVersion == "" {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.clusterVersion"), r.Spec.ClusterVersion, "不能为空"))
	}
	if r.Spec.ClusterCIDR == "" {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.clusterCIDR"), r.Spec.ClusterCIDR, "不能为空"))
	}
	if r.Spec.ServiceClusterIpRange == "" {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.serviceClusterIpRange"), r.Spec.ServiceClusterIpRange, "不能为空"))
	}
	if len(r.Spec.RegistryMirrors) == 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.registryMirrors"), r.Spec.RegistryMirrors, "不能为空"))
	}
	if r.Spec.InitSpec.Image == "" {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.initSpec.image"), r.Spec.InitSpec.Image, "不能为空"))
	}
	if r.Spec.EtcdSpec.Image == "" {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.etcdSpec.image"), r.Spec.EtcdSpec.Image, "不能为空"))
	}
	if r.Spec.EtcdSpec.Count%2 == 0 || r.Spec.EtcdSpec.Count < 3 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.etcdSpec.count"), r.Spec.EtcdSpec.Count, "不能为奇数(必须>=3)"))
	}
	if r.Spec.ApiServerSpec.Image == "" {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.apiServerSpec.image"), r.Spec.ApiServerSpec.Image, "不能为空"))
	}
	if r.Spec.ApiServerSpec.Count <= 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.apiServerSpec.count"), r.Spec.ApiServerSpec.Count, "不能<=0"))
	}
	if r.Spec.ControllerManagerSpec.Image == "" {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.controllerManagerSpec.image"), r.Spec.ControllerManagerSpec.Image, "不能为空"))
	}
	if r.Spec.ControllerManagerSpec.Count%2 == 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.controllerManagerSpec.count"), r.Spec.ControllerManagerSpec.Count, "不能为奇数"))
	}
	if r.Spec.SchedulerSpec.Image == "" {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.schedulerSpec.image"), r.Spec.SchedulerSpec.Image, "不能为空"))
	}
	if r.Spec.SchedulerSpec.Count%2 == 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.schedulerSpec.count"), r.Spec.SchedulerSpec.Count, "不能为奇数"))
	}
	if r.Spec.ClientSpec.Image == "" {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.clientSpec.count"), r.Spec.ClientSpec.Image, "不能为空"))
	}
	if r.Spec.KubeletSpec.PodInfraContainerImage == "" {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.kubeletSpec.podInfraContainerImage"), r.Spec.KubeletSpec.PodInfraContainerImage, "不能为空"))
	}
	if r.Spec.KubeProxySpec.BindAddress == "" {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.kubeProxySpec.bindAddress"), r.Spec.KubeProxySpec.BindAddress, "不能为空"))
	}

	if len(allErrs) == 0 {
		return nil
	}
	return errors.NewInvalid(r.TypeMeta.GroupVersionKind().GroupKind(), r.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Cluster) ValidateUpdate(old runtime.Object) error {
	clusterlog.Info("validate update", "name", r.Name)
	oldC := old.(*Cluster)
	var allErrs field.ErrorList

	if r.Spec.ClusterDomain != oldC.Spec.ClusterDomain {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.clusterDomain"), oldC.Spec.ClusterDomain, "不允许修改"))
	}
	if r.Spec.ClusterVersion != oldC.Spec.ClusterVersion {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.clusterVersion"), oldC.Spec.ClusterVersion, "不允许修改"))
	}
	if r.Spec.ClusterCIDR != oldC.Spec.ClusterCIDR {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.clusterCIDR"), oldC.Spec.ClusterCIDR, "不允许修改"))
	}
	if r.Spec.ServiceClusterIpRange != oldC.Spec.ServiceClusterIpRange {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.serviceClusterIpRange"), oldC.Spec.ServiceClusterIpRange, "不允许修改"))
	}
	if r.Spec.InitSpec.Image != oldC.Spec.InitSpec.Image {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.initSpec.image"), oldC.Spec.InitSpec.Image, "不允许修改"))
	}
	if r.Spec.EtcdSpec.Image != oldC.Spec.EtcdSpec.Image {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.etcdSpec.image"), oldC.Spec.EtcdSpec.Image, "不允许修改"))
	}
	if oldC.Spec.EtcdSpec.Count%2 == 0 || oldC.Spec.EtcdSpec.Count < 3 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.etcdSpec.count"), oldC.Spec.EtcdSpec.Count, "不能为奇数(必须>=3)"))
	}
	if r.Spec.ApiServerSpec.Image != oldC.Spec.ApiServerSpec.Image {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.apiServerSpec.image"), oldC.Spec.ApiServerSpec.Image, "不允许修改"))
	}
	if oldC.Spec.ApiServerSpec.Count < 3 || oldC.Spec.ApiServerSpec.Count%2 == 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.apiServerSpec.count"), oldC.Spec.ApiServerSpec.Count, "不能为奇数(必须>=3)"))
	}
	if oldC.Spec.ControllerManagerSpec.Image != oldC.Spec.ControllerManagerSpec.Image {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.controllerManagerSpec.image"), oldC.Spec.ControllerManagerSpec.Image, "不允许修改"))
	}
	if oldC.Spec.ControllerManagerSpec.Count%2 == 0 || oldC.Spec.ControllerManagerSpec.Count < 1 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.controllerManagerSpec.count"), oldC.Spec.ControllerManagerSpec.Count, "不能为奇数(必须>=1)"))
	}
	if r.Spec.SchedulerSpec.Image != oldC.Spec.SchedulerSpec.Image {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.schedulerSpec.image"), oldC.Spec.SchedulerSpec.Image, "不允许修改"))
	}
	if oldC.Spec.SchedulerSpec.Count%2 == 0 || oldC.Spec.SchedulerSpec.Count < 1 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.schedulerSpec.count"), oldC.Spec.SchedulerSpec.Count, "不能为奇数(必须>=1)"))
	}
	if oldC.Spec.ClientSpec.Image != oldC.Spec.ClientSpec.Image {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.clientSpec.count"), oldC.Spec.ClientSpec.Image, "不允许修改"))
	}
	if r.Spec.KubeletSpec.PodInfraContainerImage != oldC.Spec.KubeletSpec.PodInfraContainerImage {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.kubeletSpec.podInfraContainerImage"), oldC.Spec.KubeletSpec.PodInfraContainerImage, "不允许修改"))
	}
	if r.Spec.KubeProxySpec.BindAddress != oldC.Spec.KubeProxySpec.BindAddress {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.kubeProxySpec.bindAddress"), oldC.Spec.KubeProxySpec.BindAddress, "不允许修改"))
	}
	if len(allErrs) == 0 {
		return nil
	}
	return errors.NewInvalid(r.TypeMeta.GroupVersionKind().GroupKind(), r.Name, allErrs)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Cluster) ValidateDelete() error {
	clusterlog.Info("validate delete", "name", r.Name)

	return nil
}
