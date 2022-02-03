// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package northbound

import (
	"context"
	"github.com/onosproject/onos-pci/pkg/utils/parse"

	pciapi "github.com/onosproject/onos-api/go/onos/pci"
	e2sm_rc_pre_v2 "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc_pre/v2/e2sm-rc-pre-v2"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	service "github.com/onosproject/onos-lib-go/pkg/northbound"
	"github.com/onosproject/onos-pci/pkg/store/metrics"

	"github.com/onosproject/onos-pci/pkg/types"
	"google.golang.org/grpc"
)

var log = logging.GetLogger("e2", "subscription", "manager")

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

// NewTestServer returns a server for testing purposes
func NewTestServer(store metrics.Store) *Server {
	return &Server{store: store}
}

type Server struct {
	store metrics.Store
}

// GetNumConflicts returns how many conflicts occur for a given cell or cells in total
func (s *Server) GetConflicts(ctx context.Context, request *pciapi.GetConflictsRequest) (*pciapi.GetConflictsResponse, error) {
	log.Infof("Received PCI Conflicts Request %v", request)
	conflicts := make([]*pciapi.PciCell, 0)
	if request.CellId != 0 {
		cell, err := s.store.Get(ctx, request.CellId)
		if err != nil {
			return nil, err
		}
		pci := cell.Value.Metric.PCI
		for i := range cell.Value.Neighbors {
			neighbor := cell.Value.Neighbors[i]
			if pci == neighbor.Pci.Value {
				temp, _ := s.store.Get(ctx, metrics.NewKey(neighbor.Cgi))
				conflicts = append(conflicts, cellPciToPciCell(temp.Key, temp.Value))
			}
		}
	} else {
		ch := make(chan *metrics.Entry, 1024)
		err := s.store.Entries(ctx, ch)
		if err != nil {
			return nil, err
		}
		// if a cell has a conflict, we only have to append the cell b/c conflicting neighbors will be appended individually
		for cell := range ch {
			pci := cell.Value.Metric.PCI
			for i := range cell.Value.Neighbors {
				if pci == cell.Value.Neighbors[i].Pci.Value {
					conflicts = append(conflicts, cellPciToPciCell(cell.Key, cell.Value))
					break
				}
			}
		}
	}
	return &pciapi.GetConflictsResponse{Cells: conflicts}, nil
}

func (s *Server) GetResolvedConflicts(ctx context.Context, request *pciapi.GetResolvedConflictsRequest) (*pciapi.GetResolvedConflictsResponse, error) {
	conflicts := make([]*pciapi.CellResolution, 0)

	ch := make(chan *metrics.Entry, 1024)
	err := s.store.Entries(ctx, ch)
	if err != nil {
		return nil, err
	}
	for cell := range ch {
		if cell.Value.Metric.ResolvedConflicts != 0 {
			conflicts = append(conflicts, &pciapi.CellResolution{
				Id:                nrcgiToInt(cell.Key.CellGlobalID.GetNrCgi()),
				ResolvedPci:       uint32(cell.Value.Metric.PCI),
				OriginalPci:       uint32(cell.Value.Metric.PreviousPCI),
				ResolvedConflicts: cell.Value.Metric.ResolvedConflicts,
			})
		}
	}
	return &pciapi.GetResolvedConflictsResponse{Cells: conflicts}, nil
}

func (s *Server) GetCell(ctx context.Context, request *pciapi.GetCellRequest) (*pciapi.GetCellResponse, error) {
	log.Infof("Received PCI Cell Request %v", request)
	cell, err := s.store.Get(ctx, request.CellId)
	if err != nil {
		return nil, err
	}
	return &pciapi.GetCellResponse{Cell: cellPciToPciCell(cell.Key, cell.Value)}, nil
}

func (s *Server) GetCells(ctx context.Context, request *pciapi.GetCellsRequest) (*pciapi.GetCellsResponse, error) {
	log.Infof("Received PCI Cells Request %v", request)
	output := make([]*pciapi.PciCell, 0)
	ch := make(chan *metrics.Entry, 1024)
	err := s.store.Entries(ctx, ch)
	if err != nil {
		return nil, err
	}
	for c := range ch {
		output = append(output, cellPciToPciCell(c.Key, c.Value))
	}
	return &pciapi.GetCellsResponse{Cells: output}, nil
}

// convert from NRCGI to uint64
func nrcgiToInt(nrcgi *e2sm_rc_pre_v2.Nrcgi) uint64 {
	array := nrcgi.PLmnIdentity.Value
	plmnid := uint32(array[0])<<0 | uint32(array[1])<<8 | uint32(array[2])<<16
	nci := nrcgi.NRcellIdentity.Value.Value

	return uint64(plmnid)<<36 | parse.BitStringToUint64(nci, int(nrcgi.NRcellIdentity.Value.Len))
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
