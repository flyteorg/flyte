curl https://storage.googleapis.com/kubernetes-release/release/v1.18.10/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl
chmod a+x /usr/local/bin/kubectl
curl -s https://raw.githubusercontent.com/rancher/k3d/main/install.sh | bash
k3d cluster create -p "30081:30081" --no-lb --k3s-server-arg '--no-deploy=traefik' --k3s-server-arg '--no-deploy=servicelb' flyte
kubectl config set-context flyte
