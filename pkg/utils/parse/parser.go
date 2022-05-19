// SPDX-FileCopyrightText: 2022-present Intel Corporation
// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package parse

import (
	e2smrccomm "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc/v1/e2sm-common-ies"
	"github.com/onosproject/onos-lib-go/pkg/errors"
)

type CGIType int

const (
	CGITypeNrCGI CGIType = iota
	CGITypeECGI
	CGITypeUnknown
)

func (c CGIType) String() string {
	return [...]string{"CGITypeNRCGI", "CGITypeECGI", "CGITypeUnknown"}[c]
}

func GetNRMetricKey(e *e2smrccomm.NrCgi) ([]byte, uint64, CGIType, error) {
	if e == nil {
		return nil, 0, CGITypeUnknown, errors.NewNotFound("CellGlobalID is not found in entry Key field")
	}
	return e.GetPLmnidentity().GetValue(),
		BitStringToUint64(e.GetNRcellIdentity().GetValue().GetValue(), int(e.GetNRcellIdentity().GetValue().GetLen())),
		CGITypeNrCGI,
		nil
}

func GetEUTRAMetricKey(e *e2smrccomm.EutraCgi) ([]byte, uint64, CGIType, error) {
	if e == nil {
		return nil, 0, CGITypeUnknown, errors.NewNotFound("CellGlobalID is not found in entry Key field")
	}
	return e.GetPLmnidentity().GetValue(),
		BitStringToUint64(e.GetEUtracellIdentity().GetValue().GetValue(), int(e.GetEUtracellIdentity().GetValue().GetLen())),
		CGITypeECGI,
		nil
}

func GetCellID(cellGlobalID *e2smrccomm.Cgi) (uint64, error) {
	switch v := cellGlobalID.Cgi.(type) {
	case *e2smrccomm.Cgi_EUtraCgi:
		return BitStringToUint64(v.EUtraCgi.GetEUtracellIdentity().GetValue().GetValue(), int(v.EUtraCgi.GetEUtracellIdentity().GetValue().GetLen())), nil
	case *e2smrccomm.Cgi_NRCgi:
		return BitStringToUint64(v.NRCgi.GetNRcellIdentity().GetValue().GetValue(), int(v.NRCgi.GetNRcellIdentity().GetValue().GetLen())), nil
	}
	return 0, errors.New(errors.NotSupported, "CGI should be one of NrCGI and ECGI")
}

func BitStringToUint64(bitString []byte, bitCount int) uint64 {
	var result uint64
	for i, b := range bitString {
		result += uint64(b) << ((len(bitString) - i - 1) * 8)
	}
	if bitCount%8 != 0 {
		return result >> (8 - bitCount%8)
	}
	return result
}
