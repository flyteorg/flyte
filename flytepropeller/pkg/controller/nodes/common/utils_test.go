package common

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
)

type ParentInfo struct {
	uniqueID         string
	attempt          uint32
	isInDynamicChain bool
}

func (p ParentInfo) GetUniqueID() v1alpha1.NodeID {
	return p.uniqueID
}

func (p ParentInfo) CurrentAttempt() uint32 {
	return p.attempt
}

func (p ParentInfo) IsInDynamicChain() bool {
	return p.isInDynamicChain
}

func TestGenerateUniqueID(t *testing.T) {
	p := ParentInfo{
		uniqueID: "u1",
		attempt:  uint32(2),
	}
	uniqueID, err := GenerateUniqueID(p, "n1")
	assert.Nil(t, err)
	assert.Equal(t, "u1-2-n1", uniqueID)
}

func TestGenerateUniqueIDLong(t *testing.T) {
	p := ParentInfo{
		uniqueID: "u1111111111323131231asdadasd",
		attempt:  uint32(2),
	}
	uniqueID, err := GenerateUniqueID(p, "n1")
	assert.Nil(t, err)
	assert.Equal(t, "fraibbty", uniqueID)
}

func TestCreateParentInfo(t *testing.T) {
	gp := ParentInfo{
		uniqueID:         "u1",
		attempt:          uint32(2),
		isInDynamicChain: true,
	}
	parent, err := CreateParentInfo(gp, "n1", uint32(1), false)
	assert.Nil(t, err)
	assert.Equal(t, "u1-2-n1", parent.GetUniqueID())
	assert.Equal(t, uint32(1), parent.CurrentAttempt())
	assert.True(t, parent.IsInDynamicChain())
}

func TestCreateParentInfoNil(t *testing.T) {
	parent, err := CreateParentInfo(nil, "n1", uint32(1), true)
	assert.Nil(t, err)
	assert.Equal(t, "n1", parent.GetUniqueID())
	assert.Equal(t, uint32(1), parent.CurrentAttempt())
	assert.True(t, parent.IsInDynamicChain())
}
