package v1

import (
	"context"
	"fmt"
	tanxv1 "github.com/tangxusc/kok/api/v1"
	"github.com/tangxusc/kok/controllers"
	v1 "k8s.io/api/batch/v1"
	v12 "k8s.io/api/core/v1"
	v13 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"math/big"
	"net"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func init() {
	controllers.AddModules(Version, initModule)
}

var initModule = &controllers.Module{
	Order: 10,
	Name:  "init-job",
	Sub:   []*controllers.Module{InitServiceAccount, InitRoleBinding, InitJob},
}

var InitJob = &controllers.Module{
	GetObj: func() controllers.Object {
		return &v1.Job{}
	},
	Render: func(c *tanxv1.Cluster) controllers.Object {
		out := &v1.Job{}
		out.Name = fmt.Sprintf("%s-init", c.Name)
		out.Namespace = c.Namespace
		out.Labels = map[string]string{
			"cluster": c.Name,
			"app":     out.Name,
		}
		out.Spec = v1.JobSpec{
			Template: v12.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   out.Name,
					Labels: out.Labels,
				},
				Spec: v12.PodSpec{
					Containers: []v12.Container{{
						Name:    "init",
						Image:   c.Spec.InitSpec.Image,
						Command: []string{"sh", "/home/init.sh"},
						Env: []v12.EnvVar{{
							Name:  "APISERVER_ADDRESS",
							Value: fmt.Sprintf("%s-apiserver", c.Name),
						}, {
							Name:  "FRONT_APISERVER_ADDRESS",
							Value: fmt.Sprintf("%s-apiserver", c.Name),
						}, {
							Name:  "FRONT_APISERVER_PORT",
							Value: c.Spec.AccessSpec.Port,
						}, {
							Name:  "KUBE_SVC_ADDR",
							Value: NextIpForRange(c.Spec.ServiceClusterIpRange, 1),
						}, {
							Name:  "CA_PKI_NAME",
							Value: getCAPkiName(c),
						}, {
							Name:  "ETCD_SVC_NAME",
							Value: getEtcdSvcName(c),
						}, {
							Name:  "ETCD_SVC_CLIENT_NAME",
							Value: getEtcdSvcClientName(c),
						}, {
							Name:  "ETCD_PKI_PEER_NAME",
							Value: getEtcdPkiPeerName(c),
						}, {
							Name:  "ETCD_PKI_SERVER_NAME",
							Value: getEtcdPkiServerName(c),
						}, {
							Name:  "ETCD_PKI_CLIENT_NAME",
							Value: getEtcdPkiClientName(c),
						}, {
							Name:  "K8S_SERVER_NAME",
							Value: getServerName(c),
						}, {
							Name:  "K8S_CLIENT_NAME",
							Value: getClientName(c),
						}, {
							Name:  "ADMIN_CONFIG_NAME",
							Value: getAdminConfigName(c),
						}, {
							Name:  "NODE_CONFIG_NAME",
							Value: getNodeConfigName(c),
						}, {
							Name:  "CLUSTER_DOMAIN",
							Value: c.Spec.ClusterDomain,
						}},
					}},
					RestartPolicy:      v12.RestartPolicyNever,
					ServiceAccountName: fmt.Sprintf("%s-admin", c.Name),
				},
			},
		}

		return out
	},
	SetStatus: func(c *tanxv1.Cluster, target, now controllers.Object) (bool, controllers.Object) {
		job := now.(*v1.Job)
		c.Status.Init.Status = job.Status
		c.Status.Init.Name = job.Name
		c.Status.Init.DnsAddr = NextIpForRange(c.Spec.ServiceClusterIpRange, 2)
		envs := job.Spec.Template.Spec.Containers[0].Env
		for _, env := range envs {
			if env.Name == "CA_PKI_NAME" {
				c.Status.Init.CaPkiName = env.Value
				continue
			}
			if env.Name == "ETCD_PKI_PEER_NAME" {
				c.Status.Init.EtcdPkiPeerName = env.Value
				continue
			}
			if env.Name == "ETCD_PKI_SERVER_NAME" {
				c.Status.Init.EtcdPkiServerName = env.Value
				continue
			}
			if env.Name == "ETCD_PKI_CLIENT_NAME" {
				c.Status.Init.EtcdPkiClientName = env.Value
				continue
			}
			if env.Name == "K8S_SERVER_NAME" {
				c.Status.Init.ServerName = env.Value
				continue
			}
			if env.Name == "K8S_CLIENT_NAME" {
				c.Status.Init.ClientName = env.Value
				continue
			}
			if env.Name == "ADMIN_CONFIG_NAME" {
				c.Status.Init.AdminConfigName = env.Value
				continue
			}
			if env.Name == "NODE_CONFIG_NAME" {
				c.Status.Init.NodeConfigName = env.Value
				continue
			}
		}

		return false, now
	},
	Del: func(ctx context.Context, c *tanxv1.Cluster, client client.Client) error {
		var err error
		nameFunc := []func(cluster *tanxv1.Cluster) string{getCAPkiName, getEtcdPkiClientName, getEtcdPkiServerName, getEtcdPkiPeerName, getServerName, getClientName, getNodeConfigName, getAdminConfigName}
		for _, namef := range nameFunc {
			err = client.Delete(ctx, &v12.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: namef(c), Namespace: c.Namespace},
			})
			if err != nil && errors.IsNotFound(err) {
				return nil
			}
			if err != nil {
				return err
			}
		}
		return err
	},
	Next: func(c *tanxv1.Cluster) bool {
		for _, condition := range c.Status.Init.Status.Conditions {
			if v1.JobComplete == condition.Type {
				return true
			}
		}
		return false
	},
}

