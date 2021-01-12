#!/bin/bash
DOWNLOAD_ADDRESS=$1
apt update && apt install -y wget
mkdir -p /etc/kubernetes/
mkdir -p /etc/docker/
#下载
wget -O /etc/kubernetes/ca.pem "${DOWNLOAD_ADDRESS}"/meta/{{.ObjectMeta.Namespace}}/{{.ObjectMeta.Name}}/ca/ca.pem
wget -O /etc/kubernetes/node.config "${DOWNLOAD_ADDRESS}"/meta/{{.ObjectMeta.Namespace}}/{{.ObjectMeta.Name}}/nodeconfig/node.config
wget -O kubernetes-server-linux-amd64.tar.gz https://dl.k8s.io/v1.18.4/kubernetes-server-linux-amd64.tar.gz && tar -zxvf kubernetes-server-linux-amd64.tar.gz
cp kubernetes/server/bin/kubelet /usr/bin/kubelet
cp kubernetes/server/bin/kube-proxy /usr/bin/kube-proxy
cp kubernetes/server/bin/kubectl /usr/bin/kubectl
chmod a+x /usr/bin/kubelet && chmod a+x /usr/bin/kube-proxy && chmod a+x /usr/bin/kubectl
wget -O /etc/kubernetes/kubelet-config.yaml "${DOWNLOAD_ADDRESS}"/download/{{.ObjectMeta.Namespace}}/{{.ObjectMeta.Name}}/kubelet/kubelet-config.yaml
wget -O /lib/systemd/system/kubelet.service "${DOWNLOAD_ADDRESS}"/download/{{.ObjectMeta.Namespace}}/{{.ObjectMeta.Name}}/kubelet/kubelet.service
wget -O /etc/kubernetes/kubeproxy-config.yaml "${DOWNLOAD_ADDRESS}"/download/{{.ObjectMeta.Namespace}}/{{.ObjectMeta.Name}}/kube-proxy/kubeproxy-config.yaml
wget -O /lib/systemd/system/kubeproxy.service "${DOWNLOAD_ADDRESS}"/download/{{.ObjectMeta.Namespace}}/{{.ObjectMeta.Name}}/kube-proxy/kubeproxy.service
wget -O /etc/docker/daemon.json "${DOWNLOAD_ADDRESS}"/download/{{.ObjectMeta.Namespace}}/{{.ObjectMeta.Name}}/docker/daemon.json

#写hosts
#todo:如何解决访问问题
#echo '' >>/etc/hosts
#安装docker
curl -fsSL https://get.docker.com | bash -s docker --mirror Aliyun
service docker restart
#启动kubelet,kube-proxy
service kubelet start
service kubeproxy start

# curl -fsSL http://localhost:7788/download/test/test/node-install/install.sh | bash -s http://localhost:7788
