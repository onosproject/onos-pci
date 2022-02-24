// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package control

import (
	"github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc_pre_go/pdubuilder"
	e2smrcpre "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc_pre_go/v2/e2sm-rc-pre-v2-go"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func CreateRcControlHeader(cgi *e2smrcpre.CellGlobalId, priority *int32) ([]byte, error) {
	if cgi == nil {
		return nil, errors.NewInvalid("cgi is not set")
	}

	newE2SmRcPrePdu, err := pdubuilder.CreateE2SmRcPreControlHeader()
	ctrlHdrFormat1 := newE2SmRcPrePdu.GetControlHeaderFormat1()
	ctrlHdrFormat1.SetCGI(cgi)
	if priority != nil {
		ctrlHdrFormat1.SetRicControlMessagePriority(*priority)
	}

	if err != nil {
		return []byte{}, err
	}

	err = newE2SmRcPrePdu.Validate()
	if err != nil {
		return []byte{}, err
	}

	protoBytes, err := proto.Marshal(newE2SmRcPrePdu)
	if err != nil {
		return []byte{}, err
	}

	return protoBytes, nil
}

func CreateRcControlMessage(ranParamID int32, ranParamName string, ranParamValue int64) ([]byte, error) {
	ranParamValueInt, err := pdubuilder.CreateRanParameterValueInt(ranParamValue)
	if err != nil {
		return []byte{}, err
	}
	newE2SmRcPrePdu, err := pdubuilder.CreateE2SmRcPreControlMessage(ranParamID, ranParamName, ranParamValueInt)
	if err != nil {
		return []byte{}, err
	}

	err = newE2SmRcPrePdu.Validate()
	if err != nil {
		return []byte{}, err
	}

	protoBytes, err := proto.Marshal(newE2SmRcPrePdu)
	if err != nil {
		return []byte{}, err
	}

	return protoBytes, nil
}
