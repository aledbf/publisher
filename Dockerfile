FROM deis/base
ENV CGO_ENABLED 0
RUN apt-get update && apt-get install -yq git mercurial
ADD https://storage.googleapis.com/golang/go1.3.linux-amd64.tar.gz /tmp/
RUN tar -C /usr/local -xzf /tmp/go1.3.linux-amd64.tar.gz
ENV PATH /usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin:/go/bin
ADD . /go/src/github.com/aledbf/publisher
WORKDIR /go/src/github.com/aledbf/publisher
ENV GOPATH /go
RUN go get -a -ldflags '-s' github.com/aledbf/publisher
RUN mkdir -p /tmp/package && cp /go/bin/publisher /tmp/package/publisher
RUN tar -C /tmp/package -czf /tmp/publisher.tar.gz .
ENTRYPOINT ["/go/bin/publisher"]
