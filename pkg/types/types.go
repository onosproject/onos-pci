// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package types

import (
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	e2smrcpreies "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc_pre/v2/e2sm-rc-pre-v2"
)

const (
	LowerPCI = 1
	UpperPCI = 512
)

// PCIPool is the PCI pool to be able to assign a PCI to a cell
type PCIPool struct {
	LowerPci int32
	UpperPci int32
}

// CellMetric is the metric struct which has EARFCN-DL, size, and PCI of a cell
type CellMetric struct {
	DlEARFCN          int32
	CellSize          e2smrcpreies.CellSize
	PCI               int32
	PreviousPCI       int32
	ResolvedConflicts uint32
}

// CellPCI is the PCI-NRT information
type CellPCI struct {
	E2NodeID    topoapi.ID
	Metric      *CellMetric
	PCIPoolList []*PCIPool
	Neighbors   []*e2smrcpreies.Nrt
}
