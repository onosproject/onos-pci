ARG ONOS_BUILD_VERSION=undefined

FROM onosproject/golang-build:v0.6.7 as build

ENV GO111MODULE=on
ARG ONOS_MAKE_TARGET=build

COPY Makefile go.mod go.sum /go/src/github.com/onosproject/onos-pci/
COPY cmd/ /go/src/github.com/onosproject/onos-pci/cmd/
COPY pkg/ /go/src/github.com/onosproject/onos-pci/pkg/
COPY vendor/ /go/src/github.com/onosproject/onos-pci/vendor/
COPY api/ /go/src/github.com/onosproject/onos-pci/api

RUN cd /go/src/github.com/onosproject/onos-pci && GOFLAGS=-mod=vendor make ${ONOS_MAKE_TARGET}

FROM alpine:3.11
RUN apk add libc6-compat

USER nobody

COPY --from=build /go/src/github.com/onosproject/onos-pci/build/_output/onos-pci /usr/local/bin/onos-pci

ENTRYPOINT ["onos-pci"]