package cluster_x86

import (
	"github.com/kok-stack/kok/controllers"
	"github.com/kok-stack/kok/controllers/cluster"
)

func init() {
	config := &controllers.InitConfig{
		Version:                "x86-1.18.4",
		EtcdRepository:         "quay.io/coreos/etcd",
		EtcdVersion:            "3.2.13",
		ApiServerImage:         "registry.aliyuncs.com/google_containers/kube-apiserver:v1.18.4",
		ControllerManagerImage: "registry.aliyuncs.com/google_containers/kube-controller-manager:v1.18.4",
		SchedulerImage:         "registry.aliyuncs.com/google_containers/kube-scheduler:v1.18.4",
		ClientImage:            "ccr.ccs.tencentyun.com/k8sonk8s/init:v1",
		InitImage:              "ccr.ccs.tencentyun.com/k8sonk8s/init:v1",
		PodInfraContainerImage: "registry.aliyuncs.com/google_containers/pause:3.1",
	}
	cluster.NewApiServerModules(config)
	cluster.NewClientModules(config)
	cluster.NewControllerManagerModules(config)
	cluster.NewEtcdModules(config)
	cluster.NewInitModules(config)
	cluster.NewSchedulerModules(config)
}
