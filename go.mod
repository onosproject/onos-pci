module github.com/onosproject/onos-pci

go 1.14

require (
	github.com/onosproject/onos-api/go v0.7.16
	github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc_pre v0.7.11
	github.com/onosproject/onos-e2t v0.7.8
	github.com/onosproject/onos-lib-go v0.7.5
	github.com/onosproject/onos-ric-sdk-go v0.7.10
	github.com/stretchr/testify v1.7.0
	google.golang.org/grpc v1.33.2
	google.golang.org/protobuf v1.25.0
)

replace github.com/onosproject/onos-api/go => /Users/woojoong/workspace/onf/sd-ran/xapp/onos-api/go
