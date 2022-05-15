// SPDX-FileCopyrightText: 2022-present Intel Corporation
// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package northbound

import (
	"context"

	"github.com/onosproject/onos-pci/pkg/utils/parse"

	pciapi "github.com/onosproject/onos-api/go/onos/pci"
	e2smrccomm "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc/v1/e2sm-common-ies"
	e2smrc "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc/v1/e2sm-rc-ies"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	service "github.com/onosproject/onos-lib-go/pkg/northbound"
	"github.com/onosproject/onos-pci/pkg/store/metrics"

	"github.com/onosproject/onos-pci/pkg/types"
	"google.golang.org/grpc"
)

var log = logging.GetLogger()

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

// GetConflicts returns how many conflicts occur for a given cell or cells in total
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
			if neighbor.GetRanTypeChoiceNr() != nil {
				// 5G case
				if pci == neighbor.GetRanTypeChoiceNr().GetNRPci().GetValue() {
					temp, _ := s.store.Get(ctx, metrics.NewKey(&e2smrccomm.Cgi{
						Cgi: &e2smrccomm.Cgi_NRCgi{
							NRCgi: neighbor.GetRanTypeChoiceNr().GetNRCgi(),
						},
					}))
					conflicts = append(conflicts, cellPciToPciCell(temp.Key, temp.Value))
				}
			} else if neighbor.GetRanTypeChoiceEutra() != nil {
				// 4G case
				// ToDo: Add 4G case
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
				if cell.Value.Neighbors[i].GetRanTypeChoiceNr() != nil {
					// 5G case
					if pci == cell.Value.Neighbors[i].GetRanTypeChoiceNr().GetNRPci().GetValue() {
						conflicts = append(conflicts, cellPciToPciCell(cell.Key, cell.Value))
						break
					}
				} else if cell.Value.Neighbors[i].GetRanTypeChoiceEutra() != nil {
					// 4G case
					// ToDo: Add 4G case
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
				Id:                cgiToInt(cell.Key.CellGlobalID),
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
func cgiToInt(cgi *e2smrccomm.Cgi) uint64 {
	if cgi.GetNRCgi() != nil {
		// 5G case
		array := cgi.GetNRCgi().GetPLmnidentity().GetValue()
		plmnID := uint32(array[0])<<0 | uint32(array[1])<<8 | uint32(array[2])<<16
		return uint64(plmnID)<<36 | parse.BitStringToUint64(cgi.GetNRCgi().GetNRcellIdentity().GetValue().GetValue(), int(cgi.GetNRCgi().GetNRcellIdentity().GetValue().GetLen()))
	} else if cgi.GetEUtraCgi() != nil {
		// 4G case
		// ToDo: Add 4G case
	}
	log.Errorf("CGI does not have EUTRA CGI or NR CGI")
	return 0
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
func neighborsToIDs(list []*e2smrc.NeighborCellItem) []uint64 {
	out := make([]uint64, 0)
	for pool := range list {

		if list[pool].GetRanTypeChoiceNr() != nil {
			// 5G case
			out = append(out, cgiToInt(&e2smrccomm.Cgi{
				Cgi: &e2smrccomm.Cgi_NRCgi{
					NRCgi: list[pool].GetRanTypeChoiceNr().GetNRCgi(),
				},
			}))
		} else if list[pool].GetRanTypeChoiceEutra() != nil {
			// 4G case
			// ToDo: Add 4G case
		}
	}
	return out
}

// helper function to convert between onos-api representation and internal store
func cellPciToPciCell(key metrics.Key, cell types.CellPCI) *pciapi.PciCell {
	// 5G case
	// 4G case
	return &pciapi.PciCell{
		Id:          cgiToInt(key.CellGlobalID),
		NodeId:      string(cell.E2NodeID),
		Arfcn:       uint32(cell.Metric.ARFCN),
		Pci:         uint32(cell.Metric.PCI),
		PciPool:     pciPoolToRange(cell.PCIPoolList),
		NeighborIds: neighborsToIDs(cell.Neighbors),
	}
}
