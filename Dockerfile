FROM golang:1.13 as builder
WORKDIR /app
COPY . .
RUN make build && bin/k3d version

FROM docker:19.03-dind

# TODO: we could create a different stage for e2e tests
RUN apk add bash curl sudo
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`/bin/linux/amd64/kubectl && \
    chmod +x ./kubectl && \
    mv ./kubectl /usr/local/bin/kubectl
COPY --from=builder /app/bin/k3d /bin/k3d
