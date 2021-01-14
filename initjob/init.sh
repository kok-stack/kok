#!/bin/bash
set -e
# set ff=unix
echo '==================init variable==================='
echo "$APISERVER_ADDRESS"
echo "$FRONT_APISERVER_ADDRESS"
NAMESPACE=$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace)
echo "$NAMESPACE"

echo '==================generator ssl==================='

mkdir -p /home/test && cd /home/test
cat <<EOF >ca-config.json
{
    "signing":{
        "default":{
            "expiry":"876000h"
        },
        "profiles":{
            "etcd":{
                "usages":[
                    "signing",
                    "key encipherment",
                    "server auth",
                    "client auth"
                ],
                "expiry":"876000h"
            },
            "kubernetes":{
                "usages":[
                    "signing",
                    "key encipherment",
                    "server auth",
                    "client auth"
                ],
                "expiry":"876000h"
            }
        }
    }
}
EOF

cat <<EOF >ca-csr.json
{
  "CN": "CA",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "CN",
      "ST": "shenzhen",
      "L": "shenzhen",
      "O": "CA",
      "OU": "System"
    }
  ]
}
EOF

cfssl gencert -initca ca-csr.json | cfssljson -bare ca

ls -la

cat <<EOF >etcd-csr.json
{
  "CN": "etcd",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "hosts": [
    "example.net",
    "www.example.net"
  ],
  "names": [
    {
      "C": "CN",
      "ST": "chengdu",
      "L": "chengdu",
      "O": "etcd",
      "OU": "System"
    }
  ]
}
EOF

cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=etcd etcd-csr.json | cfssljson -bare etcd

cat <<EOF >k8s-client-csr.json
{
    "CN": "kubernetes-node",
    "hosts": [

    ],
    "key": {
        "algo": "rsa",
        "size": 2048
    },
    "names": [
        {
            "C": "CN",
            "ST": "chengdu",
            "L": "chengdu",
            "O": "system:node",
            "OU": "ops"
        }
    ]
}
EOF

cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes k8s-client-csr.json | cfssljson -bare kubernetes-node

cat <<EOF >k8s-server-csr.json
{
    "CN": "kubernetes-admin",
    "hosts": [
        "127.0.0.1",
        "localhost",
        "192.168.0.1",
        "${APISERVER_ADDRESS}",
        "${APISERVER_ADDRESS}.${NAMESPACE}",
        "${APISERVER_ADDRESS}.${NAMESPACE}.svc",
        "${APISERVER_ADDRESS}.${NAMESPACE}.svc.cluster",
        "${APISERVER_ADDRESS}.${NAMESPACE}.svc.cluster.local",
        "kubernetes.default",
        "kubernetes.default.svc",
        "kubernetes.default.svc.cluster",
        "kubernetes.default.svc.cluster.local",
        "${KUBE_SVC_ADDR}",
        "${FRONT_APISERVER_ADDRESS}"
    ],
    "key": {
        "algo": "rsa",
        "size": 2048
    },
    "names": [
        {
            "C": "CN",
            "ST": "chengdu",
            "L": "chengdu",
            "O": "system:masters",
            "OU": "ops"
        }
    ]
}
EOF

cfssl gencert -ca=ca.pem -ca-key=ca-key.pem -config=ca-config.json -profile=kubernetes k8s-server-csr.json | cfssljson -bare kubernetes-server

kubectl config --kubeconfig=admin.config set-cluster kubernetes --certificate-authority=/home/test/ca.pem --embed-certs=true --server=https://"${APISERVER_ADDRESS}":6443
kubectl config --kubeconfig=admin.config set-credentials kubernetes-admin --embed-certs=true --client-certificate=/home/test/kubernetes-server.pem --client-key=/home/test/kubernetes-server-key.pem
kubectl config --kubeconfig=admin.config set-context kubernetes --cluster=kubernetes --namespace=default --user=kubernetes-admin
kubectl config --kubeconfig=admin.config set current-context kubernetes

echo '==============admin.config======================='
cat admin.config
echo '==============admin.config======================='

#kubectl create secret generic pki --from-file=ca-config.json --from-file=ca-csr.json --from-file=ca-key.pem --from-file=ca.csr --from-file=ca.pem --from-file=etcd-csr.json --from-file=etcd-key.pem --from-file=etcd.csr --from-file=etcd.pem --from-file=k8s-client-csr.json --from-file=k8s-server-csr.json --from-file=kubernetes-node-key.pem --from-file=kubernetes-node.csr --from-file=kubernetes-node.pem --from-file=kubernetes-server-key.pem --from-file=kubernetes-server.csr --from-file=kubernetes-server.pem
kubectl create secret generic "${CA_PKI_NAME}" --from-file=ca-config.json --from-file=ca-csr.json --from-file=ca-key.pem --from-file=ca.csr --from-file=ca.pem
kubectl create secret generic "${ETCD_PKI_NAME}" --from-file=etcd-csr.json --from-file=etcd-key.pem --from-file=etcd.csr --from-file=etcd.pem
kubectl create secret generic "${K8S_SERVER_NAME}" --from-file=k8s-server-csr.json --from-file=kubernetes-server-key.pem --from-file=kubernetes-server.csr --from-file=kubernetes-server.pem
kubectl create secret generic "${K8S_CLIENT_NAME}" --from-file=k8s-client-csr.json --from-file=kubernetes-node-key.pem --from-file=kubernetes-node.csr --from-file=kubernetes-node.pem
kubectl create secret generic "${ADMIN_CONFIG_NAME}" --from-file=admin.config

kubectl config --kubeconfig=node.config set-cluster kubernetes --certificate-authority=/home/test/ca.pem --embed-certs=true --server=https://"${FRONT_APISERVER_ADDRESS}"."${NAMESPACE}":"${FRONT_APISERVER_PORT}"
kubectl config --kubeconfig=node.config set-credentials kubernetes-node --embed-certs=true --client-certificate=/home/test/kubernetes-node.pem --client-key=/home/test/kubernetes-node-key.pem
kubectl config --kubeconfig=node.config set-context kubernetes --cluster=kubernetes --namespace=default --user=kubernetes-node
kubectl config --kubeconfig=node.config set current-context kubernetes

echo '==============node.config======================='
cat node.config
echo '==============node.config======================='
kubectl create secret generic "${NODE_CONFIG_NAME}" --from-file=node.config
