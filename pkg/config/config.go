// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"os"

	"github.com/onosproject/onos-ric-sdk-go/pkg/config/event"
	configurable "github.com/onosproject/onos-ric-sdk-go/pkg/config/registry"

	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/onos-pci/pkg/utils"
	app "github.com/onosproject/onos-ric-sdk-go/pkg/config/app/default"
	configutils "github.com/onosproject/onos-ric-sdk-go/pkg/config/utils"
)

var log = logging.GetLogger("config")

const defaultConfigPath = "/etc/onos/config/config.json"

// Config xApp configuration interface
type Config interface {
	GetReportPeriodWithPath(path string) (uint64, error)
	GetReportPeriod() (uint64, error)
	GetGranularityPeriod() (uint64, error)
	Watch(context.Context, chan event.Event) error
}

// NewConfig initialize the xApp config
func NewConfig() (*AppConfig, error) {
	if _, err := os.Stat(defaultConfigPath); os.IsNotExist(err) {
		return nil, nil
	}
	appConfig, err := configurable.RegisterConfigurable(defaultConfigPath, &configurable.RegisterRequest{})
	if err != nil {
		return nil, err
	}

	cfg := &AppConfig{
		appConfig: appConfig.Config.(*app.Config),
	}
	return cfg, nil
}

// AppConfig application configuration
type AppConfig struct {
	appConfig *app.Config
}

// Watch watch config changes
func (c *AppConfig) Watch(ctx context.Context, ch chan event.Event) error {
	err := c.appConfig.Watch(ctx, ch)
	if err != nil {
		return err
	}
	return nil
}

// GetReportPeriodWithPath gets report period with a given path
func (c *AppConfig) GetReportPeriodWithPath(path string) (uint64, error) {
	interval, _ := c.appConfig.Get(path)
	val, err := configutils.ToUint64(interval.Value)
	if err != nil {
		log.Error(err)
		return 0, err
	}

	return val, nil
}

// GetReportPeriod gets report period
func (c *AppConfig) GetReportPeriod() (uint64, error) {
	interval, _ := c.appConfig.Get(utils.ReportPeriodConfigPath)
	val, err := configutils.ToUint64(interval.Value)
	if err != nil {
		log.Error(err)
		return 0, err
	}

	return val, nil
}

// GetGranularityPeriod gets granularity period
func (c *AppConfig) GetGranularityPeriod() (uint64, error) {
	granularity, _ := c.appConfig.Get(utils.GranularityPeriodConfigPath)
	val, err := configutils.ToUint64(granularity.Value)
	if err != nil {
		log.Error(err)
		return 0, err
	}
	return val, nil
}

var _ Config = &AppConfig{}
