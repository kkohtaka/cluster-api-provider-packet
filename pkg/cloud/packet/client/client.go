/*
Copyright 2019 The Kubernetes Authors.

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

package client

import (
	"github.com/packethost/packngo"

	errors "golang.org/x/xerrors"

	corev1 "k8s.io/api/core/v1"
)

const (
	secretKeyAPIKey = "apiKey"
)

type Client interface {
	// DoesDeviceExist returns true iff a specified device exists
	DoesDeviceExist(deviceID string) (bool, error)
}

func NewClient(secret *corev1.Secret) (Client, error) {
	var (
		apiKey []byte
		ok     bool
	)
	if apiKey, ok = secret.Data[secretKeyAPIKey]; !ok {
		return nil, errors.Errorf(
			"secret %v/%v doesn't contain a key %v", secret.Namespace, secret.Name, secretKeyAPIKey)
	}
	return &client{
		c: packngo.NewClientWithAuth("", string(apiKey), nil),
	}, nil
}

type client struct {
	c *packngo.Client
}

func (c *client) DoesDeviceExist(deviceID string) (bool, error) {
	_, resp, err := c.c.Devices.Get(deviceID, nil)
	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
