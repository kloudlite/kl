FROM rancher/k3s:v1.31.1-k3s1

COPY --from=defnotgustavom/curl:latest /bin/curl /usr/local/bin/curl
RUN  mkdir -p /var/lib/rancher/k3s/agent/images/

COPY ./images/* /var/lib/rancher/k3s/agent/images/