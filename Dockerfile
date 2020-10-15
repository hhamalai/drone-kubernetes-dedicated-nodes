FROM golang:1.13-alpine AS build-stage

WORKDIR /go/src/github.com/hhamalai/drone-kube-runner-tolerations
COPY . .

RUN CGO_ENABLED=0 go build -o /bin/drone-pod-admission-controller --ldflags "-w -extldflags '-static'"  cmd/webhook-server/main.go

# Final image.
FROM gcr.io/distroless/static-debian10
COPY --from=build-stage /bin/drone-pod-admission-controller /usr/local/bin/drone-pod-admission-controller
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/drone-pod-admission-controller"]