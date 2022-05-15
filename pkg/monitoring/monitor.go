// SPDX-FileCopyrightText: 2022-present Intel Corporation
// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package monitoring

import (
	"context"
	e2api "github.com/onosproject/onos-api/go/onos/e2t/e2/v1beta1"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	e2smrc "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc/v1/e2sm-rc-ies"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-pci/pkg/broker"
	appConfig "github.com/onosproject/onos-pci/pkg/config"
	"github.com/onosproject/onos-pci/pkg/rnib"
	"github.com/onosproject/onos-pci/pkg/store/metrics"
	"github.com/onosproject/onos-pci/pkg/types"
	"google.golang.org/protobuf/proto"
)

var log = logging.GetLogger()

// NewMonitor creates a new indication monitor
func NewMonitor(opts ...Option) *Monitor {
	options := Options{}

	for _, opt := range opts {
		opt.apply(&options)
	}
	return &Monitor{
		streamReader: options.Monitor.StreamReader,
		appConfig:    options.App.AppConfig,
		metricStore:  options.App.MetricStore,
		nodeID:       options.Monitor.NodeID,
		rnibClient:   options.App.RNIBClient,
	}
}

// Monitor indication monitor
type Monitor struct {
	streamReader broker.StreamReader
	appConfig    *appConfig.AppConfig
	metricStore  metrics.Store
	nodeID       topoapi.ID
	rnibClient   rnib.Client
}

func (m *Monitor) processIndicationFormat3(ctx context.Context, indication e2api.Indication, nodeID topoapi.ID) error {
	header := e2smrc.E2SmRcIndicationHeader{}
	err := proto.Unmarshal(indication.Header, &header)
	if err != nil {
		return err
	}

	message := e2smrc.E2SmRcIndicationMessage{}
	err = proto.Unmarshal(indication.Payload, &message)
	if err != nil {
		return err
	}

	headerFormat1 := header.GetRicIndicationHeaderFormats().GetIndicationHeaderFormat1()
	messageFormat3 := message.GetRicIndicationMessageFormats().GetIndicationMessageFormat3()

	log.Debugf("Indication header format 1 %v", headerFormat1)
	log.Debugf("Indication message format 3 %v", messageFormat3)

	var pciPoolList []*types.PCIPool
	pciPool := &types.PCIPool{
		LowerPci: types.LowerPCI,
		UpperPci: types.UpperPCI,
	}
	pciPoolList = append(pciPoolList, pciPool)

	for _, cellInfo := range messageFormat3.GetCellInfoList() {
		if cellInfo.GetCellGlobalId().GetNRCgi() != nil {
			// 5G case
			cgi := cellInfo.GetCellGlobalId()
			if cellInfo.GetNeighborRelationTable().GetServingCellPci().GetNR() == nil {
				log.Errorf("PCI should be NR PCI but NR PCI field is empty in E2 Indication message")
				continue
			}
			pci := cellInfo.GetNeighborRelationTable().GetServingCellPci().GetNR().GetValue()
			if cellInfo.GetNeighborRelationTable().GetServingCellArfcn().GetNR() == nil {
				log.Errorf("ARFCN should be NR ARFCN but NR ARFCN field is empty in E2 indication message")
				continue
			}
			arfcn := cellInfo.GetNeighborRelationTable().GetServingCellArfcn().GetNR()
			key := metrics.NewKey(cgi)
			_, err := m.metricStore.Put(ctx, key, metrics.Entry{
				Key: metrics.Key{
					CellGlobalID: cgi,
				},
				Value: types.CellPCI{
					E2NodeID: nodeID,
					Metric: &types.CellMetric{
						PCI:   pci,
						ARFCN: arfcn.GetNRarfcn(),
					},
					Neighbors:   cellInfo.GetNeighborRelationTable().GetNeighborCellList().GetValue(),
					PCIPoolList: pciPoolList,
				},
			})
			if err != nil {
				return err
			}

			//cellID, err := parse.GetCellID(cgi)
			//if err != nil {
			//	return err
			//}
			//cellTopoID := topoapi.ID(fmt.Sprintf("%s/%s", nodeID, strconv.FormatUint(cellID, 16)))
			// ToDo: Update RNIB
		} else {
			// 4G case
			// ToDo: Add 4G case here
		}
	}
	return nil
}

func (m *Monitor) processIndication(ctx context.Context, indication e2api.Indication, nodeID topoapi.ID) error {
	err := m.processIndicationFormat3(ctx, indication, nodeID)
	if err != nil {
		log.Warn(err)
		return err
	}

	return nil
}

// Start start monitoring of indication messages for a given subscription ID
func (m *Monitor) Start(ctx context.Context) error {
	errCh := make(chan error)
	go func() {
		for {
			indMsg, err := m.streamReader.Recv(ctx)
			if err != nil {
				errCh <- err
			}
			err = m.processIndication(ctx, indMsg, m.nodeID)
			if err != nil {
				errCh <- err
			}
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
