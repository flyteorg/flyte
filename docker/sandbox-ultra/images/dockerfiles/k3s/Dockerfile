FROM rancher/k3s:v1.24.4-k3s1

COPY images/dockerfiles/k3s/bin /bin/
COPY bootstrap/output.yaml /var/lib/rancher/k3s/server/manifests/
COPY images/tar/$BUILDARCH /var/lib/rancher/k3s/agent/images/

ENTRYPOINT [ "/bin/k3d-entrypoint.sh" ]
CMD [ "server", "--disable=traefik", "--disable=servicelb" ]
