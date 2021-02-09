package clusterplugin

import (
	"github.com/kok-stack/kok/controllers"
	v13 "k8s.io/api/core/v1"
)

func init() {
	controllers.AddVolumeFunc = addVolume
}

func addVolume(ctx *controllers.PluginModuleContext, spec v13.PodSpec) v13.PodSpec {
	mount := []v13.VolumeMount{
		{
			Name:      "kubeconfig",
			ReadOnly:  true,
			MountPath: controllers.MountPath,
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
