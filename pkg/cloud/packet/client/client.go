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

	packetv1alpha1 "github.com/kkohtaka/cluster-api-provider-packet/pkg/apis/packet/v1alpha1"
)

const (
	secretKeyAPIKey = "apiKey"

	defaultBillingCycle = "hourly"
)

type Client interface {
	// GetProjectID returns ID of specified project
	GetProjectID(project string) (string, error)
	// DoesDeviceExist returns true iff a specified device exists
	DoesDeviceExist(deviceID string) (bool, error)
	// CreateDevice creates a device on Packet
	CreateDevice(spec *packetv1alpha1.PacketMachineProviderSpec) (*packetv1alpha1.PacketMachineProviderStatus, error)
	// GetDevice gets a device on Packet
	GetDevice(deviceID string) (*packetv1alpha1.PacketMachineProviderStatus, error)
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

func (c *client) GetProjectID(project string) (string, error) {
	p, _, err := c.c.Projects.List(&packngo.ListOptions{
		Includes: []string{project},
	})
	if err != nil {
		return "", errors.Errorf("list projects: %w", err)
	}
	if len(p) == 0 {
		return "", errors.Errorf("find project by name: %v", project)
	}
	return p[0].ID, nil
}

func (c *client) CreateDevice(spec *packetv1alpha1.PacketMachineProviderSpec) (*packetv1alpha1.PacketMachineProviderStatus, error) {
	if spec.BillingCycle == "" {
		spec.BillingCycle = defaultBillingCycle
	}
	d, _, err := c.c.Devices.Create(
		&packngo.DeviceCreateRequest{
			ProjectID:    spec.ProjectID,
			Facility:     []string{spec.Facility},
			Plan:         spec.Plan,
			Hostname:     spec.Hostname,
			OS:           spec.OS,
			BillingCycle: spec.BillingCycle,
			UserData:     spec.UserData,
		},
	)
	if err != nil {
		return nil, err
	}
	return newStatus(d), nil
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

func (c *client) GetDevice(deviceID string) (*packetv1alpha1.PacketMachineProviderStatus, error) {
	device, _, err := c.c.Devices.Get(deviceID, nil)
	if err != nil {
		return nil, err
	}
	return newStatus(device), nil
}

func newStatus(d *packngo.Device) *packetv1alpha1.PacketMachineProviderStatus {
	status := &packetv1alpha1.PacketMachineProviderStatus{}
	status.State = packetv1alpha1.StringToState(d.State)
	status.ID = d.ID
	status.IPAddresses = make([]packetv1alpha1.IPAddress, len(d.Network))
	for i := range d.Network {
		ipAddress := d.Network[i]
		status.IPAddresses[i] = packetv1alpha1.IPAddress{
			ID:            ipAddress.ID,
			Address:       ipAddress.Address,
			Gateway:       ipAddress.Gateway,
			Network:       ipAddress.Network,
			AddressFamily: ipAddress.AddressFamily,
			Netmask:       ipAddress.Netmask,
			Public:        ipAddress.Public,
		}
	}

	status.Ready = status.State == packetv1alpha1.StateActive

	return status
}
