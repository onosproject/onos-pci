// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package controller

import (
	"fmt"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-pci/pkg/store"
	"github.com/onosproject/onos-pci/pkg/utils/decode"
)

var logArb = logging.GetLogger("controller", "arbitrator")

// PciArbitratorCtrl is the struct to assign appropriate PCI to E2Node
type PciArbitratorCtrl struct {
	TargetE2NodeCgi    *store.CGI
	TargetE2NodeMetric *store.CellPciNrt
	D1NeighborPciMap   map[string]int32 // Key: neighbor's CGI / neighbor's PCI
	D2NeighborPciMap   map[string]int32 // Key: neighbor's neighbor's CGI / neighbor's neighbor's PCI
	NeighborPcis       map[int32]bool
}

// NewPciArbitratorController returns the new PciArbitratorCtrl struct
func NewPciArbitratorController(targetE2NodeCgi *store.CGI, targetE2NodeMetric *store.CellPciNrt) *PciArbitratorCtrl {
	return &PciArbitratorCtrl{
		TargetE2NodeCgi:    targetE2NodeCgi,
		TargetE2NodeMetric: targetE2NodeMetric,
		D1NeighborPciMap:   make(map[string]int32),
		D2NeighborPciMap:   make(map[string]int32),
		NeighborPcis:       make(map[int32]bool),
	}
}

// Start is the function to run PCIArbitrator
func (a *PciArbitratorCtrl) Start(pciMetricMap map[string]*store.CellPciNrt, globalPciMap map[string]int32) error {
	return a.Run(pciMetricMap, globalPciMap)
}

// Run is the main function to arbitrate PCI
func (a *PciArbitratorCtrl) Run(pciMetricMap map[string]*store.CellPciNrt, globalPciMap map[string]int32) error {
	var err error
	a.setD1NeighborPciMap(pciMetricMap, globalPciMap)
	a.setD2NeighborPciMap(pciMetricMap, globalPciMap)

	logArb.Infof("Global PCI Map: %v", globalPciMap)
	logArb.Infof("D1 Neighbor PCIs: %v", a.D1NeighborPciMap)
	logArb.Infof("D2 Neighbor PCIs: %v", a.D2NeighborPciMap)

	if a.verifyPci() {
		return nil
	}

	a.TargetE2NodeMetric.Metric.Pci, err = a.getUniquePci()
	if err != nil {
		return err
	}
	logArb.Infof("PCI of E2Node %v is assigned to %d", decode.CgiToString(a.TargetE2NodeCgi), a.TargetE2NodeMetric.Metric.Pci)

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

func (a *PciArbitratorCtrl) setD1NeighborPciMap(pciMetricMap map[string]*store.CellPciNrt, globalPciMap map[string]int32) {
	for _, n := range pciMetricMap[decode.CgiToString(a.TargetE2NodeCgi)].Neighbors {
		if decode.CgiToString(n.Cgi) != decode.CgiToString(a.TargetE2NodeCgi) {
			a.D1NeighborPciMap[decode.CgiToString(n.Cgi)] = globalPciMap[decode.CgiToString(n.Cgi)]
			a.NeighborPcis[n.Metric.Pci] = true
		}
	}
}

func (a *PciArbitratorCtrl) setD2NeighborPciMap(pciMetricMap map[string]*store.CellPciNrt, globalPciMap map[string]int32) {
	for n1 := range a.D1NeighborPciMap {
		if _, ok := pciMetricMap[n1]; !ok {
			continue
		} else if pciMetricMap[n1].Neighbors == nil {
			continue
		}
		for _, n2 := range pciMetricMap[n1].Neighbors {
			if decode.CgiToString(n2.Cgi) != decode.CgiToString(a.TargetE2NodeCgi) && (!a.hasPci(n2.Cgi, globalPciMap)) {
				a.D2NeighborPciMap[decode.CgiToString(n2.Cgi)] = globalPciMap[decode.CgiToString(n2.Cgi)]
				a.NeighborPcis[n2.Metric.Pci] = true
			}
		}
	}
}

func (a *PciArbitratorCtrl) hasPci(source *store.CGI, m map[string]int32) bool {
	if _, ok := m[decode.CgiToString(source)]; ok {
		return true
	}
	return false
}
