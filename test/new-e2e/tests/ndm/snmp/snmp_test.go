// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

// Package snmp contains e2e tests for snmp
package snmp

import (
	"embed"
	"fmt"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DataDog/datadog-agent/test/fakeintake/aggregator"
	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/e2e"
	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/environments"
	"github.com/DataDog/datadog-agent/test/new-e2e/pkg/provisioners"

	"github.com/DataDog/test-infra-definitions/common/utils"
	"github.com/DataDog/test-infra-definitions/components/datadog/agent"
	"github.com/DataDog/test-infra-definitions/components/datadog/dockeragentparams"
	"github.com/DataDog/test-infra-definitions/components/docker"
	"github.com/DataDog/test-infra-definitions/components/os"
	"github.com/DataDog/test-infra-definitions/resources/aws"
	"github.com/DataDog/test-infra-definitions/scenarios/aws/ec2"
	"github.com/DataDog/test-infra-definitions/scenarios/aws/fakeintake"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

//go:embed compose/snmpCompose.yaml
var snmpCompose string

//go:embed config/cisco-nexus.yaml
var snmpConfig string

const (
	composeDataPath = "compose/data"
)

// snmpDockerProvisioner defines a stack with a docker agent on an AmazonLinuxECS VM
// with snmpsim installed and configured with snmp recordings
func snmpDockerProvisioner() provisioners.Provisioner {
	return provisioners.NewTypedPulumiProvisioner("", func(ctx *pulumi.Context, env *environments.DockerHost) error {
		name := "snmpvm"
		awsEnv, err := aws.NewEnvironment(ctx)
		if err != nil {
			return err
		}

		host, err := ec2.NewVM(awsEnv, name, ec2.WithOS(os.AmazonLinuxECSDefault))
		if err != nil {
			return err
		}
		host.Export(ctx, &env.RemoteHost.HostOutput)

		fakeIntake, err := fakeintake.NewECSFargateInstance(awsEnv, name)
		if err != nil {
			return err
		}
		fakeIntake.Export(ctx, &env.FakeIntake.FakeintakeOutput)

		// Setting up SNMP-related files
		filemanager := host.OS.FileManager()
		// upload snmpsim data files
		createDataDirCommand, dataPath, err := filemanager.TempDirectory("data")
		if err != nil {
			return err
		}
		dataFiles, err := loadDataFileNames()
		if err != nil {
			return err
		}

		fileCommands := []pulumi.Resource{}
		for _, fileName := range dataFiles {
			fileContent, err := dataFolder.ReadFile(path.Join(composeDataPath, fileName))
			if err != nil {
				return err
			}
			fileCommand, err := filemanager.CopyInlineFile(pulumi.String(fileContent), path.Join(dataPath, fileName),
				pulumi.DependsOn([]pulumi.Resource{createDataDirCommand}))
			if err != nil {
				return err
			}
			fileCommands = append(fileCommands, fileCommand)
		}

		createConfigDirCommand, configPath, err := filemanager.TempDirectory("config")
		if err != nil {
			return err
		}
		// edit snmp config file
		configCommand, err := filemanager.CopyInlineFile(pulumi.String(snmpConfig), path.Join(configPath, "snmp.yaml"),
			pulumi.DependsOn([]pulumi.Resource{createConfigDirCommand}))
		if err != nil {
			return err
		}

		installEcrCredsHelperCmd, err := ec2.InstallECRCredentialsHelper(awsEnv, host)
		if err != nil {
			return err
		}

		dockerManager, err := docker.NewManager(&awsEnv, host, utils.PulumiDependsOn(installEcrCredsHelperCmd))
		if err != nil {
			return err
		}
		dockerManager.Export(ctx, &env.Docker.ManagerOutput)

		envVars := pulumi.StringMap{"DATA_DIR": pulumi.String(dataPath), "CONFIG_DIR": pulumi.String(configPath)}
		composeDependencies := []pulumi.Resource{createDataDirCommand, configCommand}
		composeDependencies = append(composeDependencies, fileCommands...)
		dockerAgent, err := agent.NewDockerAgent(&awsEnv, host, dockerManager,
			dockeragentparams.WithFakeintake(fakeIntake),
			dockeragentparams.WithExtraComposeManifest("snmpsim", pulumi.String(snmpCompose)),
			dockeragentparams.WithEnvironmentVariables(envVars),
			dockeragentparams.WithPulumiDependsOn(pulumi.DependsOn(composeDependencies)),
		)
		if err != nil {
			return err
		}
		dockerAgent.Export(ctx, &env.Agent.DockerAgentOutput)

		return err
	}, nil)
}

//go:embed compose/data
var dataFolder embed.FS

func loadDataFileNames() (out []string, err error) {
	fileEntries, err := dataFolder.ReadDir(composeDataPath)
	if err != nil {
		return nil, err
	}
	for _, f := range fileEntries {
		out = append(out, f.Name())
	}
	return out, nil
}

type snmpDockerSuite struct {
	e2e.BaseSuite[environments.DockerHost]
}

// TestSnmpSuite runs the snmp e2e suite
func TestSnmpSuite(t *testing.T) {
	t.Skip("Skipping test until we can fix the flakiness")
	e2e.Run(t, &snmpDockerSuite{}, e2e.WithProvisioner(snmpDockerProvisioner()))
}

// TestSnmp tests that the snmpsim container is running and that the agent container
// is sending snmp metrics to the fakeintake
func (s *snmpDockerSuite) TestSnmp() {
	fakeintake := s.Env().FakeIntake.Client()
	s.EventuallyWithT(func(c *assert.CollectT) {
		metrics, err := fakeintake.GetMetricNames()
		assert.NoError(c, err)
		assert.Contains(c, metrics, "snmp.sysUpTimeInstance", "metrics %v doesn't contain snmp.sysUpTimeInstance", metrics)
	}, 5*time.Minute, 10*time.Second)
}

func (s *snmpDockerSuite) TestSnmpTagsAreStoredOnRestart() {
	fakeintake := s.Env().FakeIntake.Client()
	var initialMetrics []*aggregator.MetricSeries
	var err error

	require.EventuallyWithT(s.T(), func(t *assert.CollectT) {
		initialMetrics, err = fakeintake.FilterMetrics("snmp.device.reachable")
		assert.NoError(t, err)
	}, 5*time.Minute, 5*time.Second)

	initialTags := initialMetrics[0].Tags
	_, err = s.Env().RemoteHost.Execute("docker stop dd-snmp")
	require.NoError(s.T(), err)

	_, err = s.Env().RemoteHost.Execute(fmt.Sprintf("docker restart %s", s.Env().Agent.ContainerName))
	require.NoError(s.T(), err)

	err = fakeintake.FlushServerAndResetAggregators()
	require.NoError(s.T(), err)

	var metrics []*aggregator.MetricSeries

	require.EventuallyWithT(s.T(), func(t *assert.CollectT) {
		metrics, err = fakeintake.FilterMetrics("snmp.device.reachable")
		assert.NoError(t, err)
		assert.NotEmpty(t, metrics)
	}, 5*time.Minute, 5*time.Second)

	tags := metrics[0].Tags

	require.Zero(s.T(), metrics[0].Points[0].Value)
	require.ElementsMatch(s.T(), tags, initialTags)
}
