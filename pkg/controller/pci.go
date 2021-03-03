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

var log = logging.GetLogger("controller", "pci")

// PciCtrl is the controller for the KPI monitoring
type PciCtrl struct {
	IndChan           chan indication.Indication
	PciMetricMap      map[*store.CGI]*store.CellPciNrt
	PciMetricMapMutex sync.RWMutex
}

func NewPciController(indChan chan indication.Indication) *PciCtrl {
	log.Info("Start ONOS-PCI Application Controller")
	return &PciCtrl{
		IndChan:      indChan,
		PciMetricMap: make(map[*store.CGI]*store.CellPciNrt),
	}
}

func (c *PciCtrl) Run() {
	c.listenIndChan()
}

func (c *PciCtrl) storePciMetric(header *e2smrcpreies.E2SmRcPreIndicationHeaderFormat1, message *e2smrcpreies.E2SmRcPreIndicationMessageFormat1) {
	log.Debugf("Header: %v", header)
	log.Debugf("PLMN ID: %d", decode.PlmnIdToUint32(header.GetCgi().GetEUtraCgi().GetPLmnIdentity().GetValue()))
	log.Debugf("ECID: %d", header.GetCgi().GetEUtraCgi().GetEUtracellIdentity().GetValue().GetValue())
	log.Debugf("ECID Length: %d", header.GetCgi().GetEUtraCgi().GetEUtracellIdentity().GetValue().GetLen())

	cgi := store.NewCGI(decode.PlmnIdToUint32(header.GetCgi().GetEUtraCgi().GetPLmnIdentity().GetValue()),
		header.GetCgi().GetEUtraCgi().GetEUtracellIdentity().GetValue().GetValue(),
		header.GetCgi().GetEUtraCgi().GetEUtracellIdentity().GetValue().GetLen())

	log.Debugf("Message: %v", message)
	log.Debugf("EARFCN DL: %d", message.GetDlEarfcn().GetValue())
	log.Debugf("Cell size: %v", message.GetCellSize().String())
	log.Debugf("PCI Pool: %v", message.GetPciPool())
	log.Debugf("PCI: %d", message.GetPci().GetValue())
	log.Debugf("Neighbors: %v", message.GetNeighbors())

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

		neighborMetric := store.NewCellMetric(message.GetNeighbors()[i].GetDlEarfcn().GetValue(), message.GetNeighbors()[i].GetCellSize(), message.GetPci().GetValue())

		neighbor := store.NewNeighborCell(message.GetNeighbors()[i].GetNrIndex(), neighborCgi, neighborMetric)

		neighbors = append(neighbors, neighbor)
	}

	cellPciNrt := store.NewCellPciNrt(metric, pciPoolList, neighbors)

	c.PciMetricMapMutex.Lock()
	c.PciMetricMap[cgi] = cellPciNrt
	c.PciMetricMapMutex.Unlock()
}

func (c *PciCtrl) listenIndChan() {
	var err error
	for indMsg := range c.IndChan {
		log.Debugf("Raw message: %v", indMsg)

		indHeaderByte := indMsg.Payload.Header
		indMessageByte := indMsg.Payload.Message

		indHeader := e2smrcpreies.E2SmRcPreIndicationHeader{}
		err = proto.Unmarshal(indHeaderByte, &indHeader)
		if err != nil {
			log.Errorf("Error - Unmarshalling header protobytes to struct: %v", err)
		}

		indMessage := e2smrcpreies.E2SmRcPreIndicationMessage{}
		err = proto.Unmarshal(indMessageByte, &indMessage)
		if err != nil {
			log.Errorf("Error - Unmarshalling message protobytes to struct: %v", err)
		}

		go c.storePciMetric(indHeader.GetIndicationHeaderFormat1(), indMessage.GetIndicationMessageFormat1())
	}
}
