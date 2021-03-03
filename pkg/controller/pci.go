// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package controller

import (
	"github.com/golang/protobuf/proto"
	e2smrcpreies "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc_pre/v1/e2sm-rc-pre-ies"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-pci/pkg/utils/decode"
	"github.com/onosproject/onos-ric-sdk-go/pkg/e2/indication"
	"sync"
)

var log = logging.GetLogger("ctrl-pci")

// KpiMonCtrl is the controller for the KPI monitoring
type PciCtrl struct {
	IndChan  chan indication.Indication
	PciMetricMap map[CGI]CellPciNrt
	PciMetricMapMutex sync.RWMutex
}

// CellCGI is the ID for each cell
type CGI struct {
	PlmnID uint32
	Ecid uint64
	EcidLen uint32
}

type CellMetric struct {
	DlEarfcn int32
	CellSize e2smrcpreies.CellSize
	Pci int32
}

type CellPciNrt struct {
	Metric CellMetric
	PciPoolList []PciPool
	Neighbors []NeighborCell
}

type PciPool struct {
	LowerPci int32
	UpperPci int32
}

type NeighborCell struct {
	NrIndex int32
	Cgi CGI
	Metric CellMetric
}

func NewPciController(indChan chan indication.Indication) *PciCtrl {
	log.Info("Start ONOS-PCI Application Controller")
	return &PciCtrl{
		IndChan: indChan,
		PciMetricMap: make(map[CGI]CellPciNrt),
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

	cgi := CGI{
		PlmnID: decode.PlmnIdToUint32(header.GetCgi().GetEUtraCgi().GetPLmnIdentity().GetValue()),
		Ecid: header.GetCgi().GetEUtraCgi().GetEUtracellIdentity().GetValue().GetValue(),
		EcidLen: header.GetCgi().GetEUtraCgi().GetEUtracellIdentity().GetValue().GetLen(),
	}

	log.Debugf("Message: %v", message)
	log.Debugf("EARFCN DL: %d", message.GetDlEarfcn().GetValue())
	log.Debugf("Cell size: %v", message.GetCellSize().String())
	log.Debugf("PCI Pool: %v", message.GetPciPool())
	log.Debugf("PCI: %d", message.GetPci().GetValue())
	log.Debugf("Neighbors: %v", message.GetNeighbors())

	metric := CellMetric{
		DlEarfcn: message.GetDlEarfcn().GetValue(),
		CellSize: message.GetCellSize(),
		Pci: message.GetPci().GetValue(),
	}

	var pciPoolList []PciPool
	for i := 0; i < len(message.GetPciPool()); i++ {
		pciPool := PciPool {
			LowerPci: message.GetPciPool()[i].GetLowerPci().GetValue(),
			UpperPci: message.GetPciPool()[i].GetUpperPci().GetValue(),
		}
		pciPoolList = append(pciPoolList, pciPool)
	}

	var neighbors []NeighborCell
	for i := 0; i < len(message.GetNeighbors()); i++ {
		neighborCgi := CGI{
			PlmnID: decode.PlmnIdToUint32(message.GetNeighbors()[i].GetCgi().GetEUtraCgi().GetPLmnIdentity().GetValue()),
			Ecid: message.GetNeighbors()[i].GetCgi().GetEUtraCgi().GetEUtracellIdentity().GetValue().GetValue(),
			EcidLen: message.GetNeighbors()[i].GetCgi().GetEUtraCgi().GetEUtracellIdentity().GetValue().GetLen(),
		}
		neighborMetric := CellMetric {
			DlEarfcn: message.GetNeighbors()[i].GetDlEarfcn().GetValue(),
			CellSize: message.GetNeighbors()[i].GetCellSize(),
			Pci: message.GetPci().GetValue(),
		}
		neighbor := NeighborCell{
			NrIndex: message.GetNeighbors()[i].GetNrIndex(),
			Cgi: neighborCgi,
			Metric: neighborMetric,
		}
		neighbors = append(neighbors, neighbor)
	}

	cellPciNrt := CellPciNrt{
		Metric: metric,
		PciPoolList: pciPoolList,
		Neighbors: neighbors,
	}

	c.PciMetricMapMutex.Lock()
	c.PciMetricMap[cgi] = cellPciNrt
	log.Infof("PciNrt: %v", c.PciMetricMap)
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
