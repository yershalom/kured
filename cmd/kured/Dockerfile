FROM alpine
ADD https://storage.googleapis.com/kubernetes-release/release/v1.6.2/bin/darwin/amd64/kubectl /usr/bin/kubectl
COPY ./kured /usr/bin/kured
ENTRYPOINT ["/usr/bin/kured"]
