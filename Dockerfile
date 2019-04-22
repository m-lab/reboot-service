FROM golang:1.12 as build
ADD . /go/src/github.com/m-lab/reboot-service
RUN go get \
    -v \
    -ldflags "-X github.com/m-lab/go/prometheusx.GitShortCommit=$(git log -1 --format=%h)" \
    github.com/m-lab/reboot-service

# Now copy the built image into the minimal base image
FROM alpine:3.9
COPY --from=build /go/bin/reboot-service /
WORKDIR /
ENTRYPOINT ["/reboot-service"]
