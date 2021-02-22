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

func NewControllerManagerModules(cfg *controllers.InitConfig) {
	var controllerMgrDept = &controllers.Module{
		GetObj: func() controllers.Object {
			return &v12.Deployment{}
		},
		Render: func(c *tanxv1.Cluster) controllers.Object {
			var rep = c.Spec.ControllerManagerSpec.Count
			name := fmt.Sprintf("%s-controller-manager", c.Name)
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
									Name:  "controller-manager",
									Image: c.Spec.ControllerManagerSpec.Image,
									Command: []string{
										"kube-controller-manager",
										"--allocate-node-cidrs=true",
										"--authentication-kubeconfig=/pki/config/admin.config",
										"--authorization-kubeconfig=/pki/config/admin.config",
										"--bind-address=127.0.0.1",
										"--client-ca-file=/pki/ca/ca.pem",
										fmt.Sprintf("--cluster-cidr=%s", c.Spec.ClusterCIDR),
										"--cluster-signing-cert-file=/pki/ca/ca.pem",
										"--cluster-signing-key-file=/pki/ca/ca-key.pem",
										"--controllers=*,bootstrapsigner,tokencleaner",
										"--kubeconfig=/pki/config/admin.config",
										"--leader-elect=true",
										//TODO:指定node-cidr-mask-size
										"--node-cidr-mask-size=24",
										"--requestheader-client-ca-file=/pki/ca/ca.pem",
										"--root-ca-file=/pki/ca/ca.pem",
										"--service-account-private-key-file=/pki/server/kubernetes-server-key.pem",
										fmt.Sprintf("--service-cluster-ip-range=%s", c.Spec.ServiceClusterIpRange),
										"--use-service-account-credentials=true",
										//在virtual kubelet下,在loadbalance的service中排除virtual node
										"--feature-gates=ServiceNodeExclusion=true",
									},
									LivenessProbe: &v1.Probe{
										InitialDelaySeconds: 10,
										TimeoutSeconds:      15,
										PeriodSeconds:       10,
										SuccessThreshold:    1,
										FailureThreshold:    8,
										Handler: v1.Handler{
											HTTPGet: &v1.HTTPGetAction{
												Path:   "/healthz",
												Port:   intstr.FromInt(10257),
												Scheme: "HTTPS",
												Host:   "127.0.0.1",
											},
										},
									},
									StartupProbe: &v1.Probe{
										Handler: v1.Handler{
											HTTPGet: &v1.HTTPGetAction{
												Path:   "/healthz",
												Port:   intstr.FromInt(10257),
												Scheme: "HTTPS",
												Host:   "127.0.0.1",
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
											Name:      "k8s-server",
											ReadOnly:  true,
											MountPath: "/pki/server",
										},
										{
											Name:      "k8s-config",
											ReadOnly:  true,
											MountPath: "/pki/config",
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
								Name: "k8s-server",
								VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{
									SecretName: c.Status.Init.ServerName,
								}},
							}, {
								Name: "k8s-config",
								VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{
									SecretName: c.Status.Init.AdminConfigName,
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
			c.Status.ControllerManager.Status = dept.Status
			c.Status.ControllerManager.Name = dept.Name

			t := target.(*v12.Deployment)
			n := now.(*v12.Deployment)
			if !reflect.DeepEqual(t.Spec, n.Spec) {
				n.Spec = t.Spec
				return true, n
			}
			return false, n
		},
		SetDefault: func(r *tanxv1.Cluster) {
			if r.Spec.ControllerManagerSpec.Image == "" {
				r.Spec.ControllerManagerSpec.Image = cfg.ControllerManagerImage
			}
			if r.Spec.ControllerManagerSpec.Count == 0 {
				r.Spec.ControllerManagerSpec.Count = 1
			}
		},
		ValidateCreateModule: func(r *tanxv1.Cluster) field.ErrorList {
			var allErrs field.ErrorList
			if r.Spec.ControllerManagerSpec.Image == "" {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec.controllerManagerSpec.image"), r.Spec.ControllerManagerSpec.Image, "不能为空"))
			}
			if r.Spec.ControllerManagerSpec.Count < 1 {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec.controllerManagerSpec.count"), r.Spec.ControllerManagerSpec.Count, "不能<1"))
			}
			return allErrs
		},
		ValidateUpdateModule: func(now *tanxv1.Cluster, old *tanxv1.Cluster) field.ErrorList {
			var allErrs field.ErrorList
			if now.Spec.ControllerManagerSpec.Image != old.Spec.ControllerManagerSpec.Image {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec.controllerManagerSpec.image"), now.Spec.ControllerManagerSpec.Image, "不允许修改"))
			}
			if now.Spec.ControllerManagerSpec.Count < 1 {
				allErrs = append(allErrs, field.Invalid(field.NewPath("spec.controllerManagerSpec.count"), now.Spec.ControllerManagerSpec.Count, "不能<1"))
			}
			return allErrs
		},
	}

	var ctrMgtModule = &controllers.Module{
		Order: 40,
		Name:  "controllerManager-dept",
		Sub:   []*controllers.Module{controllerMgrDept},
	}
	controllers.AddModules(cfg.Version, ctrMgtModule)
}
