package controllers

//func TestReconcile(t *testing.T) {
//	c := &v1.Cluster{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      "test",
//			Namespace: "test",
//		},
//		Spec: v1.ClusterSpec{
//			ClusterVersion:        "1.18.4",
//			ClusterCIDR:           "10.0.0.0/8",
//			ClusterDNSAddr:        "10.96.0.2",
//			ServiceClusterIpRange: "10.96.0.0/12",
//			AccessSpec: v1.ClusterAccessSpec{
//				Address: "1.1.1.1",
//				Port:    "8899",
//			},
//			InitSpec: v1.ClusterInitSpec{
//				v1.ImageBase{Image: "ccr.ccs.tencentyun.com/k8s-test/init:v1"},
//			},
//			EtcdSpec: v1.ClusterEtcdSpec{v1.ImageBase{Image: "registry.aliyuncs.com/google_containers/etcd:3.3.10"}},
//		},
//	}
//	fakeClient := fake.NewFakeClient()
//	Reconcile(context.TODO(), c, nil, fakeClient, nil)
//}
