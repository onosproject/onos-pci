// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package northbound

import (
	"context"

	pciapi "github.com/onosproject/onos-api/go/onos/pci"
	service "github.com/onosproject/onos-lib-go/pkg/northbound"
	"google.golang.org/grpc"
)

// NewService returns a new PCI interface service.
func NewService() service.Service {
	return &Service{}
}

// Service is a service implementation for administration.
type Service struct {
	service.Service
}

// Register registers the Service with the gRPC server.
func (s Service) Register(r *grpc.Server) {
	server := &Server{}
	pciapi.RegisterPciServer(r, server)

}

type Server struct {
}

// GetNumConflicts returns how many conflicts are happened for a specific cell
func (s Server) GetNumConflicts(ctx context.Context, request *pciapi.GetRequest) (*pciapi.GetResponse, error) {
	panic("implement me")
}

// GetNumConflictsAll returns how many conflicts are happened for all cells
func (s Server) GetNumConflictsAll(ctx context.Context, request *pciapi.GetRequest) (*pciapi.GetResponse, error) {
	// ignore ID here since it will return results for all cells
	panic("implement me")
}

// GetNeighbors returns neighbor cells for a specific cell
func (s Server) GetNeighbors(ctx context.Context, request *pciapi.GetRequest) (*pciapi.GetResponse, error) {
	panic("implement me")
}

// GetNeighborsAll returns neighbor cells for all cells
func (s Server) GetNeighborsAll(ctx context.Context, request *pciapi.GetRequest) (*pciapi.GetResponse, error) {
	panic("implement me")
}

// GetMetric returns metrics for a specific cell
func (s Server) GetMetric(ctx context.Context, request *pciapi.GetRequest) (*pciapi.GetResponse, error) {
	panic("implement me")
}

// GetMetricAll returns metrics for all cells
func (s Server) GetMetricAll(ctx context.Context, request *pciapi.GetRequest) (*pciapi.GetResponse, error) {
	panic("implement me")
}

// GetPci returns PCI for a specific cell
func (s Server) GetPci(ctx context.Context, request *pciapi.GetRequest) (*pciapi.GetResponse, error) {
	panic("implement me")
}

// GetPciAll returns PCIs for all cells
func (s Server) GetPciAll(ctx context.Context, request *pciapi.GetRequest) (*pciapi.GetResponse, error) {
	panic("implement me")
}
