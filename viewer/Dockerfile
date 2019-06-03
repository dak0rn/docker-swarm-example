# Development Dockerfile

FROM golang:1.11-alpine3.8

# Required dependencies for live-reloading
RUN apk add --update git make musl-dev gcc

# Install the live-reloading tool
RUN go get -u github.com/cortesi/modd/cmd/modd

# Because of the lack of module support the application is to be
# mounted into the $GOPATH source folder
VOLUME /go/src/github.com/dak0rn/docker-swarm-example/viewer
WORKDIR /go/src/github.com/dak0rn/docker-swarm-example/viewer

CMD ["make", "run"]
