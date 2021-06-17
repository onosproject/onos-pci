// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package monitoring

import (
	"context"

	"github.com/onosproject/onos-pci/pkg/types"

	"github.com/onosproject/onos-pci/pkg/store/metrics"

	e2smrcpreies "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc_pre/v2/e2sm-rc-pre-v2"

	e2api "github.com/onosproject/onos-api/go/onos/e2t/e2/v1beta1"

	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	e2client "github.com/onosproject/onos-ric-sdk-go/pkg/e2/v1beta1"

	appConfig "github.com/onosproject/onos-pci/pkg/config"

	"github.com/onosproject/onos-lib-go/pkg/logging"
	"google.golang.org/protobuf/proto"

	"github.com/onosproject/onos-pci/pkg/broker"
)

var log = logging.GetLogger("monitoring")

// NewMonitor creates a new indication monitor
func NewMonitor(streams broker.Broker,
	appConfig *appConfig.AppConfig, pciStore metrics.Store) *Monitor {
	return &Monitor{
		streams:   streams,
		appConfig: appConfig,
		pciStore:  pciStore,
	}
}

// Monitor indication monitor
type Monitor struct {
	streams   broker.Broker
	appConfig *appConfig.AppConfig
	pciStore  metrics.Store
}

func (m *Monitor) processIndicationFormat1(ctx context.Context, indication e2api.Indication, nodeID topoapi.ID) error {
	header := e2smrcpreies.E2SmRcPreIndicationHeader{}
	err := proto.Unmarshal(indication.Header, &header)
	if err != nil {
		return err
	}

	message := e2smrcpreies.E2SmRcPreIndicationMessage{}
	err = proto.Unmarshal(indication.Payload, &message)
	if err != nil {
		return err
	}

	headerFormat1 := header.GetIndicationHeaderFormat1()
	messageFormat1 := message.GetIndicationMessageFormat1()

	log.Debugf("Indication header format 1 %v", headerFormat1)
	log.Debugf("Indication message format 1 %v", messageFormat1)

	cellCGI := headerFormat1.GetCgi()
	cellPCI := messageFormat1.GetPci().GetValue()

	var pciPoolList []*types.PCIPool
	pciPool := &types.PCIPool{
		LowerPci: types.LowerPCI,
		UpperPci: types.UpperPCI,
	}

	pciPoolList = append(pciPoolList, pciPool)
	cellKey := metrics.NewKey(cellCGI)
	_, err = m.pciStore.Put(ctx, cellKey, types.CellPCI{
		Metric: &types.CellMetric{
			PCI: cellPCI,
		},
		Neighbors:   messageFormat1.GetNeighbors(),
		PCIPoolList: pciPoolList,
	})
	if err != nil {
		return err
	}

	return nil
}

func (m *Monitor) processIndication(ctx context.Context, indication e2api.Indication, nodeID topoapi.ID) error {
	err := m.processIndicationFormat1(ctx, indication, nodeID)
	if err != nil {
		log.Warn(err)
		return err
	}

	return nil
}

// Start start monitoring of indication messages for a given subscription ID
func (m *Monitor) Start(ctx context.Context, node e2client.Node, e2sub e2api.Subscription, nodeID topoapi.ID) error {
	streamReader, err := m.streams.OpenReader(node, e2sub)
	if err != nil {
		return err
	}

	for {
		indMsg, err := streamReader.Recv(ctx)
		if err != nil {
			return err
		}
		err = m.processIndication(ctx, indMsg, nodeID)
		if err != nil {
			return err
		}
	}
}
