// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package metrics

// Event store event data structure
type Event struct {
	Key   uint64
	Value Entry
	Type  interface{}
}
