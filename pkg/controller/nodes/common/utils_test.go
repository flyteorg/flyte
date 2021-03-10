package common

import (
	"testing"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/stretchr/testify/assert"
)

type ParentInfo struct {
	uniqueID string
	attempt  uint32
}

func (p ParentInfo) GetUniqueID() v1alpha1.NodeID {
	return p.uniqueID
}

func (p ParentInfo) CurrentAttempt() uint32 {
	return p.attempt
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
		uniqueID: "u1",
		attempt:  uint32(2),
	}
	parent, err := CreateParentInfo(gp, "n1", uint32(1))
	assert.Nil(t, err)
	assert.Equal(t, "u1-2-n1", parent.GetUniqueID())
	assert.Equal(t, uint32(1), parent.CurrentAttempt())
}

func TestCreateParentInfoNil(t *testing.T) {
	parent, err := CreateParentInfo(nil, "n1", uint32(1))
	assert.Nil(t, err)
	assert.Equal(t, "n1", parent.GetUniqueID())
	assert.Equal(t, uint32(1), parent.CurrentAttempt())
}
