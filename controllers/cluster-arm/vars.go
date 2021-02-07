package cluster1

import (
	"github.com/kok-stack/kok/controllers"
	"github.com/kok-stack/kok/controllers/cluster"
)

func init() {
	config := &controllers.InitConfig{
		Version:                "arm-1.18.1",
		EtcdRepository:         "etcd-arm64",
		EtcdVersion:            "3.4.3-0",
		ApiServerImage:         "mirrorgcrio/kube-apiserver-arm64:v1.18.1",
		ControllerManagerImage: "mirrorgcrio/kube-controller-manager-arm64:v1.18.1",
		SchedulerImage:         "mirrorgcrio/kube-scheduler-arm64:v1.18.1",
		ClientImage:            "ccr.ccs.tencentyun.com/k8sonk8s/init:v1-arm64",
		InitImage:              "ccr.ccs.tencentyun.com/k8sonk8s/init:v1-arm64",
		PodInfraContainerImage: "mirrorgcrio/pause-arm64:3.2",
	}
	cluster.NewApiServerModules(config)
	cluster.NewClientModules(config)
	cluster.NewControllerManagerModules(config)
	cluster.NewEtcdModules(config)
	cluster.NewInitModules(config)
	cluster.NewSchedulerModules(config)
}
