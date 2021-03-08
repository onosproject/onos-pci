// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package controller

import (
	"fmt"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-pci/pkg/store"
)

var logArb = logging.GetLogger("controller", "arbitrator")

// PciArbitratorCtrl is the struct to assign appropriate PCI to E2Node
type PciArbitratorCtrl struct {
	TargetE2NodeCgi    *store.CGI
	TargetE2NodeMetric *store.CellPciNrt
	D1NeighborPciMap   map[*store.CGI]int32 // Key: neighbor's CGI / neighbor's PCI
	D2NeighborPciMap   map[*store.CGI]int32 // Key: neighbor's neighbor's CGI / neighbor's neighbor's PCI
	NeighborPcis       map[int32]bool
}

// NewPciArbitratorController returns the new PciArbitratorCtrl struct
func NewPciArbitratorController(targetE2NodeCgi *store.CGI, targetE2NodeMetric *store.CellPciNrt) *PciArbitratorCtrl {
	return &PciArbitratorCtrl{
		TargetE2NodeCgi:    targetE2NodeCgi,
		TargetE2NodeMetric: targetE2NodeMetric,
		D1NeighborPciMap:   make(map[*store.CGI]int32),
		D2NeighborPciMap:   make(map[*store.CGI]int32),
		NeighborPcis:       make(map[int32]bool),
	}
}

// Start is the function to run PCIArbitrator
func (a *PciArbitratorCtrl) Start(pciMetricMap map[*store.CGI]*store.CellPciNrt) error {
	return a.Run(pciMetricMap)
}

// Run is the main function to arbitrate PCI
func (a *PciArbitratorCtrl) Run(pciMetricMap map[*store.CGI]*store.CellPciNrt) error {
	var err error
	a.setD1NeighborPciMap(pciMetricMap)
	a.setD2NeighborPciMap(pciMetricMap)

	if a.verifyPci() {
		return nil
	}

	a.TargetE2NodeMetric.Metric.Pci, err = a.getUniquePci()
	if err != nil {
		return err
	}
	logArb.Infof("PCI of E2Node %v is assigned to %d", a.TargetE2NodeCgi, a.TargetE2NodeMetric.Metric.Pci)

	// push code to send control message
	return nil
}


func (a *PciArbitratorCtrl) getUniquePci() (int32, error) {
	for _, pool := range a.TargetE2NodeMetric.PciPoolList {
		for i := pool.LowerPci; i <= pool.UpperPci; i++ {
			if !a.NeighborPcis[i] {
				return i, nil
			}
		}
	}
	return -1, fmt.Errorf("there is no available PCIs in PCI Pool - target E2node information: CGI-%v, Message-%v", *a.TargetE2NodeCgi, *a.TargetE2NodeMetric)
}

func (a *PciArbitratorCtrl) verifyPci() bool {
	// Search D1Map
	for _, d1NPci := range a.D1NeighborPciMap {
		if d1NPci == a.TargetE2NodeMetric.Metric.Pci {
			return false
		}
	}

	// Search D2Map
	for _, d2NPci := range a.D2NeighborPciMap {
		if d2NPci == a.TargetE2NodeMetric.Metric.Pci {
			return false
		}
	}

	return true
}

func (a *PciArbitratorCtrl) setD1NeighborPciMap(pciMetricMap map[*store.CGI]*store.CellPciNrt) {
	for _, n := range pciMetricMap[a.TargetE2NodeCgi].Neighbors {
		if n.Cgi != a.TargetE2NodeCgi {
			a.D1NeighborPciMap[n.Cgi] = n.Metric.Pci
			a.NeighborPcis[n.Metric.Pci] = true
		}
	}
}

func (a *PciArbitratorCtrl) setD2NeighborPciMap(pciMetricMap map[*store.CGI]*store.CellPciNrt) {
	for n1 := range a.D1NeighborPciMap {
		for _, n2 := range pciMetricMap[n1].Neighbors {
			if n2.Cgi != a.TargetE2NodeCgi && (!a.hasNode(n2.Cgi, a.D1NeighborPciMap)) {
				a.D2NeighborPciMap[n2.Cgi] = n2.Metric.Pci
				a.NeighborPcis[n2.Metric.Pci] = true
			}
		}
	}
}

func (a *PciArbitratorCtrl) hasNode(source *store.CGI, m map[*store.CGI]int32) bool {
	_, ok := m[source]
	return ok
}