func getEtcdSvcClientName(c *tanxv1.Cluster) string {
	return fmt.Sprintf("%s-etcd-client", c.Name)
}

func getEtcdSvcName(c *tanxv1.Cluster) string {
	return fmt.Sprintf("%s-etcd", c.Name)
}

func getEtcdPkiClientName(c *tanxv1.Cluster) string {
	return fmt.Sprintf("%s-etcd-pki-client", c.Name)
}

func getEtcdPkiServerName(c *tanxv1.Cluster) string {
	return fmt.Sprintf("%s-etcd-pki-server", c.Name)
}

func getEtcdPkiPeerName(c *tanxv1.Cluster) string {
	return fmt.Sprintf("%s-etcd-pki-peer", c.Name)
}

func NextIpForRange(ipRange string, step int64) string {
	ip, _, _ := net.ParseCIDR(ipRange)
	ret := big.NewInt(0)
	ret.SetBytes(ip.To4())
	i := ret.Int64() + step
	return fmt.Sprintf("%d.%d.%d.%d", byte(i>>24), byte(i>>16), byte(i>>8), byte(i))
}

func getCAPkiName(c *tanxv1.Cluster) string {
	return fmt.Sprintf("%s-ca-pki", c.Name)
}

func getServerName(c *tanxv1.Cluster) string {
	return fmt.Sprintf("%s-k8s-server", c.Name)
}

func getClientName(c *tanxv1.Cluster) string {
	return fmt.Sprintf("%s-k8s-client", c.Name)
}

func getNodeConfigName(c *tanxv1.Cluster) string {
	return fmt.Sprintf("%s-node-config", c.Name)
}

func getAdminConfigName(c *tanxv1.Cluster) string {
	return fmt.Sprintf("%s-admin-config", c.Name)
}

var InitServiceAccount = &controllers.Module{
	GetObj: func() controllers.Object {
		return &v12.ServiceAccount{}
	},
	Render: func(c *tanxv1.Cluster) controllers.Object {
		out := &v12.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-admin", c.Name),
				Namespace: c.Namespace,
			},
		}
		return out
	},
	SetStatus: func(c *tanxv1.Cluster, target, now controllers.Object) (bool, controllers.Object) {
		out := now.(*v12.ServiceAccount)
		c.Status.Init.ServiceAccountName = out.Name
		return false, now
	},
}

var InitRoleBinding = &controllers.Module{
	GetObj: func() controllers.Object {
		return &v13.RoleBinding{}
	},
	Render: func(c *tanxv1.Cluster) controllers.Object {
		out := &v13.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-admin", c.Name),
				Namespace: c.Namespace,
			},
			Subjects: []v13.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      fmt.Sprintf("%s-admin", c.Name),
					Namespace: c.Namespace,
				},
			},
			RoleRef: v13.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     "cluster-admin",
			},
		}
		return out
	},
	SetStatus: func(c *tanxv1.Cluster, target, now controllers.Object) (bool, controllers.Object) {
		out := now.(*v13.RoleBinding)
		c.Status.Init.RoleBindingName = out.Name
		return false, now
	},
	SetDefault: func(r *tanxv1.Cluster) {
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
		if r.Spec.KubeletSpec.PodInfraContainerImage == "" {
			r.Spec.KubeletSpec.PodInfraContainerImage = "registry.aliyuncs.com/google_containers/pause:3.1"
		}
		if r.Spec.KubeProxySpec.BindAddress == "" {
			r.Spec.KubeProxySpec.BindAddress = "0.0.0.0"
		}
	},
	ValidateCreateModule: func(r *tanxv1.Cluster) field.ErrorList {
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
		if r.Spec.KubeletSpec.PodInfraContainerImage == "" {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec.kubeletSpec.podInfraContainerImage"), r.Spec.KubeletSpec.PodInfraContainerImage, "不能为空"))
		}
		if r.Spec.KubeProxySpec.BindAddress == "" {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec.kubeProxySpec.bindAddress"), r.Spec.KubeProxySpec.BindAddress, "不能为空"))
		}
		return allErrs
	},
	ValidateUpdateModule: func(now *tanxv1.Cluster, old *tanxv1.Cluster) field.ErrorList {
		var allErrs field.ErrorList
		if now.Spec.ClusterDomain != old.Spec.ClusterDomain {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec.clusterDomain"), now.Spec.ClusterDomain, "不允许修改"))
		}
		if now.Spec.ClusterVersion != old.Spec.ClusterVersion {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec.clusterVersion"), now.Spec.ClusterVersion, "不允许修改"))
		}
		if now.Spec.ClusterCIDR != old.Spec.ClusterCIDR {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec.clusterCIDR"), now.Spec.ClusterCIDR, "不允许修改"))
		}
		if now.Spec.ServiceClusterIpRange != old.Spec.ServiceClusterIpRange {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec.serviceClusterIpRange"), now.Spec.ServiceClusterIpRange, "不允许修改"))
		}
		if now.Spec.InitSpec.Image != old.Spec.InitSpec.Image {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec.initSpec.image"), now.Spec.InitSpec.Image, "不允许修改"))
		}
		if now.Spec.KubeletSpec.PodInfraContainerImage != old.Spec.KubeletSpec.PodInfraContainerImage {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec.kubeletSpec.podInfraContainerImage"), now.Spec.KubeletSpec.PodInfraContainerImage, "不允许修改"))
		}
		if now.Spec.KubeProxySpec.BindAddress != old.Spec.KubeProxySpec.BindAddress {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec.kubeProxySpec.bindAddress"), now.Spec.KubeProxySpec.BindAddress, "不允许修改"))
		}
		return allErrs
	},
}
