// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-Only-1.0

package pci

import (
	"github.com/onosproject/onos-lib-go/pkg/certs"
	"github.com/onosproject/onos-pci/pkg/manager"
	"github.com/onosproject/onos-pci/test/utils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func (s *TestSuite) TestThreeCellPci(t *testing.T) {
	cfg := manager.Config{
		CAPath:      "/tmp/tls.cacrt",
		KeyPath:     "/tmp/tls.key",
		CertPath:    "/tmp/tls.crt",
		ConfigPath:  "/tmp/config.json",
		E2tEndpoint: "onos-e2t:5150",   // TODO: Deprecated; remove
		GRPCPort:    5150,
		SMName:      "oran-e2sm-rc-pre",
		SMVersion:   "v2",
	}

	_, err := certs.HandleCertPaths(cfg.CAPath, cfg.KeyPath, cfg.CertPath, true)
	assert.NoError(t, err)

	mgr := manager.NewManager(cfg)
	mgr.Run()

	err = utils.WaitForNoConflicts(t, mgr)
	assert.NoError(t, err)
}
