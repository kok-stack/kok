package controllers

import (
	"fmt"
	tanxv1 "github.com/tangxusc/kok/api/v1"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ctrMgtModule = ParentModule{
	Name: "controllerManager-dept",
	Sub:  []Module{controllerMgrDept},
}

var controllerMgrDept = &SubModule{
	getObj: func() Object {
		return &v12.Deployment{}
	},
	render: func(c *tanxv1.Cluster, s *SubModule) Object {
		var rep int32 = c.Spec.ControllerManagerSpec.Count
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
	updateStatus: func(c *tanxv1.Cluster, object Object) {
		dept := object.(*v12.Deployment)
		c.Status.ControllerManager.Status = dept.Status
		c.Status.ControllerManager.Name = dept.Name
	},
}
