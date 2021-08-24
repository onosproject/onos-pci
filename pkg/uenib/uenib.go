// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package uenib

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	types2 "github.com/gogo/protobuf/types"
	"github.com/onosproject/onos-api/go/onos/uenib"
	e2sm_rc_pre_v2 "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc_pre/v2/e2sm-rc-pre-v2"
	"github.com/onosproject/onos-lib-go/pkg/certs"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-lib-go/pkg/southbound"
	"github.com/onosproject/onos-pci/pkg/rnib"
	"github.com/onosproject/onos-pci/pkg/store/metrics"
	"github.com/onosproject/onos-pci/pkg/utils/decode"
	"github.com/onosproject/onos-pci/pkg/utils/parse"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	// UENIBAddress has UENIB endpoint
	UENIBAddress = "onos-uenib:5150"
)

var log = logging.GetLogger("uenib")

func NewUENIBDialOpt(certPath string, keyPath string) ([]grpc.DialOption, error) {
	dialOpts := []grpc.DialOption{
		grpc.WithStreamInterceptor(southbound.RetryingStreamClientInterceptor(100 * time.Millisecond)),
		grpc.WithUnaryInterceptor(southbound.RetryingUnaryClientInterceptor()),
	}
	if certPath != "" && keyPath != "" {
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, err
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
		})))
	} else {
		// Load default Certificates
		cert, err := tls.X509KeyPair([]byte(certs.DefaultClientCrt), []byte(certs.DefaultClientKey))
		if err != nil {
			log.Errorf("failed to make tls key pair: %v", err)
			return nil, err
		}
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: true,
		})))
	}

	return dialOpts, nil
}

func NewUENIBClient(ctx context.Context, store metrics.Store, certPath string, keyPath string) Client {
	dialOpts, err := NewUENIBDialOpt(certPath, keyPath)
	if err != nil {
		log.Error(err)
	}
	conn, err := grpc.Dial(UENIBAddress, dialOpts...)
	if err != nil {
		log.Error(err)
	}
	rnibClient, err := rnib.NewClient()
	if err != nil {
		log.Error(err)
	}
	return Client{
		uenibClient:  uenib.NewUEServiceClient(conn),
		rnibClient:   rnibClient,
		metricsStore: store,
	}
}

type Client struct {
	uenibClient  uenib.UEServiceClient
	rnibClient   rnib.Client
	metricsStore metrics.Store
}

func (c *Client) Run(ctx context.Context) {
	go c.watchMetricStore(ctx)
}

func (c *Client) watchMetricStore(ctx context.Context) {
	ch := make(chan metrics.Event)
	err := c.metricsStore.Watch(ctx, ch)
	if err != nil {
		log.Error(err)
	}
	for e := range ch {
		// new indication message arrives
		if e.Type == metrics.Created {
			err := c.storeNeighborCellList(ctx, e.Value)
			if err != nil {
				log.Errorf("Error happened when storing neighbors to UENIB: %v", err)
			}
		}
	}
}

func (c *Client) storeNeighborCellList(ctx context.Context, entry metrics.Entry) error {
	uenibReq, err := c.createUENIBUpdateRequest(entry)
	if err != nil {
		return err
	}
	log.Debugf("UENIB Request message: uenibReq: %v", uenibReq)
	resp, err := c.uenibClient.UpdateUE(ctx, uenibReq)
	if err != nil {
		return err
	}
	log.Debugf("UENIB Response message: %v", resp)
	return nil
}

func (c *Client) createUENIBUpdateRequest(entry metrics.Entry) (*uenib.UpdateUERequest, error) {
	entryKey := entry.Key
	entryValue := entry.Value
	plmnIDByte, cid, cType, err := parse.GetMetricKey(entryKey.CellGlobalID)
	if err != nil {
		return nil, err
	}
	plmnID := decode.PlmnIDToUint32(plmnIDByte)
	nodeID := entryValue.E2NodeID

	uenibKey := fmt.Sprintf("%s:%x:%x:%s", nodeID, plmnID, cid, cType.String())
	uenibValue, err := c.encodeNeighborListToString(entryValue.Neighbors)
	if err != nil {
		return nil, err
	}
	log.Debugf("Stored UENIB Key:%v, value:%v", uenibKey, uenibValue)

	uenibObj := uenib.UE{
		ID:      uenib.ID(uenibKey),
		Aspects: make(map[string]*types2.Any),
	}

	uenibObj.Aspects["neighbors"] = &types2.Any{
		TypeUrl: "neighbors",
		Value:   []byte(uenibValue),
	}

	return &uenib.UpdateUERequest{
		UE: uenibObj,
	}, nil
}

func (c *Client) encodeNeighborListToString(neighbors []*e2sm_rc_pre_v2.Nrt) (string, error) {
	encNeighbors := ""

	for i := 0; i < len(neighbors); i++ {
		n := neighbors[i]
		nPlmnIDByte, nCid, nCType, err := parse.GetMetricKey(n.GetCgi())
		if err != nil {
			return "", err
		}
		nPlmnID := decode.PlmnIDToUint32(nPlmnIDByte)
		if i == 0 {
			encNeighbors = fmt.Sprintf("%x:%x:%s", nPlmnID, nCid, nCType.String())
			continue
		}
		encNeighbors = encNeighbors + "," + fmt.Sprintf("%x:%x:%s", nPlmnID, nCid, nCType.String())
	}

	return encNeighbors, nil
}
