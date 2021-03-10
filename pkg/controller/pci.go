// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package controller

import (
	e2smrcpreies "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc_pre/v1/e2sm-rc-pre-ies"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-pci/pkg/store"
	"github.com/onosproject/onos-pci/pkg/utils/decode"
	"github.com/onosproject/onos-ric-sdk-go/pkg/e2/indication"
	"google.golang.org/protobuf/proto"
	"sync"
)

var logPci = logging.GetLogger("controller", "pci")

// PciCtrl is the controller for the KPI monitoring
type PciCtrl struct {
	IndChan           chan indication.Indication
	PciMetricMap      map[string]*store.CellPciNrt
	PciMetricMapMutex sync.RWMutex
	GlobalPciMap      map[string]int32
}

// NewPciController returns the struct for PCI logic
func NewPciController(indChan chan indication.Indication) *PciCtrl {
	logPci.Info("Start ONOS-PCI Application Controller")
	return &PciCtrl{
		IndChan:      indChan,
		PciMetricMap: make(map[string]*store.CellPciNrt),
		GlobalPciMap: make(map[string]int32),
	}
}

// Run starts to listen Indication message and then save the result to its struct
func (c *PciCtrl) Run() {
	c.listenIndChan()
}

func (c *PciCtrl) storePciMetric(header *e2smrcpreies.E2SmRcPreIndicationHeaderFormat1, message *e2smrcpreies.E2SmRcPreIndicationMessageFormat1) {
	logPci.Debugf("Header: %v", header)
	logPci.Debugf("PLMN ID: %d", decode.PlmnIdToUint32(header.GetCgi().GetEUtraCgi().GetPLmnIdentity().GetValue()))
	logPci.Debugf("ECID: %d", header.GetCgi().GetEUtraCgi().GetEUtracellIdentity().GetValue().GetValue())
	logPci.Debugf("ECID Length: %d", header.GetCgi().GetEUtraCgi().GetEUtracellIdentity().GetValue().GetLen())

	cgi := store.NewCGI(decode.PlmnIdToUint32(header.GetCgi().GetEUtraCgi().GetPLmnIdentity().GetValue()),
		header.GetCgi().GetEUtraCgi().GetEUtracellIdentity().GetValue().GetValue(),
		header.GetCgi().GetEUtraCgi().GetEUtracellIdentity().GetValue().GetLen())

	logPci.Debugf("Message: %v", message)
	logPci.Debugf("EARFCN DL: %d", message.GetDlEarfcn().GetValue())
	logPci.Debugf("Cell size: %v", message.GetCellSize().String())
	logPci.Debugf("PCI Pool: %v", message.GetPciPool())
	logPci.Debugf("PCI: %d", message.GetPci().GetValue())
	logPci.Debugf("Neighbors: %v", message.GetNeighbors())

	metric := store.NewCellMetric(message.GetDlEarfcn().GetValue(), message.GetCellSize(), message.GetPci().GetValue())

	var pciPoolList []*store.PciPool
	for i := 0; i < len(message.GetPciPool()); i++ {
		pciPool := store.NewPciPool(message.GetPciPool()[i].GetLowerPci().GetValue(), message.GetPciPool()[i].GetUpperPci().GetValue())
		pciPoolList = append(pciPoolList, pciPool)
	}

	var neighbors []*store.NeighborCell
	for i := 0; i < len(message.GetNeighbors()); i++ {
		neighborCgi := store.NewCGI(decode.PlmnIdToUint32(message.GetNeighbors()[i].GetCgi().GetEUtraCgi().GetPLmnIdentity().GetValue()),
			message.GetNeighbors()[i].GetCgi().GetEUtraCgi().GetEUtracellIdentity().GetValue().GetValue(),
			message.GetNeighbors()[i].GetCgi().GetEUtraCgi().GetEUtracellIdentity().GetValue().GetLen())

		neighborMetric := store.NewCellMetric(message.GetNeighbors()[i].GetDlEarfcn().GetValue(), message.GetNeighbors()[i].GetCellSize(), message.GetNeighbors()[i].GetPci().GetValue())

		neighbor := store.NewNeighborCell(message.GetNeighbors()[i].GetNrIndex(), neighborCgi, neighborMetric)

		neighbors = append(neighbors, neighbor)
		c.GlobalPciMap[decode.CgiToString(neighborCgi)] = message.GetNeighbors()[i].GetPci().GetValue()
	}

	cellPciNrt := store.NewCellPciNrt(metric, pciPoolList, neighbors)

	c.PciMetricMapMutex.Lock()
	c.PciMetricMap[decode.CgiToString(cgi)] = cellPciNrt
	pciArbitrator := NewPciArbitratorController(cgi, cellPciNrt)
	err := pciArbitrator.Start(c.PciMetricMap, c.GlobalPciMap)
	if err != nil {
		logPci.Errorf("PCI Arbitrator has an error - %v", err)
	}
	c.GlobalPciMap[decode.CgiToString(cgi)] = c.PciMetricMap[decode.CgiToString(cgi)].Metric.Pci
	c.PciMetricMapMutex.Unlock()
}

func (c *PciCtrl) listenIndChan() {
	var err error
	for indMsg := range c.IndChan {
		logPci.Debugf("Raw message: %v", indMsg)

		indHeaderByte := indMsg.Payload.Header
		indMessageByte := indMsg.Payload.Message

		indHeader := e2smrcpreies.E2SmRcPreIndicationHeader{}
		err = proto.Unmarshal(indHeaderByte, &indHeader)
		if err != nil {
			logPci.Errorf("Error - Unmarshalling header protobytes to struct: %v", err)
		}

		indMessage := e2smrcpreies.E2SmRcPreIndicationMessage{}
		err = proto.Unmarshal(indMessageByte, &indMessage)
		if err != nil {
			logPci.Errorf("Error - Unmarshalling message protobytes to struct: %v", err)
		}

		c.storePciMetric(indHeader.GetIndicationHeaderFormat1(), indMessage.GetIndicationMessageFormat1())
	}
}
