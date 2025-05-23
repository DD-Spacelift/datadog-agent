// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build darwin

// Package apm contains the APM check
package apm

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"go.uber.org/atomic"
	yaml "gopkg.in/yaml.v2"

	"github.com/DataDog/datadog-agent/comp/core/autodiscovery/integration"
	diagnose "github.com/DataDog/datadog-agent/comp/core/diagnose/def"
	"github.com/DataDog/datadog-agent/pkg/aggregator/sender"
	"github.com/DataDog/datadog-agent/pkg/collector/check"
	checkid "github.com/DataDog/datadog-agent/pkg/collector/check/id"
	"github.com/DataDog/datadog-agent/pkg/collector/check/stats"
	"github.com/DataDog/datadog-agent/pkg/collector/corechecks"
	"github.com/DataDog/datadog-agent/pkg/collector/corechecks/embed/common"
	pkgconfigsetup "github.com/DataDog/datadog-agent/pkg/config/setup"
	"github.com/DataDog/datadog-agent/pkg/config/utils"
	"github.com/DataDog/datadog-agent/pkg/util/hostname"
	"github.com/DataDog/datadog-agent/pkg/util/log"
	"github.com/DataDog/datadog-agent/pkg/util/option"
)

const (
	// CheckName is the name of the check
	CheckName = "apm"
)

type apmCheckConf struct {
	BinPath string `yaml:"bin_path,omitempty"`
}

// APMCheck keeps track of the running command
type APMCheck struct {
	binPath        string
	commandOpts    []string
	running        *atomic.Bool
	stop           chan struct{}
	stopDone       chan struct{}
	source         string
	telemetry      bool
	initConfig     string
	instanceConfig string
}

// String displays the Agent name
func (c *APMCheck) String() string {
	return "APM Agent"
}

// Version displays the command's version
func (c *APMCheck) Version() string {
	return ""
}

// ConfigSource displays the command's source
func (c *APMCheck) ConfigSource() string {
	return c.source
}

// Loader returns the check loader
func (*APMCheck) Loader() string {
	// the apm check is scheduled by the Go loader
	return corechecks.GoCheckLoaderName
}

// Run executes the check with retries
func (c *APMCheck) Run() error {
	c.running.Store(true)
	// TODO: retries should be configurable with meaningful default values
	err := check.Retry(common.DefaultRetryDuration, common.DefaultRetries, c.run, c.String())
	c.running.Store(false)

	return err
}

// run executes the check
func (c *APMCheck) run() error {
	select {
	// poll the stop channel once to make sure no stop was requested since the last call to `run`
	case <-c.stop:
		log.Info("Not starting APM check: stop requested")
		c.stopDone <- struct{}{}
		return nil
	default:
	}

	cmd := exec.Command(c.binPath, c.commandOpts...)

	hname, _ := hostname.Get(context.TODO())

	env := os.Environ()
	env = append(env, fmt.Sprintf("DD_API_KEY=%s", utils.SanitizeAPIKey(pkgconfigsetup.Datadog().GetString("api_key"))))
	env = append(env, fmt.Sprintf("DD_HOSTNAME=%s", hname))
	env = append(env, fmt.Sprintf("DD_DOGSTATSD_PORT=%s", pkgconfigsetup.Datadog().GetString("dogstatsd_port")))
	env = append(env, fmt.Sprintf("DD_LOG_LEVEL=%s", pkgconfigsetup.Datadog().GetString("log_level")))
	cmd.Env = env

	// forward the standard output to the Agent logger
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	go func() {
		in := bufio.NewScanner(stdout)
		for in.Scan() {
			log.Info(in.Text())
		}
	}()

	// forward the standard error to the Agent logger
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	go func() {
		in := bufio.NewScanner(stderr)
		for in.Scan() {
			log.Error(in.Text())
		}
	}()

	if err = cmd.Start(); err != nil {
		return common.RetryExitError(err)
	}

	processDone := make(chan error)
	go func() {
		processDone <- cmd.Wait()
	}()

	select {
	case err = <-processDone:
		return common.RetryExitError(err)
	case <-c.stop:
		err = cmd.Process.Signal(os.Kill)
		if err != nil {
			log.Errorf("unable to stop APM check: %s", err)
		}
	}

	// wait for process to exit
	err = <-processDone
	c.stopDone <- struct{}{}
	return err
}

