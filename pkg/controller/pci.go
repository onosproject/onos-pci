// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package controller

import (
	"fmt"
	e2tapi "github.com/onosproject/onos-api/go/onos/e2t/e2"
	e2smrcpreies "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc_pre/v2/e2sm-rc-pre-v2"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-pci/pkg/southbound/ricapie2"
	"github.com/onosproject/onos-pci/pkg/store"
	"github.com/onosproject/onos-pci/pkg/utils/decode"
	"google.golang.org/protobuf/proto"
	"sync"
	"time"
)

var logPci = logging.GetLogger("controller", "pci")

const (
	LowerPCI   = 1
	UpperPCI   = 512
	CellIDLength		   = 36
)

// PciCtrl is the controller for the KPI monitoring
type PciCtrl struct {
	IndChan           chan *store.E2NodeIndication
	CtrlReqChans      map[string]chan *e2tapi.ControlRequest
	CtrlAckChan       chan *store.ControlAckMessages
	PciMetricMap      map[string]*store.CellPciNrt
	PciMetricMapMutex sync.RWMutex
	GlobalPciMap      map[string]int32
	CtrlAckTimer      int32 // ms
	// Monitor
	PciMonitor      map[string]*store.PciStat
	PciMonitorMutex *sync.RWMutex
}

// NewPciController returns the struct for PCI logic
func NewPciController(indChan chan *store.E2NodeIndication, ctrlReqChs map[string]chan *e2tapi.ControlRequest, ctrlAckChan chan *store.ControlAckMessages, pciMonitor map[string]*store.PciStat, pciMonitorMutex *sync.RWMutex, ctrlAckTimer int32) *PciCtrl {
	logPci.Info("Start ONOS-PCI Application Controller")
	return &PciCtrl{
		IndChan:         indChan,
		CtrlReqChans:    ctrlReqChs,
		CtrlAckChan:     ctrlAckChan,
		PciMetricMap:    make(map[string]*store.CellPciNrt),
		GlobalPciMap:    make(map[string]int32),
		PciMonitor:      pciMonitor,
		PciMonitorMutex: pciMonitorMutex,
		CtrlAckTimer:    ctrlAckTimer,
	}
}

// Run starts to listen Indication message and then save the result to its struct
func (c *PciCtrl) Run() {
	c.listenIndChan()
}

func (c *PciCtrl) getDlArfcn(message *e2smrcpreies.E2SmRcPreIndicationMessageFormat1) (int32, error){
	if message.GetDlArfcn().GetEArfcn() != nil {
		return message.GetDlArfcn().GetEArfcn().GetValue(), nil
	} else if message.GetDlArfcn().GetNrArfcn() != nil {
		return message.GetDlArfcn().GetNrArfcn().GetValue(), nil
	}

	return 0, fmt.Errorf("DlArfcn is empty")
}

