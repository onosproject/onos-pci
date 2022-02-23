// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package monitoring

import (
	"context"
	"fmt"
	"strconv"

	"github.com/onosproject/onos-pci/pkg/utils/parse"

	"github.com/onosproject/onos-pci/pkg/rnib"

	"github.com/onosproject/onos-pci/pkg/types"

	"github.com/onosproject/onos-pci/pkg/store/metrics"

	e2smrcpreies "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc_pre_go/v2/e2sm-rc-pre-v2-go"

	e2api "github.com/onosproject/onos-api/go/onos/e2t/e2/v1beta1"

	topoapi "github.com/onosproject/onos-api/go/onos/topo"

	appConfig "github.com/onosproject/onos-pci/pkg/config"

	"github.com/onosproject/onos-lib-go/pkg/logging"
	"google.golang.org/protobuf/proto"

	"github.com/onosproject/onos-pci/pkg/broker"
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
	cellSize := messageFormat1.GetCellSize()

	var pciPoolList []*types.PCIPool
	pciPool := &types.PCIPool{
		LowerPci: types.LowerPCI,
		UpperPci: types.UpperPCI,
	}

	pciPoolList = append(pciPoolList, pciPool)
	cellKey := metrics.NewKey(cellCGI)
	_, err = m.metricStore.Put(ctx, cellKey, metrics.Entry{
		Key: metrics.Key{
			CellGlobalID: cellCGI,
		},
		Value: types.CellPCI{
			E2NodeID: nodeID,
			Metric: &types.CellMetric{
				PCI:      cellPCI,
				CellSize: cellSize,
			},
			Neighbors:   messageFormat1.GetNeighbors(),
			PCIPoolList: pciPoolList,
		},
	})
	if err != nil {
		return err
	}

	cellID, err := parse.GetCellID(cellCGI)
	if err != nil {
		return err
	}
	cellTopoID := topoapi.ID(fmt.Sprintf("%s/%s", nodeID, strconv.FormatUint(cellID, 16)))
	err = m.rnibClient.UpdateCellAspects(ctx, cellTopoID, uint32(cellPCI), messageFormat1.GetNeighbors(), cellSize.String())

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
