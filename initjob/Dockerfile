FROM ubuntu
RUN apt update && apt install wget -y
RUN wget -O kubernetes-server-linux-amd64.tar.gz https://dl.k8s.io/v1.18.4/kubernetes-server-linux-amd64.tar.gz && tar -zxvf kubernetes-server-linux-amd64.tar.gz && rm kubernetes-server-linux-amd64.tar.gz
RUN wget -O /usr/bin/cfssl https://pkg.cfssl.org/R1.2/cfssl_linux-amd64 && wget -O /usr/bin/cfssljson https://pkg.cfssl.org/R1.2/cfssljson_linux-amd64 && wget -O /usr/bin/cfssl-certinfo https://pkg.cfssl.org/R1.2/cfssl-certinfo_linux-amd64 && chmod +x /usr/bin/cfssl*
RUN cp kubernetes/server/bin/kubectl /usr/bin/kubectl
ADD init.sh /home/init.sh
RUN chmod a+x /home/init.sh
WORKDIR /home
CMD ['sh /home/init.sh']