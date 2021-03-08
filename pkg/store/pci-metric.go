// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package store

import (
	e2smrcpreies "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc_pre/v1/e2sm-rc-pre-ies"
)

// CGI is the ID for each cell
type CGI struct {
	PlmnID  uint32
	Ecid    uint64
	EcidLen uint32
}

// CellMetric is the metric struct which has EARFCN-DL, size, and PCI of a cell
type CellMetric struct {
	DlEarfcn int32
	CellSize e2smrcpreies.CellSize
	Pci      int32
}

// CellPciNrt is the PCI-NRT information
type CellPciNrt struct {
	Metric      *CellMetric
	PciPoolList []*PciPool
	Neighbors   []*NeighborCell
}

// PciPool is the PCI pool to be able to assign a PCI to a cell
type PciPool struct {
	LowerPci int32
	UpperPci int32
}

// NeighborCell is the struct including neighbor cell information
type NeighborCell struct {
	NrIndex int32
	Cgi     *CGI
	Metric  *CellMetric
}

// NewCGI makes a new CGI object and returns its address
func NewCGI(plmnID uint32, ecid uint64, ecidLen uint32) *CGI {
	return &CGI{
		PlmnID: plmnID,
		Ecid: ecid,
		EcidLen: ecidLen,
	}
}

// NewCellMetric makes a new CellMetric object and returns its address
func NewCellMetric(dlEarfcn int32, cellSize e2smrcpreies.CellSize, pci int32) *CellMetric {
	return &CellMetric{
		DlEarfcn: dlEarfcn,
		CellSize: cellSize,
		Pci: pci,
	}
}

// NewCellPciNrt makes a new CellPciNrt object and returns its address
func NewCellPciNrt(metric *CellMetric, pciPoolList []*PciPool, neighbors []*NeighborCell) *CellPciNrt {
	return &CellPciNrt{
		Metric: metric,
		PciPoolList: pciPoolList,
		Neighbors: neighbors,
	}
}

// NewPciPool makes a new PciPool object and returns its address
func NewPciPool(lowerPci int32, upperPci int32) *PciPool {
	return &PciPool {
		LowerPci: lowerPci,
		UpperPci: upperPci,
	}
}

// NewNeighborCell makes a new NeighborCell object and returns its address
func NewNeighborCell(nrIndex int32, cgi *CGI, metric *CellMetric) *NeighborCell {
	return &NeighborCell{
		NrIndex: nrIndex,
		Cgi: cgi,
		Metric: metric,
	}
}

