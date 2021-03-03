// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package decode

//DecodePlmnIdToUint32 decodes PLMN ID from byte array to uint32
func PlmnIdToUint32(plmnBytes []byte) uint32 {
	return uint32(plmnBytes[0]) | uint32(plmnBytes[1])<<8 | uint32(plmnBytes[2])<<16
}