// Configure configures the APM check with the provided configuration
func (c *APMCheck) Configure(_ sender.SenderManager, _ uint64, data integration.Data, initConfig integration.Data, source string) error {
	var checkConf apmCheckConf
	if err := yaml.Unmarshal(data, &checkConf); err != nil {
		return err
	}

	c.binPath = ""
	defaultBinPath, defaultBinPathErr := getAPMAgentDefaultBinPath()
	if checkConf.BinPath != "" {
		if _, err := os.Stat(checkConf.BinPath); err == nil {
			c.binPath = checkConf.BinPath
		} else {
			log.Warnf("Can't access apm binary at %s, falling back to default path at %s", checkConf.BinPath, defaultBinPath)
		}
	}

	if c.binPath == "" {
		if defaultBinPathErr != nil {
			return defaultBinPathErr
		}

		c.binPath = defaultBinPath
	}

	configFile := pkgconfigsetup.Datadog().ConfigFileUsed()

	c.commandOpts = []string{}

	// explicitly provide to the trace-agent the agent configuration file
	if _, err := os.Stat(configFile); !os.IsNotExist(err) {
		c.commandOpts = append(c.commandOpts, fmt.Sprintf("-config=%s", configFile))
	}

	c.source = source
	c.telemetry = utils.IsCheckTelemetryEnabled("apm", pkgconfigsetup.Datadog())
	c.initConfig = string(initConfig)
	c.instanceConfig = string(data)
	return nil
}

// Interval returns the scheduling time for the check, this will be scheduled only once
// since `Run` won't return, thus implementing a long running check.
func (c *APMCheck) Interval() time.Duration {
	return 0
}

// ID returns the name of the check since there should be only one instance running
func (c *APMCheck) ID() checkid.ID {
	return "APM_AGENT"
}

// IsTelemetryEnabled returns if the telemetry is enabled for this check
func (c *APMCheck) IsTelemetryEnabled() bool {
	return c.telemetry
}

// Stop sends a termination signal to the APM process
func (c *APMCheck) Stop() {
	if !c.running.Load() {
		log.Info("APM Agent not running.")
		return
	}

	c.stop <- struct{}{}
	<-c.stopDone
}

// Cancel does nothing
func (c *APMCheck) Cancel() {}

// GetWarnings does not return anything in APM
func (c *APMCheck) GetWarnings() []error {
	return []error{}
}

// GetSenderStats returns the stats from the last run of the check, but there aren't any
func (c *APMCheck) GetSenderStats() (stats.SenderStats, error) {
	return stats.NewSenderStats(), nil
}

// InitConfig returns the initConfig for the APM check
func (c *APMCheck) InitConfig() string {
	return c.initConfig
}

// InstanceConfig returns the instance config for the APM check
func (c *APMCheck) InstanceConfig() string {
	return c.instanceConfig
}

// GetDiagnoses returns the diagnoses of the check
func (c *APMCheck) GetDiagnoses() ([]diagnose.Diagnosis, error) {
	return nil, nil
}

// IsHASupported returns if the check is compatible with High Availability
func (c *APMCheck) IsHASupported() bool {
	return false
}

// Factory creates a new check factory
func Factory() option.Option[func() check.Check] {
	return option.New(newCheck)
}

func newCheck() check.Check {
	return &APMCheck{
		running:  atomic.NewBool(false),
		stop:     make(chan struct{}),
		stopDone: make(chan struct{}),
	}
}
