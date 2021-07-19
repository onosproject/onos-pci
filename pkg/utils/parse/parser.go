// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package parse

import (
	e2smrcprev2 "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc_pre/v2/e2sm-rc-pre-v2"
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

func ParseMetricKey(e *e2smrcprev2.CellGlobalId) ([]byte, uint64, CGIType, error) {
	if e == nil {
		return nil, 0, CGITypeUnknown, errors.NewNotFound("CellGlobalID is not found in entry Key field")
	} else if e.GetNrCgi() != nil {
		return e.GetNrCgi().GetPLmnIdentity().GetValue(),
			BitStringToUint64(e.GetNrCgi().GetNRcellIdentity().GetValue().Value, int(e.GetNrCgi().GetNRcellIdentity().GetValue().Len)),
			CGITypeNrCGI,
			nil
	} else if e.GetEUtraCgi() != nil {
		return e.GetEUtraCgi().GetPLmnIdentity().GetValue(),
			BitStringToUint64(e.GetEUtraCgi().GetEUtracellIdentity().GetValue().Value, int(e.GetEUtraCgi().GetEUtracellIdentity().GetValue().Len)),
			CGITypeECGI,
			nil
	}
	return nil, 0, CGITypeUnknown, errors.NewNotSupported("CGI should be one of NrCGI and ECGI")
}

func GetCellID(cellGlobalID *e2smrcprev2.CellGlobalId) (uint64, error) {
	switch v := cellGlobalID.GetCellGlobalId().(type) {
	case *e2smrcprev2.CellGlobalId_EUtraCgi:
		return BitStringToUint64(v.EUtraCgi.EUtracellIdentity.Value.Value, int(v.EUtraCgi.EUtracellIdentity.Value.Len)), nil
	case *e2smrcprev2.CellGlobalId_NrCgi:
		return BitStringToUint64(v.NrCgi.NRcellIdentity.Value.Value, int(v.NrCgi.NRcellIdentity.Value.Len)), nil
	}
	return 0, errors.New(errors.NotSupported, "CGI should be one of NrCGI and ECGI")
}

func BitStringToUint64(bitString []byte, bitCount int) uint64 {
	unusedBits := 8 - bitCount%8
	var result uint64 = 0
	for i, b := range bitString {
		result += (uint64(b) << ((len(bitString)-i-1) * 8))

	}

	return result >> unusedBits
}