// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package northbound

import (
	"context"
	"fmt"
	service "github.com/onosproject/onos-lib-go/pkg/northbound"
	pciapi "github.com/onosproject/onos-api/go/onos/pci"
	"github.com/onosproject/onos-pci/pkg/controller"
	"github.com/onosproject/onos-pci/pkg/utils/decode"
	"google.golang.org/grpc"
)

// NewService returns a new PCI interface service.
func NewService(ctrl *controller.PciCtrl) service.Service {
	return &Service{
		Ctrl: ctrl,
	}
}

// Service is a service implementation for administration.
type Service struct {
	service.Service
	Ctrl *controller.PciCtrl
}

// Register registers the Service with the gRPC server.
func (s Service) Register(r *grpc.Server) {
	server := &Server{
		Ctrl: s.Ctrl,
	}
	pciapi.RegisterPciServer(r, server)

}

type Server struct {
	Ctrl *controller.PciCtrl
}

// GetNumConflicts returns how many conflicts are happened for a specific cell
func (s Server) GetNumConflicts(ctx context.Context, request *pciapi.GetRequest) (*pciapi.GetResponse, error) {
	id := request.GetId()
	s.Ctrl.PciMonitorMutex.RLock()
	numConflicts := s.Ctrl.PciMonitor[id].NumConflicts
	s.Ctrl.PciMonitorMutex.RUnlock()

	attr := make(map[string]string)
	attr[id] = fmt.Sprintf("%d", numConflicts)

	response := &pciapi.GetResponse{
		Object: &pciapi.Object{
			Id: id,
			Revision: 0,
			Attributes: attr,
		},
	}

	return response, nil
}

// GetNumConflictsAll returns how many conflicts are happened for all cells
func (s Server) GetNumConflictsAll(ctx context.Context, request *pciapi.GetRequest) (*pciapi.GetResponse, error) {
	// ignore ID here since it will return results for all cells
	attr := make(map[string]string)
	s.Ctrl.PciMonitorMutex.RLock()
	for k, v := range s.Ctrl.PciMonitor {
		attr[k] = fmt.Sprintf("%d", v.NumConflicts)
	}
	s.Ctrl.PciMonitorMutex.RUnlock()

	response := &pciapi.GetResponse{
		Object: &pciapi.Object{
			Id: "all",
			Revision: 0,
			Attributes: attr,
		},
	}

	return response, nil
}

// GetNeighbors returns neighbor cells for a specific cell
func (s Server) GetNeighbors(ctx context.Context, request *pciapi.GetRequest) (*pciapi.GetResponse, error) {
	id := request.GetId()
	attr := make(map[string]string)
	s.Ctrl.PciMetricMapMutex.RLock()
	for k, v := range s.Ctrl.PciMetricMap[id].Neighbors {
		attr[fmt.Sprintf("%s:%s", id, fmt.Sprintf("%d", k))] = decode.CgiToString(v.Cgi)
	}
	s.Ctrl.PciMetricMapMutex.RUnlock()

	response := &pciapi.GetResponse{
		Object: &pciapi.Object{
			Id: id,
			Revision: 0,
			Attributes: attr,
		},
	}

	return response, nil
}

// GetNeighborsAll returns neighbor cells for all cells
func (s Server) GetNeighborsAll(ctx context.Context, request *pciapi.GetRequest) (*pciapi.GetResponse, error) {
	// ignore ID here since it will return results for all cells
	attr := make(map[string]string)
	s.Ctrl.PciMetricMapMutex.RLock()
	for k1 := range s.Ctrl.PciMetricMap {
		for k2, v2 := range s.Ctrl.PciMetricMap[k1].Neighbors {
			attr[fmt.Sprintf("%s:%s", k1, fmt.Sprintf("%d", k2))] = decode.CgiToString(v2.Cgi)
		}
	}
	s.Ctrl.PciMetricMapMutex.RUnlock()

	response := &pciapi.GetResponse{
		Object: &pciapi.Object{
			Id: "all",
			Revision: 0,
			Attributes: attr,
		},
	}

	return response, nil
}

// GetMetric returns metrics for a specific cell
func (s Server) GetMetric(ctx context.Context, request *pciapi.GetRequest) (*pciapi.GetResponse, error) {
	id := request.GetId()
	attr := make(map[string]string)
	s.Ctrl.PciMetricMapMutex.RLock()
	attr[fmt.Sprintf("%s:PCI", id)] = fmt.Sprintf("%d", s.Ctrl.PciMetricMap[id].Metric.Pci)
	attr[fmt.Sprintf("%s:DlEarfcn", id)] = fmt.Sprintf("%d", s.Ctrl.PciMetricMap[id].Metric.DlEarfcn)
	attr[fmt.Sprintf("%s:CellSize", id)] = fmt.Sprintf("%d", s.Ctrl.PciMetricMap[id].Metric.CellSize)
	s.Ctrl.PciMetricMapMutex.RUnlock()

	response := &pciapi.GetResponse{
		Object: &pciapi.Object{
			Id: id,
			Revision: 0,
			Attributes: attr,
		},
	}

	return response, nil
}

// GetMetricAll returns metrics for all cells
func (s Server) GetMetricAll(ctx context.Context, request *pciapi.GetRequest) (*pciapi.GetResponse, error) {
	// ignore ID here since it will return results for all cells
	attr := make(map[string]string)
	s.Ctrl.PciMetricMapMutex.RLock()
	for k, v := range s.Ctrl.PciMetricMap {
		attr[fmt.Sprintf("%s:PCI", k)] = fmt.Sprintf("%d", v.Metric.Pci)
		attr[fmt.Sprintf("%s:DlEarfcn", k)] = fmt.Sprintf("%d", v.Metric.DlEarfcn)
		attr[fmt.Sprintf("%s:CellSize", k)] = fmt.Sprintf("%d", v.Metric.CellSize)
	}
	s.Ctrl.PciMetricMapMutex.RUnlock()

	response := &pciapi.GetResponse{
		Object: &pciapi.Object{
			Id: "all",
			Revision: 0,
			Attributes: attr,
		},
	}

	return response, nil
}

// GetPci returns PCI for a specific cell
func (s Server) GetPci(ctx context.Context, request *pciapi.GetRequest) (*pciapi.GetResponse, error) {
	id := request.GetId()
	attr := make(map[string]string)
	s.Ctrl.PciMetricMapMutex.RLock()
	pci := s.Ctrl.GlobalPciMap[id]
	s.Ctrl.PciMetricMapMutex.RUnlock()

	attr[id] = fmt.Sprintf("%d", pci)
	response := &pciapi.GetResponse{
		Object: &pciapi.Object{
			Id: id,
			Revision: 0,
			Attributes: attr,
		},
	}

	return response, nil
}

// GetPciAll returns PCIs for all cells
func (s Server) GetPciAll(ctx context.Context, request *pciapi.GetRequest) (*pciapi.GetResponse, error) {
	// ignore ID here since it will return results for all cells
	attr := make(map[string]string)
	s.Ctrl.PciMetricMapMutex.RLock()
	for k, v := range s.Ctrl.GlobalPciMap {
		attr[k] = fmt.Sprintf("%d", v)
	}
	s.Ctrl.PciMetricMapMutex.RUnlock()

	response := &pciapi.GetResponse{
		Object: &pciapi.Object{
			Id: "all",
			Revision: 0,
			Attributes: attr,
		},
	}

	return response, nil
}
