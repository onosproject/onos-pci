// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package store

import (
	"fmt"
	e2smrcpreies "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc_pre/v1/e2sm-rc-pre-ies"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestNewCGI(t *testing.T) {
	assert.Equal(t, 1, 1)
	sampleCGI := CGI{
		PlmnID: uint32(1279014),
		Ecid: uint64(82530),
		EcidLen: uint32(28),
	}
	newCGI := NewCGI(uint32(1279014), uint64(82530), uint32(28))
	fmt.Printf("sampleCGI: %v\n", sampleCGI)
	fmt.Printf("newCGI: %v\n", *newCGI)
	assert.Equal(t, sampleCGI, *newCGI)
}

func TestNewCellMetric(t *testing.T) {
	sampleCellMetric := CellMetric{
		Pci: int32(1),
		CellSize: e2smrcpreies.CellSize_CELL_SIZE_FEMTO,
		DlEarfcn: int32(1),
	}
	newCellMetric := NewCellMetric(int32(1), e2smrcpreies.CellSize_CELL_SIZE_FEMTO, int32(1))
	fmt.Printf("sampleCellMetric: %v\n", sampleCellMetric)
	fmt.Printf("newCellMetric: %v\n", *newCellMetric)
	assert.Equal(t, sampleCellMetric, *newCellMetric)
}

func TestNewPciPool(t *testing.T) {
	samplePciPool := PciPool{
		LowerPci: int32(1),
		UpperPci: int32(10),
	}
	newPciPool := NewPciPool(int32(1), int32(10))
	fmt.Printf("samplePciPool: %v\n", samplePciPool)
	fmt.Printf("newPciPool: %v\n", newPciPool)
	assert.Equal(t, samplePciPool, *newPciPool)
}

func TestNewNeighborCell(t *testing.T) {
	sampleNeighborCell := NeighborCell{
		Metric: NewCellMetric(int32(1), e2smrcpreies.CellSize_CELL_SIZE_FEMTO, int32(1)),
		Cgi: NewCGI(uint32(1279014), uint64(82530), uint32(28)),
		NrIndex: 0,
	}
	newNeighborCell := NewNeighborCell(0, NewCGI(uint32(1279014), uint64(82530), uint32(28)), NewCellMetric(int32(1), e2smrcpreies.CellSize_CELL_SIZE_FEMTO, int32(1)))
	fmt.Printf("sampleNeighborCell: %v\n", sampleNeighborCell)
	fmt.Printf("newNeighborCell: %v\n", newNeighborCell)
	assert.Equal(t, sampleNeighborCell, *newNeighborCell)
}

func TestNewCellPciNrt(t *testing.T) {
	sampleNeighborCell1 := NeighborCell{
		Metric: NewCellMetric(int32(1), e2smrcpreies.CellSize_CELL_SIZE_FEMTO, int32(1)),
		Cgi: NewCGI(uint32(1279014), uint64(82530), uint32(28)),
		NrIndex: 0,
	}
	sampleNeighborCell2 := NeighborCell{
		Metric: NewCellMetric(int32(1), e2smrcpreies.CellSize_CELL_SIZE_FEMTO, int32(2)),
		Cgi: NewCGI(uint32(1279015), uint64(82530), uint32(28)),
		NrIndex: 1,
	}
	sampleNeighborCell3 := NeighborCell{
		Metric: NewCellMetric(int32(1), e2smrcpreies.CellSize_CELL_SIZE_FEMTO, int32(3)),
		Cgi: NewCGI(uint32(1279016), uint64(82530), uint32(28)),
		NrIndex: 2,
	}
	samplePciPool1 := PciPool{
		LowerPci: int32(1),
		UpperPci: int32(10),
	}
	samplePciPool2 := PciPool{
		LowerPci: int32(21),
		UpperPci: int32(30),
	}
	sampleNeighbors := []*NeighborCell{&sampleNeighborCell1, &sampleNeighborCell2, &sampleNeighborCell3}
	samplePciPools := []*PciPool{&samplePciPool1, &samplePciPool2}
	sampleMetric := NewCellMetric(int32(1), e2smrcpreies.CellSize_CELL_SIZE_FEMTO, int32(1))
	sampleCellPciNrt := CellPciNrt{
		Metric: sampleMetric,
		Neighbors: sampleNeighbors,
		PciPoolList: samplePciPools,
	}

	newCellPciNrt := NewCellPciNrt(sampleMetric, samplePciPools, sampleNeighbors)
	fmt.Printf("sampleCellPciNrt: %v\n", sampleCellPciNrt)
	fmt.Printf("newCellPciNrt: %v\n", newCellPciNrt)
	assert.Equal(t, sampleCellPciNrt, *newCellPciNrt)
}