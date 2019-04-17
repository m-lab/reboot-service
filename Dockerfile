FROM golang:alpine as reboot-api-build
RUN apk add --no-cache git
ADD . /go/src/github.com/m-lab/reboot-service
RUN /go/src/github.com/m-lab/reboot-service/build.sh

# Now copy the built image into the minimal base image
FROM alpine
COPY --from=reboot-api-build /go/bin/reboot-service /
WORKDIR /
ENTRYPOINT ["/reboot-service"]
