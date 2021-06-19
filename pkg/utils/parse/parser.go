// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package parse

import (
	e2sm_rc_pre_v2 "github.com/onosproject/onos-e2-sm/servicemodels/e2sm_rc_pre/v2/e2sm-rc-pre-v2"
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

func ParseMetricKey(e *e2sm_rc_pre_v2.CellGlobalId) ([]byte, uint64, CGIType, error) {
	if e == nil {
		return nil, 0, CGITypeUnknown, errors.NewNotFound("CellGlobalID is not found in entry Key field")
	} else if e.GetNrCgi() != nil {
		return e.GetNrCgi().GetPLmnIdentity().GetValue(),
		e.GetNrCgi().GetNRcellIdentity().GetValue().Value,
		CGITypeNrCGI,
		nil
	} else if e.GetEUtraCgi() != nil {
		return e.GetEUtraCgi().GetPLmnIdentity().GetValue(),
		e.GetEUtraCgi().GetEUtracellIdentity().GetValue().Value,
		CGITypeECGI,
		nil
	}
	return nil, 0, CGITypeUnknown, errors.NewNotSupported("CGI should be one of NrCGI and ECGI")
}