func (c *PciCtrl) storePciMetric(header *e2smrcpreies.E2SmRcPreIndicationHeaderFormat1, message *e2smrcpreies.E2SmRcPreIndicationMessageFormat1, e2NodeID string) {
	logPci.Debugf("Header: %v", header)
	logPci.Debugf("PLMN ID: %d", decode.PlmnIdToUint32(header.GetCgi().GetEUtraCgi().GetPLmnIdentity().GetValue()))
	logPci.Debugf("ECID: %d", header.GetCgi().GetEUtraCgi().GetEUtracellIdentity().GetValue().GetValue())
	logPci.Debugf("ECID Length: %d", header.GetCgi().GetEUtraCgi().GetEUtracellIdentity().GetValue().GetLen())

	cgi := store.NewCGI(decode.PlmnIdToUint32(header.GetCgi().GetEUtraCgi().GetPLmnIdentity().GetValue()),
		header.GetCgi().GetEUtraCgi().GetEUtracellIdentity().GetValue().GetValue(),
		header.GetCgi().GetEUtraCgi().GetEUtracellIdentity().GetValue().GetLen())

	arfcn, err := c.getDlArfcn(message)
	if err != nil {
		logPci.Errorf("Cannot parse arfcn value: %v", err)
		return
	}

	logPci.Debugf("Message: %v", message)
	logPci.Debugf("ARFCN DL: %d", arfcn)
	logPci.Debugf("Cell size: %v", message.GetCellSize().String())
	logPci.Debugf("PCI: %d", message.GetPci().GetValue())
	logPci.Debugf("Neighbors: %v", message.GetNeighbors())

	metric := store.NewCellMetric(arfcn, message.GetCellSize(), message.GetPci().GetValue())

	var pciPoolList []*store.PciPool
	pciPool := store.NewPciPool(LowerPCI, UpperPCI)
	pciPoolList = append(pciPoolList, pciPool)

	var neighbors []*store.NeighborCell
	for i := 0; i < len(message.GetNeighbors()); i++ {
		neighborCgi := store.NewCGI(decode.PlmnIdToUint32(message.GetNeighbors()[i].GetCgi().GetEUtraCgi().GetPLmnIdentity().GetValue()),
			message.GetNeighbors()[i].GetCgi().GetEUtraCgi().GetEUtracellIdentity().GetValue().GetValue(),
			message.GetNeighbors()[i].GetCgi().GetEUtraCgi().GetEUtracellIdentity().GetValue().GetLen())

		neighborMetric := store.NewCellMetric(arfcn, message.GetNeighbors()[i].GetCellSize(), message.GetNeighbors()[i].GetPci().GetValue())
		neighbor := store.NewNeighborCell(int32(i), neighborCgi, neighborMetric)

		neighbors = append(neighbors, neighbor)
		if _, ok := c.GlobalPciMap[decode.CgiToString(neighborCgi)]; !ok {
			c.GlobalPciMap[decode.CgiToString(neighborCgi)] = message.GetNeighbors()[i].GetPci().GetValue()
		}
	}

	cellPciNrt := store.NewCellPciNrt(metric, pciPoolList, neighbors)

	c.PciMetricMapMutex.Lock()
	c.PciMetricMap[decode.CgiToString(cgi)] = cellPciNrt
	pciArbitrator := NewPciArbitratorController(cgi, cellPciNrt)

	// for Monitor
	c.PciMonitorMutex.Lock()
	if _, ok := c.PciMonitor[decode.CgiToString(cgi)]; !ok {
		c.PciMonitor[decode.CgiToString(cgi)] = &store.PciStat{
			NumConflicts: int32(0),
		}
	}
	c.PciMonitorMutex.Unlock()

	changed, err := pciArbitrator.ArbitratePCI(c.PciMetricMap, c.GlobalPciMap)
	if err != nil {
		logPci.Errorf("PCI Arbitrator has an error - %v", err)
	}
	if changed {
		// send control message to the E2Node
		e2smRcPreControlHandler := &ricapie2.E2SmRcPreControlHandler{
			NodeID:              e2NodeID,
			EncodingType:        e2tapi.EncodingType_PROTO,
			ServiceModelName:    ricapie2.ServiceModelName,
			ServiceModelVersion: ricapie2.ServiceModelVersion,
			ControlAckRequest:   e2tapi.ControlAckRequest_ACK,
		}
		cellID := header.GetCgi().GetEUtraCgi().GetEUtracellIdentity().GetValue().GetValue()
		//cellIdLen := header.GetCgi().GetEUtraCgi().GetEUtracellIdentity().GetValue().GetLen()
		controlPriority := int32(10) //hard-coded at this moment; should be replaced with the other value
		plmnID := header.GetCgi().GetEUtraCgi().GetPLmnIdentity().GetValue()
		e2smRcPreControlHandler.ControlHeader, err = e2smRcPreControlHandler.CreateRcControlHeader(cellID, CellIDLength, controlPriority, plmnID)
		if err != nil {
			logPci.Errorf("Error when generating control header - %v", err)
		}
		ranParamID := int32(1)
		ranParamName := "pci"
		ranParamValue := cellPciNrt.Metric.Pci
		e2smRcPreControlHandler.ControlMessage, err = e2smRcPreControlHandler.CreateRcControlMessage(ranParamID, ranParamName, ranParamValue)
		if err != nil {
			logPci.Errorf("Error when generating control message - %v", err)
		}
		controlRequest, err := e2smRcPreControlHandler.CreateRcControlRequest()
		if err != nil {
			logPci.Errorf("Error when generating control request - %v", err)
		}
		logPci.Debugf("Control Request message for e2NodeID %s: %v", e2NodeID, controlRequest)

		c.CtrlReqChans[e2NodeID] <- controlRequest
		err = c.waitAck()
		if err != nil {
			logPci.Errorf("Control ACK error: %v", err)
			// ToDo: Add handler when ACK does not arrive or Failure happens
		}
	}
	c.GlobalPciMap[decode.CgiToString(cgi)] = c.PciMetricMap[decode.CgiToString(cgi)].Metric.Pci
	c.PciMetricMapMutex.Unlock()

	// for Monitor
	if changed {
		c.PciMonitorMutex.Lock()
		c.PciMonitor[decode.CgiToString(cgi)].NumConflicts++
		c.PciMonitorMutex.Unlock()
	}
	c.PciMonitorMutex.RLock()
	logPci.Debugf("Num conflicts for %v: %d", decode.CgiToString(cgi), c.PciMonitor[decode.CgiToString(cgi)].NumConflicts)
	c.PciMonitorMutex.RUnlock()
}

func (c *PciCtrl) listenIndChan() {
	var err error
	for indMsg := range c.IndChan {
		logPci.Debugf("Raw message: %v", indMsg)

		indHeaderByte := indMsg.IndMsg.Payload.Header
		indMessageByte := indMsg.IndMsg.Payload.Message

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

		c.storePciMetric(indHeader.GetIndicationHeaderFormat1(), indMessage.GetIndicationMessageFormat1(), indMsg.NodeID)
	}
}

func (c *PciCtrl) waitAck() error {
	select {
	case ctrlAck := <- c.CtrlAckChan:
		if ctrlAck.CtrlAckFail {
			return fmt.Errorf("Control ACK failed or failure message arrived - ctrlAck:%v; ctrlFailure:%v", ctrlAck.CtrlACK, ctrlAck.CtrlFailure)
		}
		return nil
	case <- time.After(time.Duration(c.CtrlAckTimer) * time.Millisecond):
		return fmt.Errorf("CtrlAckTimer expired - Control ACK did not arrive in time")
	}
}