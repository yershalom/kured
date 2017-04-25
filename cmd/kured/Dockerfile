FROM ubuntu
ADD https://storage.googleapis.com/kubernetes-release/release/v1.6.2/bin/linux/amd64/kubectl /usr/bin/kubectl
RUN chmod 0755 /usr/bin/kubectl
COPY ./kured /usr/bin/kured
ENTRYPOINT ["/usr/bin/kured"]
