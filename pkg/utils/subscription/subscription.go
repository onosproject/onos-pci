// SPDX-FileCopyrightText: 2022-present Intel Corporation
// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package subscription

import (
	e2api "github.com/onosproject/onos-api/go/onos/e2t/e2/v1beta1"
	"github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc/pdubuilder"
	e2smrc "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc/v1/e2sm-rc-ies"
	"google.golang.org/protobuf/proto"
)

const (
	e2nodeInformationChangeID1CellConfigChange      = 1
	e2nodeInformationChangeID2CellNeighborRelChange = 2
)

// CreateEventTriggerDefinition creates RC event trigger data
func CreateEventTriggerDefinition() ([]byte, error) {
	// for cell configuration change action
	e2NodeInformationChangeID1, err := pdubuilder.CreateE2SmRcEventTriggerFormat3Item(e2nodeInformationChangeID1CellConfigChange,
		e2nodeInformationChangeID1CellConfigChange)
	if err != nil {
		return nil, err
	}
	// for cell neighbor relation change
	e2NodeInformationChangeID2, err := pdubuilder.CreateE2SmRcEventTriggerFormat3Item(e2nodeInformationChangeID2CellNeighborRelChange,
		e2nodeInformationChangeID2CellNeighborRelChange)
	if err != nil {
		return nil, err
	}

	itemList := []*e2smrc.E2SmRcEventTriggerFormat3Item{e2NodeInformationChangeID1, e2NodeInformationChangeID2}

	rcEventTriggerDefinitionFormat3, err := pdubuilder.CreateE2SmRcEventTriggerFormat3(itemList)
	if err != nil {
		return nil, err
	}

	err = rcEventTriggerDefinitionFormat3.Validate()
	if err != nil {
		return nil, err
	}

	protoBytes, err := proto.Marshal(rcEventTriggerDefinitionFormat3)
	if err != nil {
		return nil, err
	}

	return protoBytes, nil
}

// CreateSubscriptionActions creates subscription actions for report
func CreateSubscriptionActions() []e2api.Action {
	actions := make([]e2api.Action, 0)
	action := &e2api.Action{
		ID:   int32(3),
		Type: e2api.ActionType_ACTION_TYPE_REPORT,
	}
	actions = append(actions, *action)
	return actions

}
