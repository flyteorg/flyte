package task

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/errors"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/task/codex"
)

type CodecVersion uint8

const (
	GobCodecVersion CodecVersion = iota
)

const currentCodec = GobCodecVersion

// TODO Configurable?
const MaxPluginStateSizeBytes = 256

type stateCodec interface {
	Encode(interface{}, io.Writer) error
	Decode(io.Reader, interface{}) error
}

type pluginStateManager struct {
	prevState        *bytes.Buffer
	prevStateVersion uint8
	newState         *bytes.Buffer
	newStateVersion  uint8
	codec            stateCodec
	codecVersion     CodecVersion
}

func (p *pluginStateManager) Put(stateVersion uint8, v interface{}) error {
	p.newStateVersion = stateVersion
	if v != nil {
		buf := make([]byte, 0, MaxPluginStateSizeBytes)
		p.newState = bytes.NewBuffer(buf)
		return p.codec.Encode(v, p.newState)
	}
	return nil
}

func (p *pluginStateManager) Reset() error {
	p.newState = nil
	return nil
}

func (p pluginStateManager) GetStateVersion() uint8 {
	return p.prevStateVersion
}

func (p pluginStateManager) Get(v interface{}) (stateVersion uint8, err error) {
	if p.prevState == nil {
		return p.prevStateVersion, nil
	}
	if v == nil {
		return p.prevStateVersion, fmt.Errorf("cannot get state for a nil object, please initialize the type before requesting")
	}
	return p.prevStateVersion, p.codec.Decode(p.prevState, v)
}

func (p pluginStateManager) GetCodeVersion() CodecVersion {
	return p.codecVersion
}

func newPluginStateManager(_ context.Context, prevCodecVersion CodecVersion, prevStateVersion uint32, prevState *bytes.Buffer) (*pluginStateManager, error) {
	if prevCodecVersion != currentCodec {
		return nil, errors.Errorf(errors.IllegalStateError, "x", "prev codec [%d] != current codec [%d]", prevCodecVersion, currentCodec)
	}
	return &pluginStateManager{
		codec:            codex.GobStateCodec{},
		codecVersion:     GobCodecVersion,
		prevStateVersion: uint8(prevStateVersion),
		prevState:        prevState,
	}, nil
}
