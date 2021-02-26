// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package controller

import (
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-ric-sdk-go/pkg/e2/indication"
	"sync"
)

var log = logging.GetLogger("ctrl-pci")

// KpiMonCtrl is the controller for the KPI monitoring
type PciCtrl struct {
	IndChan  chan indication.Indication
	PciMutex sync.RWMutex
}

// CellIdentity is the ID for each cell
type CellIdentity struct {
	CuCpName string
	PlmnID   string
	NodeID   string
}

func NewPciController(indChan chan indication.Indication) *PciCtrl {
	log.Info("Start ONOS-PCI Application Controller")
	return &PciCtrl{
		IndChan: indChan,
	}
}

func (c *PciCtrl) Run() {
	c.listenIndChan()
}

func (c *PciCtrl) listenIndChan() {
	for indMsg := range c.IndChan {
		log.Infof("Raw message: %v", indMsg)
	}
}
