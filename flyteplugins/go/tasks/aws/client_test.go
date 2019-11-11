/*
 * Copyright (c) 2018 Lyft. All rights reserved.
 */

package aws

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_GetConfig(t *testing.T) {
	assert.NoError(t, Init(context.TODO(), &Config{
		Retries: 2,
		Region:  "us-east-1",
	}))

	c, err := GetClient()
	assert.NoError(t, err)
	assert.NotNil(t, c)
	assert.NotNil(t, GetConfig())
}
