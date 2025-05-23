// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build test

package aggregator

import (
	"errors"

	log "github.com/DataDog/datadog-agent/comp/core/log/def"
	"github.com/DataDog/datadog-agent/comp/forwarder/defaultforwarder"
	"github.com/DataDog/datadog-agent/pkg/aggregator/sender"
	checkid "github.com/DataDog/datadog-agent/pkg/collector/check/id"
	pkgconfigsetup "github.com/DataDog/datadog-agent/pkg/config/setup"
)

// PeekSender returns a Sender with passed ID or an error if the sender is not registered
func (s *senders) PeekSender(cid checkid.ID) (sender.Sender, error) {
	return s.senderPool.getSender(cid)
}

// PeekSender returns a Sender with passed ID or an error if the sender is not registered
func (d *AgentDemultiplexer) PeekSender(cid checkid.ID) (sender.Sender, error) {
	d.m.Lock()
	defer d.m.Unlock()
	if d.senders == nil {
		return nil, errors.New("demultiplexer is stopped")
	}
	return d.senders.PeekSender(cid)
}

//nolint:revive // TODO(AML) Fix revive linter
func NewForwarderTest(log log.Component) defaultforwarder.Forwarder {
	options, _ := defaultforwarder.NewOptions(pkgconfigsetup.Datadog(), log, nil)
	return defaultforwarder.NewDefaultForwarder(pkgconfigsetup.Datadog(), log, options)
}
