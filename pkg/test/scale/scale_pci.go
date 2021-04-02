// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package scale

import (
	modelapi "github.com/onosproject/onos-api/go/onos/ransim/model"
	"github.com/onosproject/onos-pci/pkg/manager"
	"github.com/onosproject/onos-pci/pkg/test/utils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func (s *TestSuite) TestScalePci(t *testing.T) {

	e2tEndpoint := "onos-e2t:5150"
	e2subEndpoint := "onos-e2sub:5150"

	cfg := manager.Config{
		E2tEndpoint:   e2tEndpoint,
		E2SubEndpoint: e2subEndpoint,
		GRPCPort:      5150,
		RicActionID:   int32(10),
	}

	time.Sleep(10 * time.Second)
	ready := make(chan bool)
	mgr := manager.NewManager(cfg)
	mgr.Run()
	<- ready
}

// LoadScaleModelAndMetrics dynamically reloads model and metrics for scale testing
func LoadScaleModelAndMetrics(t *testing.T) {
	dataSets := make([]*modelapi.DataSet, 0, 2)

	dataSets, err := utils.AddDataSet(dataSets, "model", "onos-pci/files/test/scale-model.yaml")
	assert.NoError(t, err)

	dataSets, err = utils.AddDataSet(dataSets, "rc.pci", "onos-pci/files/test/scale-rc-pci.yaml")
	assert.NoError(t, err)

	err = utils.LoadNewModel(dataSets)
	assert.NoError(t, err)
}
