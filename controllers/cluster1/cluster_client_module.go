package cluster1

import (
	"fmt"
	tanxv1 "github.com/kok-stack/kok/api/v1"
	"github.com/kok-stack/kok/controllers"
	v12 "k8s.io/api/apps/v1"
	v13 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"reflect"
)

func init() {
	controllers.AddModules(Version, clientModule)
}

var clientModule = &controllers.Module{
	Order: 60,
	Name:  "client-dept",
	Sub:   []*controllers.Module{clientDept, installPostJob},
}

var clientDept = &controllers.Module{
	GetObj: func() controllers.Object {
		return &v12.Deployment{}
	},
	Render: func(c *tanxv1.Cluster) controllers.Object {
		var rep int32 = 1
		var termination int64 = 1
		name := fmt.Sprintf("%s-client", c.Name)
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
						TerminationGracePeriodSeconds: &termination,
						Containers: []v1.Container{
							{
								Name:  "scheduler",
								Image: c.Spec.ClientSpec.Image,
								Command: []string{
									"cat",
								},
								Stdin: true,
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
									{
										Name:      "nodeconfig",
										ReadOnly:  true,
										MountPath: "/home/node",
									},
									{
										Name:      "adminconfig",
										ReadOnly:  true,
										MountPath: "/home/admin",
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
						}, {
							Name: "nodeconfig",
							VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{
								SecretName: c.Status.Init.NodeConfigName,
							}},
						}, {
							Name: "adminconfig",
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
		c.Status.Client.Status = dept.Status
		c.Status.Client.Name = dept.Name

		t := target.(*v12.Deployment)
		n := now.(*v12.Deployment)
		if !reflect.DeepEqual(t.Spec, n.Spec) {
			n.Spec = t.Spec
			return true, n
		}
		return false, n
	},
	SetDefault: func(r *tanxv1.Cluster) {
		if r.Spec.ClientSpec.Image == "" {
			r.Spec.ClientSpec.Image = "ccr.ccs.tencentyun.com/k8sonk8s/init:v1"
		}
	},
	ValidateCreateModule: func(r *tanxv1.Cluster) field.ErrorList {
		var allErrs field.ErrorList
		if r.Spec.ClientSpec.Image == "" {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec.clientSpec.count"), r.Spec.ClientSpec.Image, "不能为空"))
		}
		return allErrs
	},
	ValidateUpdateModule: func(now *tanxv1.Cluster, old *tanxv1.Cluster) field.ErrorList {
		var allErrs field.ErrorList
		if now.Spec.ClientSpec.Image != old.Spec.ClientSpec.Image {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec.clientSpec.count"), now.Spec.ClientSpec.Image, "不允许修改"))
		}
		return allErrs
	},
}
var installPostJob = &controllers.Module{
	GetObj: func() controllers.Object {
		return &v13.Job{}
	},
	Render: func(c *tanxv1.Cluster) controllers.Object {
		out := &v13.Job{}
		out.Name = fmt.Sprintf("%s-install-post", c.Name)
		out.Namespace = c.Namespace
		out.Labels = map[string]string{
			"cluster": c.Name,
			"app":     out.Name,
		}
		out.Spec = v13.JobSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:   out.Name,
					Labels: out.Labels,
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{
						Name:  "install-post",
						Image: c.Spec.InitSpec.Image,
						Command: []string{
							"kubectl",
							"--kubeconfig=admin/admin.config",
							"create",
							"clusterrolebinding",
							"cluster-node",
							"--clusterrole=cluster-admin",
							"--user=kubernetes-node",
							"--group=system:node",
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
							{
								Name:      "nodeconfig",
								ReadOnly:  true,
								MountPath: "/home/node",
							},
							{
								Name:      "adminconfig",
								ReadOnly:  true,
								MountPath: "/home/admin",
							},
						},
					}},
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
					}, {
						Name: "nodeconfig",
						VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{
							SecretName: c.Status.Init.NodeConfigName,
						}},
					}, {
						Name: "adminconfig",
						VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{
							SecretName: c.Status.Init.AdminConfigName,
						}},
					},
					},
					RestartPolicy: v1.RestartPolicyNever,
				},
			},
		}
		return out
	},
	SetStatus: func(c *tanxv1.Cluster, target, now controllers.Object) (bool, controllers.Object) {
		job := now.(*v13.Job)
		c.Status.PostInstall.Status = job.Status
		c.Status.PostInstall.Name = job.Name

		return false, now
	},
}
