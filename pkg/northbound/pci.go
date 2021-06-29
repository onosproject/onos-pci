// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package northbound

import (
	"context"

	pciapi "github.com/onosproject/onos-api/go/onos/pci"
	e2sm_rc_pre_v2 "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc_pre/v2/e2sm-rc-pre-v2"
	service "github.com/onosproject/onos-lib-go/pkg/northbound"
	"github.com/onosproject/onos-pci/pkg/store/metrics"

	ransim_types "github.com/onosproject/onos-api/go/onos/ransim/types"
	"github.com/onosproject/onos-pci/pkg/types"
	"google.golang.org/grpc"
)

// NewService returns a new PCI interface service.
func NewService(store metrics.Store) service.Service {
	return &Service{
		store: store,
	}
}

// Service is a service implementation for administration.
type Service struct {
	store metrics.Store
}

// Register registers the Service with the gRPC server.
func (s Service) Register(r *grpc.Server) {
	server := &Server{
		store: s.store,
	}
	pciapi.RegisterPciServer(r, server)
}

type Server struct {
	store metrics.Store
}

// GetNumConflicts returns how many conflicts occur for a given cell or cells in total
func (s *Server) GetConflicts(ctx context.Context, request *pciapi.GetConflictsRequest) (*pciapi.GetConflictsResponse, error) {
	conflicts := make([]*pciapi.PciCell, 0)
	if request.CellId != 0 {
		cell, err := s.store.Get(ctx, metrics.NewKey(&e2sm_rc_pre_v2.CellGlobalId{CellGlobalId: &e2sm_rc_pre_v2.CellGlobalId_NrCgi{NrCgi: intToNRCGI(request.CellId)}}))
		if err != nil {
			return nil, err
		}
		pci := cell.Value.(types.CellPCI).Metric.PCI
		for i := range cell.Value.(types.CellPCI).Neighbors {
			neighbor := cell.Value.(types.CellPCI).Neighbors[i]
			if pci == neighbor.Pci.Value {
				temp, _ := s.store.Get(ctx, metrics.NewKey(neighbor.Cgi))
				conflicts = append(conflicts, cellPciToPciCell(temp.Key, temp.Value.(types.CellPCI)))
			}
		}
	} else {
		ch := make(chan *metrics.Entry)
		err := s.store.Entries(ctx, ch)
		if err != nil {
			return nil, err
		}
		// if a cell has a conflict, we only have to append the cell b/c conflicting neighbors will be appended individually
		for cell := range ch {
			pci := cell.Value.(types.CellPCI).Metric.PCI
			for i := range cell.Value.(types.CellPCI).Neighbors {
				if pci == cell.Value.(types.CellPCI).Neighbors[i].Pci.Value {
					conflicts = append(conflicts, cellPciToPciCell(cell.Key, cell.Value.(types.CellPCI)))
				}
			}
		}
	}
	return &pciapi.GetConflictsResponse{Cells: conflicts}, nil
}

func (s *Server) GetCell(ctx context.Context, request *pciapi.GetCellRequest) (*pciapi.GetCellResponse, error) {
	cell, err := s.store.Get(ctx, metrics.NewKey(&e2sm_rc_pre_v2.CellGlobalId{CellGlobalId: &e2sm_rc_pre_v2.CellGlobalId_NrCgi{NrCgi: intToNRCGI(request.CellId)}}))
	if err != nil {
		return nil, err
	}
	return &pciapi.GetCellResponse{Cell: cellPciToPciCell(cell.Key, cell.Value.(types.CellPCI))}, nil
}

func (s *Server) GetCells(ctx context.Context, request *pciapi.GetCellsRequest) (*pciapi.GetCellsResponse, error) {
	output := make([]*pciapi.PciCell, 0)
	ch := make(chan *metrics.Entry)
	err := s.store.Entries(ctx, ch)
	if err != nil {
		return nil, err
	}
	for c := range ch {
		cell, ok := c.Value.(types.CellPCI)
		if ok {
			output = append(output, cellPciToPciCell(c.Key, cell))
		}
	}
	return &pciapi.GetCellsResponse{Cells: output}, nil
}

// converting from uint64 to NRCGI's
func intToNRCGI(ncgi uint64) *e2sm_rc_pre_v2.Nrcgi {
	plmnid := uint32(ransim_types.GetPlmnID(ncgi))
	bitmask := 0xFF
	nci := uint64(ransim_types.GetNCI(ransim_types.NCGI(ncgi)))

	return &e2sm_rc_pre_v2.Nrcgi{
		PLmnIdentity: &e2sm_rc_pre_v2.PlmnIdentity{
			Value: []byte{byte((plmnid >> 0) & uint32(bitmask)), byte((plmnid >> 8) & uint32(bitmask)), byte((plmnid >> 16) & uint32(bitmask))},
		},
		NRcellIdentity: &e2sm_rc_pre_v2.NrcellIdentity{
			Value: &e2sm_rc_pre_v2.BitString{
				Value: nci,
				Len:   36,
			},
		},
	}
}

// convert from NRCGI to uint64
func nrcgiToInt(nrcgi *e2sm_rc_pre_v2.Nrcgi) uint64 {
	array := nrcgi.PLmnIdentity.Value
	plmnid := uint32(array[0])<<0 | uint32(array[1])<<8 | uint32(array[2])<<16
	nci := nrcgi.NRcellIdentity.Value.Value

	return uint64(plmnid)<<36 | nci
}

// helper function used in cellPciToPciCell
func pciPoolToRange(list []*types.PCIPool) []*pciapi.PciRange {
	out := make([]*pciapi.PciRange, 0)
	for pool := range list {
		out = append(out, &pciapi.PciRange{
			Min: uint32(list[pool].LowerPci),
			Max: uint32(list[pool].UpperPci),
		})
	}
	return out
}

// helper function used in cellPciToPciCell
func neighborsToIDs(list []*e2sm_rc_pre_v2.Nrt) []uint64 {
	out := make([]uint64, 0)
	for pool := range list {
		out = append(out, nrcgiToInt(list[pool].Cgi.GetNrCgi()))
	}
	return out
}

// helper function to convert between onos-api representation and internal store
func cellPciToPciCell(key metrics.Key, cell types.CellPCI) *pciapi.PciCell {
	return &pciapi.PciCell{
		Id:          nrcgiToInt(key.CellGlobalID.GetNrCgi()),
		NodeId:      string(cell.E2NodeID),
		Dlearfcn:    uint32(cell.Metric.DlEARFCN),
		CellType:    pciapi.CellType(cell.Metric.CellSize),
		Pci:         uint32(cell.Metric.PCI),
		PciPool:     pciPoolToRange(cell.PCIPoolList),
		NeighborIds: neighborsToIDs(cell.Neighbors),
	}
}
