FROM deis/go:latest
MAINTAINER OpDemand <info@opdemand.com>

WORKDIR /go/src/github.com/aledbf/publisher
CMD /go/bin/publisher

ADD . /go/src/github.com/aledbf/publisher
RUN CGO_ENABLED=0 go get -a -ldflags '-s' github.com/aledbf/publisher
