// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package control

import (
	"github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc_pre/pdubuilder"
	e2smrcpre "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc_pre/v2/e2sm-rc-pre-v2"
	"google.golang.org/protobuf/proto"
)

func CreateRcControlHeader(cgi *e2smrcpre.CellGlobalId, priority *int32) ([]byte, error) {

	newE2SmRcPrePdu, err := pdubuilder.CreateE2SmRcPreControlHeader(priority, cgi)

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

func CreateRcControlMessage(ranParamID int32, ranParamName string, ranParamValue uint32) ([]byte, error) {
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
