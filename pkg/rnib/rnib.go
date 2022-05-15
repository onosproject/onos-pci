// SPDX-FileCopyrightText: 2022-present Intel Corporation
// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package rnib

import (
	"context"
	"fmt"

	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	e2smrc "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc/v1/e2sm-rc-ies"
	"github.com/onosproject/onos-pci/pkg/utils/decode"
	"github.com/onosproject/onos-pci/pkg/utils/parse"
	toposdk "github.com/onosproject/onos-ric-sdk-go/pkg/topo"
)

// TopoClient R-NIB client interface
type TopoClient interface {
	WatchE2Connections(ctx context.Context, ch chan topoapi.Event) error
	GetE2NodeAspects(ctx context.Context, nodeID topoapi.ID) (*topoapi.E2Node, error)
	GetCells(ctx context.Context, nodeID topoapi.ID) ([]*topoapi.E2Cell, error)
	E2NodeIDs(ctx context.Context) ([]topoapi.ID, error)
}

// NewClient creates a new topo SDK client
func NewClient() (Client, error) {
	sdkClient, err := toposdk.NewClient()
	if err != nil {
		return Client{}, err
	}
	cl := Client{
		client: sdkClient,
	}

	return cl, nil

}

// Client topo SDK client
type Client struct {
	client toposdk.Client
}

func (c *Client) UpdateCellAspects(ctx context.Context, cellID topoapi.ID, pci uint32, neighborIDs []*e2smrc.NeighborCellItem, arfcn uint32) error {
	object, err := c.client.Get(ctx, cellID)
	if err != nil {
		return err
	}

	if object != nil && object.GetEntity().GetKindID() == topoapi.E2CELL {
		cellObject := &topoapi.E2Cell{}
		err := object.GetAspect(cellObject)
		if err != nil {
			return err
		}
		cellObject.PCI = pci
		cellObject.NeighborCellIDs = make([]*topoapi.NeighborCellID, 0)
		for _, nID := range neighborIDs {
			switch v := nID.NeighborCellItem.(type) {
			case *e2smrc.NeighborCellItem_RanTypeChoiceNr:
				// 5G case
				nPlmnIDByte, nCid, _, err := parse.GetNRMetricKey(v.RanTypeChoiceNr.NRCgi)
				nPlmnID := decode.PlmnIDToUint32(nPlmnIDByte)
				nIDObj := &topoapi.NeighborCellID{
					CellGlobalID: &topoapi.CellGlobalID{
						Value: fmt.Sprintf("%x", nCid),
					},
					PlmnID: fmt.Sprintf("%x", nPlmnID),
				}
				cellObject.NeighborCellIDs = append(cellObject.NeighborCellIDs, nIDObj)
				if err != nil {
					return err
				}
			case *e2smrc.NeighborCellItem_RanTypeChoiceEutra:
				// 4G case
				// ToDo: add EUTRA case
				fmt.Println("4G case is not supported yet")
			}
		}
		cellObject.ARFCN = arfcn
		err = object.SetAspect(cellObject)
		if err != nil {
			return err
		}
		err = c.client.Update(ctx, object)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetCells get list of cells for each E2 node
func (c *Client) GetCells(ctx context.Context, nodeID topoapi.ID) ([]*topoapi.E2Cell, error) {
	filter := &topoapi.Filters{
		RelationFilter: &topoapi.RelationFilter{SrcId: string(nodeID),
			RelationKind: topoapi.CONTAINS,
			TargetKind:   ""}}

	objects, err := c.client.List(ctx, toposdk.WithListFilters(filter))
	if err != nil {
		return nil, err
	}
	var cells []*topoapi.E2Cell
	for _, obj := range objects {
		targetEntity := obj.GetEntity()
		if targetEntity.GetKindID() == topoapi.E2CELL {
			cellObject := &topoapi.E2Cell{}
			err = obj.GetAspect(cellObject)
			if err != nil {
				return nil, err
			}
			cells = append(cells, cellObject)
		}
	}
	return cells, nil
}

// E2NodeIDs lists all of connected E2 nodes
func (c *Client) E2NodeIDs(ctx context.Context) ([]topoapi.ID, error) {
	objects, err := c.client.List(ctx, toposdk.WithListFilters(getControlRelationFilter()))
	if err != nil {
		return nil, err
	}

	e2NodeIDs := make([]topoapi.ID, len(objects))
	for _, object := range objects {
		relation := object.Obj.(*topoapi.Object_Relation)
		e2NodeID := relation.Relation.TgtEntityID
		e2NodeIDs = append(e2NodeIDs, e2NodeID)
	}

	return e2NodeIDs, nil
}

// GetE2NodeAspects gets E2 node aspects
func (c *Client) GetE2NodeAspects(ctx context.Context, nodeID topoapi.ID) (*topoapi.E2Node, error) {
	object, err := c.client.Get(ctx, nodeID)
	if err != nil {
		return nil, err
	}
	e2Node := &topoapi.E2Node{}
	err = object.GetAspect(e2Node)
	return e2Node, err

}

func getControlRelationFilter() *topoapi.Filters {
	controlRelationFilter := &topoapi.Filters{
		KindFilter: &topoapi.Filter{
			Filter: &topoapi.Filter_Equal_{
				Equal_: &topoapi.EqualFilter{
					Value: topoapi.CONTROLS,
				},
			},
		},
	}
	return controlRelationFilter
}

// WatchE2Connections watch e2 node connection changes
func (c *Client) WatchE2Connections(ctx context.Context, ch chan topoapi.Event) error {
	err := c.client.Watch(ctx, ch, toposdk.WithWatchFilters(getControlRelationFilter()))
	if err != nil {
		return err
	}
	return nil
}

var _ TopoClient = &Client{}
