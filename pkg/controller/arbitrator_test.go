// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package controller

import (
	"fmt"
	e2smrcpreies "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc_pre/v1/e2sm-rc-pre-ies"
	"github.com/onosproject/onos-pci/pkg/store"
	"github.com/onosproject/onos-pci/pkg/utils/decode"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPciArbitratorController(t *testing.T) {

	sampleTargetE2NodeCgi, sampleTargetE2NodeMetric, samplePciArbitratorController, _, _ := GenSamplePciArbitratorController()
	newPciArbitratorController := NewPciArbitratorController(sampleTargetE2NodeCgi, sampleTargetE2NodeMetric)
	fmt.Printf("samplePciArbitratorController: %v\n", samplePciArbitratorController)
	fmt.Printf("newPciArbitratorController: %v\n", newPciArbitratorController)

	assert.Equal(t, *samplePciArbitratorController, *newPciArbitratorController)
}

func GenSamplePciArbitratorController() (*store.CGI, *store.CellPciNrt ,*PciArbitratorCtrl, map[string]int32, map[string]*store.CellPciNrt) {
	sampleTargetE2NodeCgi := store.NewCGI(uint32(1279014), uint64(82530), uint32(28))
	sampleNeighborCell1 := store.NeighborCell{
		Metric: store.NewCellMetric(int32(1), e2smrcpreies.CellSize_CELL_SIZE_FEMTO, int32(1)),
		Cgi: store.NewCGI(uint32(1279015), uint64(82530), uint32(28)),
		NrIndex: 0,
	}
	sampleNeighborCell2 := store.NeighborCell{
		Metric: store.NewCellMetric(int32(1), e2smrcpreies.CellSize_CELL_SIZE_FEMTO, int32(2)),
		Cgi: store.NewCGI(uint32(1279016), uint64(82530), uint32(28)),
		NrIndex: 1,
	}
	sampleNeighborCell3 := store.NeighborCell{
		Metric: store.NewCellMetric(int32(1), e2smrcpreies.CellSize_CELL_SIZE_FEMTO, int32(3)),
		Cgi: store.NewCGI(uint32(1279017), uint64(82530), uint32(28)),
		NrIndex: 2,
	}
	samplePciPool1 := store.PciPool{
		LowerPci: int32(1),
		UpperPci: int32(10),
	}
	samplePciPool2 := store.PciPool{
		LowerPci: int32(21),
		UpperPci: int32(30),
	}
	sampleNeighbors := []*store.NeighborCell{&sampleNeighborCell1, &sampleNeighborCell2, &sampleNeighborCell3}
	samplePciPools := []*store.PciPool{&samplePciPool1, &samplePciPool2}
	sampleMetric := store.NewCellMetric(int32(1), e2smrcpreies.CellSize_CELL_SIZE_FEMTO, int32(1))
	sampleTargetE2NodeMetric := store.NewCellPciNrt(sampleMetric, samplePciPools, sampleNeighbors)
	sampleGlobalPciMap := make(map[string]int32)
	sampleGlobalPciMap[decode.CgiToString(sampleTargetE2NodeCgi)] = 1
	sampleGlobalPciMap[decode.CgiToString(store.NewCGI(uint32(1279015), uint64(82530), uint32(28)))] = 1
	sampleGlobalPciMap[decode.CgiToString(store.NewCGI(uint32(1279016), uint64(82530), uint32(28)))] = 2
	sampleGlobalPciMap[decode.CgiToString(store.NewCGI(uint32(1279017), uint64(82530), uint32(28)))] = 3
	sampleCellPciNrtMap := make(map[string]*store.CellPciNrt)
	sampleCellPciNrtMap[decode.CgiToString(sampleTargetE2NodeCgi)] = sampleTargetE2NodeMetric

	fmt.Printf("sampleTargetE2NodeCgi: %v\n", sampleTargetE2NodeCgi)
	fmt.Printf("sampleNeighborCell1: %v\n", sampleNeighborCell1)
	fmt.Printf("sampleNeighborCell2: %v\n", sampleNeighborCell2)
	fmt.Printf("sampleNeighborCell3: %v\n", sampleNeighborCell3)
	fmt.Printf("samplePciPool1: %v\n", samplePciPool1)
	fmt.Printf("samplePciPool2: %v\n", samplePciPool2)
	fmt.Printf("sampleMetric: %v\n", sampleMetric)

	return sampleTargetE2NodeCgi, sampleTargetE2NodeMetric, &PciArbitratorCtrl{
		TargetE2NodeCgi: sampleTargetE2NodeCgi,
		TargetE2NodeMetric: sampleTargetE2NodeMetric,
		D1NeighborPciMap:   make(map[string]int32),
		D2NeighborPciMap:   make(map[string]int32),
		NeighborPcis:       make(map[int32]bool),
	}, sampleGlobalPciMap, sampleCellPciNrtMap
}

func TestGetUniquePci(t *testing.T) {
	_, _, samplePciArbitratorController, sampleGlobalPciMap, sampleCellPciNrtMap := GenSamplePciArbitratorController()
	samplePciArbitratorController.setD1NeighborPciMap(sampleCellPciNrtMap, sampleGlobalPciMap)
	samplePciArbitratorController.setD2NeighborPciMap(sampleCellPciNrtMap, sampleGlobalPciMap)
	pci, _ := samplePciArbitratorController.getUniquePci()

	fmt.Printf("assigned pci: %v\n", pci)
	assert.NotEqual(t, 1, pci)
	assert.NotEqual(t, 2, pci)
	assert.NotEqual(t, 3, pci)
	assert.Condition(t, func() bool {
		for i := 0; i < len(samplePciArbitratorController.TargetE2NodeMetric.PciPoolList); i++ {
			if pci >= samplePciArbitratorController.TargetE2NodeMetric.PciPoolList[i].LowerPci ||
				pci <= samplePciArbitratorController.TargetE2NodeMetric.PciPoolList[i].UpperPci {
				return true
			}
		}
		return false
	})
}

func TestVerifyPci(t *testing.T) {
	_, _, samplePciArbitratorController, sampleGlobalPciMap, sampleCellPciNrtMap := GenSamplePciArbitratorController()
	samplePciArbitratorController.setD1NeighborPciMap(sampleCellPciNrtMap, sampleGlobalPciMap)
	samplePciArbitratorController.setD2NeighborPciMap(sampleCellPciNrtMap, sampleGlobalPciMap)
	verifyResult := samplePciArbitratorController.verifyPci()
	fmt.Printf("Verification results: %v\n", verifyResult)
	assert.False(t, verifyResult)
}

func TestArbitratePCI(t *testing.T) {
	_, _, samplePciArbitratorController, sampleGlobalPciMap, sampleCellPciNrtMap := GenSamplePciArbitratorController()
	changed, _ := samplePciArbitratorController.ArbitratePCI(sampleCellPciNrtMap, sampleGlobalPciMap)
	fmt.Printf("Is PCI changed: %v\n", changed)
	assert.True(t, changed)
	fmt.Printf("PCI: %v\n", samplePciArbitratorController.TargetE2NodeMetric.Metric.Pci)
	assert.NotEqual(t, 1, samplePciArbitratorController.TargetE2NodeMetric.Metric.Pci)
	assert.NotEqual(t, 2, samplePciArbitratorController.TargetE2NodeMetric.Metric.Pci)
	assert.NotEqual(t, 3, samplePciArbitratorController.TargetE2NodeMetric.Metric.Pci)
	assert.Condition(t, func() bool {
		for i := 0; i < len(samplePciArbitratorController.TargetE2NodeMetric.PciPoolList); i++ {
			if samplePciArbitratorController.TargetE2NodeMetric.Metric.Pci >= samplePciArbitratorController.TargetE2NodeMetric.PciPoolList[i].LowerPci ||
				samplePciArbitratorController.TargetE2NodeMetric.Metric.Pci <= samplePciArbitratorController.TargetE2NodeMetric.PciPoolList[i].UpperPci {
				return true
			}
		}
		return false
	})
}