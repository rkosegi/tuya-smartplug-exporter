/*
Copyright 2026 Richard Kosegi

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package internal

import (
	"os"
	"testing"
	"time"

	"github.com/rkosegi/tuya-proto/proto"
	"github.com/stretchr/testify/assert"
)

var (
	device31Host = os.Getenv("DEVICE_31_HOST")
	device31Key  = os.Getenv("DEVICE_31_KEY")
	device31Id   = os.Getenv("DEVICE_31_ID")
	device34Host = os.Getenv("DEVICE_34_HOST")
	device34Key  = os.Getenv("DEVICE_34_KEY")
)

func Test31(t *testing.T) {
	if device31Host == "" || device31Key == "" || device31Id == "" {
		t.Skip("Protocol v3.1 device is not configured, skipping")
	}
	c := NewClient(proto.Version31, device31Host, []byte(device31Key),
		WithTimeout(3*time.Second),
		WithReadTimeout(30*time.Second),
		WithWriteTimeout(10*time.Second))
	err := c.Connect()
	assert.NoError(t, err)
	assert.NotNil(t, c)

	err = c.Send(proto.CmdIdTypeDpQuery, &DpQueryRequest{
		GwId:  device31Id,
		DevId: device31Id,
	})
	assert.NoError(t, err)

	var out DpQueryResponse
	err = c.Read(&out)
	assert.NoError(t, err)

	assert.NoError(t, c.Close())
}

func Test34(t *testing.T) {
	if device34Host == "" || device34Key == "" {
		t.Skip("Protocol v3.4 device is not configured, skipping")
	}
	c := NewClient(proto.Version34, device34Host, []byte(device34Key),
		WithTimeout(10*time.Second),
		WithReadTimeout(30*time.Second),
		WithWriteTimeout(10*time.Second))
	err := c.Connect()
	assert.NoError(t, err)
	assert.NotNil(t, c)
	if err != nil {
		t.Fatal(err)
	}
	err = c.Send(proto.CmdIdTypeDpQueryNew, make(map[string]interface{}))
	assert.NoError(t, err)

	var out DpQueryResponse
	err = c.Read(&out)
	assert.NoError(t, err)

	t.Log(out)

	assert.NoError(t, c.Close())
}
