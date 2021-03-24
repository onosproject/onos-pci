// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package utils

import (
	"strconv"
)

const (
	SubscriptionServiceHost = "onos-e2sub"
	SubscriptionServicePort = 5150
	RcServiceModelID        = "e2sm_rc_pre-v1"
	RansimServicePort       = 5150
)

var (
	SubscriptionServiceAddress = SubscriptionServiceHost + ":" + strconv.Itoa(SubscriptionServicePort)
)
