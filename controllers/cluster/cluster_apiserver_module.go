package cluster

import (
	"fmt"
	tanxv1 "github.com/kok-stack/kok/api/v1"
	"github.com/kok-stack/kok/controllers"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"reflect"
)

func NewApiServerModules(cfg *controllers.InitConfig) {
	var apiServerDept = &controllers.Module{
		GetObj: func() controllers.Object {
			return &v12.Deployment{}
		},
		Render: func(c *tanxv1.Cluster) controllers.Object {
			var rep = c.Spec.ApiServerSpec.Count
			name := fmt.Sprintf("%s-apiserver", c.Name)
			var out = &v12.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: c.Namespace,
				},
				Spec: v12.DeploymentSpec{
					Replicas: &rep,
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Name:      name,
							Namespace: c.Namespace,
							Labels: map[string]string{
								"cluster": c.Name,
								"app":     name,
							},
						},
						Spec: v1.PodSpec{
							Containers: []v1.Container{
								{
									Name:  "apiserver",
									Image: c.Spec.ApiServerSpec.Image,
									Command: []string{
										"kube-apiserver",
										"--allow-privileged=true",
										"--authorization-mode=Node,RBAC",
										"--client-ca-file=/pki/ca/ca.pem",
										"--enable-admission-plugins=NamespaceLifecycle,LimitRanger,ServiceAccount,TaintNodesByCondition,Priority,DefaultTolerationSeconds,DefaultStorageClass,StorageObjectInUseProtection,PersistentVolumeClaimResize,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,RuntimeClass,ResourceQuota",
										"--etcd-cafile=/pki/ca/ca.pem",
										"--etcd-certfile=/pki/etcd/etcd-client.crt",
										"--etcd-keyfile=/pki/etcd/etcd-client.key",
										fmt.Sprintf("--etcd-servers=https://%s:%v", c.Status.Etcd.SvcName, c.Status.Etcd.Status.ClientPort),
										"--insecure-port=0",
										"--kubelet-client-certificate=/pki/client/kubernetes-node.pem",
										"--kubelet-client-key=/pki/client/kubernetes-node-key.pem",
										"--kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname",
										"--secure-port=6443",
										fmt.Sprintf("--service-cluster-ip-range=%s", c.Spec.ServiceClusterIpRange),
										"--tls-cert-file=/pki/server/kubernetes-server.pem",
										"--tls-private-key-file=/pki/server/kubernetes-server-key.pem",
									},
									Ports: []v1.ContainerPort{{
										Name:          "https-6443",
										ContainerPort: 6443,
									}},
									LivenessProbe: &v1.Probe{
										InitialDelaySeconds: 10,
										TimeoutSeconds:      15,
										PeriodSeconds:       10,
										SuccessThreshold:    1,
										FailureThreshold:    8,
										Handler: v1.Handler{
											HTTPGet: &v1.HTTPGetAction{
												Path:   "/livez",
												Port:   intstr.FromInt(6443),
												Scheme: "HTTPS",
											},
										},
									},
									ReadinessProbe: &v1.Probe{
										Handler: v1.Handler{
											HTTPGet: &v1.HTTPGetAction{
												Path:   "/readyz",
												Port:   intstr.FromInt(6443),
												Scheme: "HTTPS",
											},
										},
										TimeoutSeconds:   15,
										PeriodSeconds:    1,
										SuccessThreshold: 1,
										FailureThreshold: 3,
									},
									StartupProbe: &v1.Probe{
										Handler: v1.Handler{
											HTTPGet: &v1.HTTPGetAction{
												Path:   "/livez",
												Port:   intstr.FromInt(6443),
												Scheme: "HTTPS",
											},
										},
										InitialDelaySeconds: 10,
										TimeoutSeconds:      15,
										PeriodSeconds:       10,
										SuccessThreshold:    1,
										FailureThreshold:    24,
									},
									VolumeMounts: []v1.VolumeMount{
										{
											Name:      "ca-pki",
											ReadOnly:  true,
											MountPath: "/pki/ca",
										},
										{
											Name:      "etcd-pki",
											ReadOnly:  true,
											MountPath: "/pki/etcd",
										},
										{
											Name:      "k8s-server",
											ReadOnly:  true,
											MountPath: "/pki/server",
										},
										{
											Name:      "k8s-client",
											ReadOnly:  true,
											MountPath: "/pki/client",
										},
									},
								},
							},
							Volumes: []v1.Volume{{
								Name: "ca-pki",
								VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{
									SecretName: c.Status.Init.CaPkiName,
								}},
							}, {
								Name: "etcd-pki",
								VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{
									SecretName: c.Status.Init.EtcdPkiClientName,
								}},
							}, {
								Name: "k8s-server",
								VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{
									SecretName: c.Status.Init.ServerName,
								}},
							}, {
								Name: "k8s-client",
								VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{
									SecretName: c.Status.Init.ClientName,
								}},
							},
							},
						},
					},
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"cluster": c.Name,
							"app":     name,
						},
					},
				},
			}
			return out
		},
		SetStatus: func(c *tanxv1.Cluster, target, now controllers.Object) (bool, controllers.Object) {
			dept := now.(*v12.Deployment)
			c.Status.ApiServer.Status = dept.Status
			c.Status.ApiServer.Name = dept.Name

			t := target.(*v12.Deployment)
			n := now.(*v12.Deployment)
			if !reflect.DeepEqual(t.Spec, n.Spec) {
				n.Spec = t.Spec
				return true, n
			}
			return false, n
		},
		Next: func(c *tanxv1.Cluster) bool {
			for _, condition := range c.Status.ApiServer.Status.Conditions {
				if v12.DeploymentAvailable == condition.Type && v1.ConditionTrue == condition.Status {
					return true
				}
			}
			return false
		},
		SetDefault: func(r *tanxv1.Cluster) {
			if r.Spec.ApiServerSpec.Image == "" {
				r.Spec.ApiServerSpec.Image = cfg.ApiServerImage
			}
			if r.Spec.ApiServerSpec.Count == 0 {
				r.Spec.ApiServerSpec.Count = 3
			}
		},
		ValidateCreateModule: func(r *tanxv1.Cluster) field.ErrorList {
			var allErrs field.ErrorList
			if r.Spec.ApiServerSpec.Image == "" {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec.apiServerSpec.image"), r.Spec.ApiServerSpec.Image, "不能为空"))
			}
			if r.Spec.ApiServerSpec.Count <= 0 {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec.apiServerSpec.count"), r.Spec.ApiServerSpec.Count, "必须>0"))
			}
			return allErrs
		},
		ValidateUpdateModule: func(now *tanxv1.Cluster, old *tanxv1.Cluster) field.ErrorList {
			var allErrs field.ErrorList
			if now.Spec.ApiServerSpec.Image != old.Spec.ApiServerSpec.Image {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec.apiServerSpec.image"), now.Spec.ApiServerSpec.Image, "不允许修改"))
			}
			if now.Spec.ApiServerSpec.Count < 0 {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec.apiServerSpec.count"), now.Spec.ApiServerSpec.Count, "必须>0"))
			}
			return allErrs
		},
	}

	var apiServerSvc = &controllers.Module{
		GetObj: func() controllers.Object {
			return &v1.Service{}
		},
		Render: func(c *tanxv1.Cluster) controllers.Object {
			name := fmt.Sprintf("%s-apiserver", c.Name)
			out := &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: c.Namespace,
				},
				Spec: v1.ServiceSpec{
					Selector: map[string]string{
						"cluster": c.Name,
						"app":     name,
					},
					Type: v1.ServiceTypeNodePort,
					Ports: []v1.ServicePort{
						{
							Name: "https-6443",
							Port: 6443,
						},
					},
				},
			}
			return out
		},
		SetStatus: func(c *tanxv1.Cluster, target, now controllers.Object) (bool, controllers.Object) {
			svc := now.(*v1.Service)
			c.Status.ApiServer.SvcName = svc.Name

			return false, now
		},
	}
	var apiServerModule = &controllers.Module{
		Order: 30,
		Name:  "apiserver-dept",
		Sub:   []*controllers.Module{apiServerDept, apiServerSvc},
	}
	controllers.AddModules(cfg.Version, apiServerModule)
}
