package cluster1

import (
	"fmt"
	tanxv1 "github.com/tangxusc/kok/api/v1"
	"github.com/tangxusc/kok/controllers"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"reflect"
)

func init() {
	controllers.AddModules(Version, schedulerModule)
}

var schedulerModule = &controllers.Module{
	Order: 50,
	Name:  "scheduler-dept",
	Sub:   []*controllers.Module{schedulerDept},
}

var schedulerDept = &controllers.Module{
	GetObj: func() controllers.Object {
		return &v12.Deployment{}
	},
	Render: func(c *tanxv1.Cluster) controllers.Object {
		var rep = c.Spec.SchedulerSpec.Count
		name := fmt.Sprintf("%s-scheduler", c.Name)
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
								Name:  "scheduler",
								Image: c.Spec.SchedulerSpec.Image,
								Command: []string{
									"kube-scheduler",
									"--kubeconfig=/pki/config/admin.config",
									"--authentication-kubeconfig=/pki/config/admin.config",
									"--authorization-kubeconfig=/pki/config/admin.config",
									"--leader-elect=true",
									"--requestheader-client-ca-file=/pki/ca/ca.pem",
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
		c.Status.Scheduler.Status = dept.Status
		c.Status.Scheduler.Name = dept.Name

		t := target.(*v12.Deployment)
		n := now.(*v12.Deployment)
		if !reflect.DeepEqual(t.Spec, n.Spec) {
			n.Spec = t.Spec
			return true, n
		}
		return false, n
	},
	SetDefault: func(r *tanxv1.Cluster) {
		if r.Spec.SchedulerSpec.Image == "" {
			r.Spec.SchedulerSpec.Image = "registry.aliyuncs.com/google_containers/kube-scheduler:v1.18.4"
		}
		if r.Spec.SchedulerSpec.Count == 0 {
			r.Spec.SchedulerSpec.Count = 1
		}
	},
	ValidateCreateModule: func(r *tanxv1.Cluster) field.ErrorList {
		var allErrs field.ErrorList
		if r.Spec.SchedulerSpec.Image == "" {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec.schedulerSpec.image"), r.Spec.SchedulerSpec.Image, "不能为空"))
		}
		if r.Spec.SchedulerSpec.Count < 1 {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec.schedulerSpec.count"), r.Spec.SchedulerSpec.Count, "必须>1"))
		}
		return allErrs
	},
	ValidateUpdateModule: func(now *tanxv1.Cluster, old *tanxv1.Cluster) field.ErrorList {
		var allErrs field.ErrorList
		if now.Spec.SchedulerSpec.Image != old.Spec.SchedulerSpec.Image {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec.schedulerSpec.image"), now.Spec.SchedulerSpec.Image, "不允许修改"))
		}
		if now.Spec.SchedulerSpec.Count < 1 {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec.schedulerSpec.count"), now.Spec.SchedulerSpec.Count, "必须>1"))
		}
		return allErrs
	},
}
