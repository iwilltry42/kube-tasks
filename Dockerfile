FROM golang:1.14 as builder
WORKDIR /app
COPY . .
RUN make build

FROM alpine:3.12
RUN apk --no-cache add ca-certificates
RUN addgroup -g 1001 -S kube-tasks \
    && adduser -u 1001 -D -S -G kube-tasks kube-tasks
USER kube-tasks
COPY --from=builder /app/bin/kube-tasks /usr/local/bin/kube-tasks
CMD ["kube-tasks"]
