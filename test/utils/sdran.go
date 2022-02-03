// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"fmt"
	"testing"

	"github.com/onosproject/helmit/pkg/input"
	"github.com/onosproject/onos-api/go/onos/pci"
	"github.com/onosproject/onos-pci/pkg/manager"
	"github.com/onosproject/onos-pci/pkg/northbound"
	"github.com/onosproject/onos-pci/pkg/store/metrics"

	"github.com/onosproject/helmit/pkg/helm"
	"github.com/onosproject/helmit/pkg/kubernetes"
	"github.com/onosproject/helmit/pkg/util/random"
	"github.com/onosproject/onos-test/pkg/onostest"
	"github.com/stretchr/testify/assert"
)

func getCredentials() (string, string, error) {
	kubClient, err := kubernetes.New()
	if err != nil {
		return "", "", err
	}
	secrets, err := kubClient.CoreV1().Secrets().Get(context.Background(), onostest.SecretsName)
	if err != nil {
		return "", "", err
	}
	username := string(secrets.Object.Data["sd-ran-username"])
	password := string(secrets.Object.Data["sd-ran-password"])

	return username, password, nil
}

// CreateSdranRelease creates a helm release for an sd-ran instance
func CreateSdranRelease(c *input.Context) (*helm.HelmRelease, error) {
	username, password, err := getCredentials()
	registry := c.GetArg("registry").String("")

	if err != nil {
		return nil, err
	}

	sdran := helm.Chart("sd-ran", onostest.SdranChartRepo).
		Release("sd-ran").
		SetUsername(username).
		SetPassword(password).
		Set("import.onos-config.enabled", false).
		Set("import.onos-topo.enabled", true).
		Set("import.ran-simulator.enabled", true).
		Set("import.onos-pci.enabled", false).
		Set("global.image.registry", registry)

	return sdran, nil
}

// CreateRanSimulator creates a ran simulator
func CreateRanSimulator(t *testing.T) *helm.HelmRelease {
	return CreateRanSimulatorWithName(t, random.NewPetName(2))
}

// CreateRanSimulatorWithNameOrDie creates a simulator and fails the test if the creation returned an error
func CreateRanSimulatorWithNameOrDie(t *testing.T, simName string) *helm.HelmRelease {
	sim := CreateRanSimulatorWithName(t, simName)
	assert.NotNil(t, sim)
	return sim
}

// CreateRanSimulatorWithName creates a ran simulator
func CreateRanSimulatorWithName(t *testing.T, name string) *helm.HelmRelease {
	username, password, err := getCredentials()
	assert.NoError(t, err)

	simulator := helm.
		Chart("ran-simulator", onostest.SdranChartRepo).
		Release(name).
		SetUsername(username).
		SetPassword(password).
		Set("image.tag", "latest").
		Set("fullnameOverride", "")
	err = simulator.Install(true)
	assert.NoError(t, err, "could not install device simulator %v", err)

	return simulator
}

// WaitForNoConflicts uses the current manager and events in the store to wait until no PCI conflicts exist
func WaitForNoConflicts(t *testing.T, mgr *manager.Manager) error {
	// Get the metrics store and a test service
	store := mgr.GetMetricsStore()
	server := northbound.NewTestServer(store)

	// Wait for changes in the metrics store...
	ch := make(chan metrics.Event)
	err := store.Watch(context.Background(), ch)
	assert.NoError(t, err)

	// After each event, check for number of remaining conflicts
	for e := range ch {
		pciEntry := e.Value.Value
		t.Log(fmt.Sprintf("Call %v has PCI %d", e.Key, pciEntry.Metric.PCI))

		resp, err := server.GetConflicts(context.Background(), &pci.GetConflictsRequest{})
		assert.NoError(t, err)
		if len(resp.Cells) == 0 {
			t.Log("All PCI conflicts eliminated")
			break
		}
		t.Log(fmt.Sprintf("Remaining PCI conflicts: %v", resp.Cells))
	}
	return err
}
