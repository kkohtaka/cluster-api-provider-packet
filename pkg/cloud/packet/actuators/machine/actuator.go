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

package machine

import (
	"context"
	"fmt"
	"log"
	"time"

	errors "golang.org/x/xerrors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	clustererror "sigs.k8s.io/cluster-api/pkg/controller/error"
	"sigs.k8s.io/controller-runtime/pkg/client"

	packetv1alpha1 "github.com/kkohtaka/cluster-api-provider-packet/pkg/apis/packet/v1alpha1"
	packet "github.com/kkohtaka/cluster-api-provider-packet/pkg/cloud/packet/client"
	"github.com/kkohtaka/cluster-api-provider-packet/pkg/cloud/packet/util"
)

const (
	ProviderName = "packet"

	DefaultOS = "coreos_stable"
)

// Actuator is responsible for performing machine reconciliation
type Actuator struct {
	client client.Client
}

// ActuatorParams holds parameter information for Actuator
type ActuatorParams struct {
	Client client.Client
}

// NewActuator creates a new Actuator
func NewActuator(params ActuatorParams) (*Actuator, error) {
	return &Actuator{
		client: params.Client,
	}, nil
}

// Create creates a machine and is invoked by the Machine Controller
func (a *Actuator) Create(ctx context.Context, cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	machineKey := types.NamespacedName{
		Namespace: machine.Namespace,
		Name:      machine.Name,
	}

	log.Printf("Creating machine %v for cluster %v.", machineKey, cluster.Name)

	if cluster == nil {
		return errors.Errorf("missing cluster for machine %v", machineKey)
	}

	clusterSpec, err := util.ToClusterProviderSpec(&cluster.Spec)
	if err != nil {
		return errors.Errorf("decode cluster provider spec for machine %v: %w", machineKey, err)
	}

	clusterStatus, err := util.ToClusterProviderStatus(&cluster.Status)
	if err != nil {
		return errors.Errorf("decode cluster provider status for machine %v: %w", machineKey, err)
	}
	if clusterStatus.ProjectID == "" {
		return errors.Errorf("Packet project ID is not set in cluster provider status for machine %v", machineKey)
	}

	spec, err := util.ToMachineProviderSpec(&machine.Spec)
	if err != nil {
		return errors.Errorf("decode machine provider spec for machine %v: %w", machineKey, err)
	}

	status, err := util.ToMachineProviderStatus(&machine.Status)
	if err != nil {
		return errors.Errorf("decode machine provider status for machine %v: %w", machineKey, err)
	}

	if status.ID != "" {
		log.Printf("machine already has .Status.ID: %v", status.ID)
	}

	newSpec := spec.DeepCopy()
	newSpec.ProjectID = clusterStatus.ProjectID
	newSpec.Hostname = machine.Name
	newSpec.Facility = clusterSpec.Facility
	newSpec.Plan = clusterSpec.Plan
	newSpec.BillingCycle = clusterSpec.BillingCycle
	newSpec.OS = DefaultOS

	err = util.UpdateMachineProviderSpec(a.client, machineKey, newSpec)
	if err != nil {
		return errors.Errorf("update spec for machine %v: %w", machineKey, err)
	}

	var secret corev1.Secret
	secretKey := types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      clusterSpec.SecretRef,
	}
	err = a.client.Get(ctx, secretKey, &secret)
	if err != nil {
		return errors.Errorf("get secret %v: %w", secretKey, err)
	}
	c, err := packet.NewClient(&secret)
	if err != nil {
		return errors.Errorf("create Packet API client: %w", err)
	}

	newStatus, err := c.CreateDevice(newSpec)
	if err != nil {
		return errors.Errorf("create device for machine %v: %w", machineKey, err)
	}

	err = util.UpdateMachineProviderStatus(a.client, machineKey, newStatus)
	if err != nil {
		return errors.Errorf("update status for machine %v: %w", machineKey, err)
	}

	return &clustererror.RequeueAfterError{
		RequeueAfter: 15 * time.Second,
	}
}

// Delete deletes a machine and is invoked by the Machine Controller
func (a *Actuator) Delete(ctx context.Context, cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	machineKey := types.NamespacedName{
		Namespace: machine.Namespace,
		Name:      machine.Name,
	}

	log.Printf("Deleting machine %v for cluster %v.", machineKey, cluster.Name)

	if cluster == nil {
		return errors.Errorf("missing cluster for machine %v", machineKey)
	}

	clusterSpec, err := util.ToClusterProviderSpec(&cluster.Spec)
	if err != nil {
		return errors.Errorf("decode cluster provider spec for machine %v: %w", machineKey, err)
	}
	var secret corev1.Secret
	secretKey := types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      clusterSpec.SecretRef,
	}
	err = a.client.Get(ctx, secretKey, &secret)
	if err != nil {
		return errors.Errorf("get secret %v: %w", secretKey, err)
	}
	c, err := packet.NewClient(&secret)
	if err != nil {
		return errors.Errorf("create Packet API client: %w", err)
	}

	status, err := util.ToMachineProviderStatus(&machine.Status)
	if err != nil {
		return errors.Errorf("decode machine provider status for machine %v: %w", machineKey, err)
	}
	if status.ID != "" {
		err = c.DeleteDevice(status.ID)
		if err != nil {
			if packet.IsNotFoundError(err) {
				log.Printf("specified device %v for machine %v has been already deleted",
					status.ID, machineKey)
			} else {
				return errors.Errorf("delete a device for machine %v: %w", machineKey, err)
			}
		}

		newStatus := &packetv1alpha1.PacketMachineProviderStatus{}

		err = util.UpdateMachineProviderStatus(a.client, machineKey, newStatus)
		if err != nil {
			return errors.Errorf("update machine %v", machineKey, err)
		}
	}
	return nil
}

// Update updates a machine and is invoked by the Machine Controller
func (a *Actuator) Update(ctx context.Context, cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	machineKey := types.NamespacedName{
		Namespace: machine.Namespace,
		Name:      machine.Name,
	}

	log.Printf("Updating machine %v for cluster %v.", machineKey, cluster.Name)

	if cluster == nil {
		return errors.Errorf("missing cluster for machine %v", machineKey)
	}

	clusterSpec, err := util.ToClusterProviderSpec(&cluster.Spec)
	if err != nil {
		return errors.Errorf("decode cluster provider spec for machine %v: %w", machineKey, err)
	}
	var secret corev1.Secret
	secretKey := types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      clusterSpec.SecretRef,
	}
	err = a.client.Get(ctx, secretKey, &secret)
	if err != nil {
		return errors.Errorf("get secret %v: %w", secretKey, err)
	}
	c, err := packet.NewClient(&secret)
	if err != nil {
		return errors.Errorf("create Packet API client: %w", err)
	}

	status, err := util.ToMachineProviderStatus(&machine.Status)
	if err != nil {
		return errors.Errorf("decode machine provider status for machine %v: %w", machineKey, err)
	}

	newStatus, err := c.GetDevice(status.ID)
	if err != nil {
		if packet.IsNotFoundError(err) {
			// If a device specified by .Status.ID is not found, reset the .Status field
			newStatus = &packetv1alpha1.PacketMachineProviderStatus{}
		} else {
			return errors.Errorf("get device for machine %v: %w", machineKey, err)
		}
	}

	err = util.UpdateMachineProviderStatus(a.client, machineKey, newStatus)
	if err != nil {
		return errors.Errorf("update spec for machine %v", machineKey, err)
	}

	if !newStatus.Ready {
		return &clustererror.RequeueAfterError{
			RequeueAfter: 15 * time.Second,
		}
	}
	return nil
}

// Exists test for the existance of a machine and is invoked by the Machine Controller
func (a *Actuator) Exists(ctx context.Context, cluster *clusterv1.Cluster, machine *clusterv1.Machine) (bool, error) {
	machineKey := types.NamespacedName{
		Namespace: machine.Namespace,
		Name:      machine.Name,
	}

	log.Printf("Checking if machine %v for cluster %v exists.", machineKey, cluster.Name)

	if cluster == nil {
		return false, errors.Errorf("missing cluster for machine %v", machineKey)
	}

	clusterSpec, err := util.ToClusterProviderSpec(&cluster.Spec)
	if err != nil {
		return false, errors.Errorf("decode machine provider clusterSpec for machine %v: %w", machineKey, err)
	}

	var secret corev1.Secret
	secretKey := types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      clusterSpec.SecretRef,
	}
	err = a.client.Get(ctx, secretKey, &secret)
	if err != nil {
		return false, errors.Errorf("get secret %v: %w", secretKey, err)
	}

	status, err := util.ToMachineProviderStatus(&machine.Status)
	if err != nil {
		return false, errors.Errorf("decode machine provider status for machine %v: %w", machineKey, err)
	}
	if status.ID == "" {
		return false, nil
	}

	c, err := packet.NewClient(&secret)
	if err != nil {
		return false, errors.Errorf("create Packet API client: %w", err)
	}
	exist, err := c.DoesDeviceExist(status.ID)
	if err != nil {
		return false, errors.Errorf("check device existence for machine %v: %w", machineKey, err)
	}
	return exist, nil
}

// The Machine Actuator interface must implement GetIP and GetKubeConfig functions as a workaround for issues
// cluster-api#158 (https://github.com/kubernetes-sigs/cluster-api/issues/158) and cluster-api#160
// (https://github.com/kubernetes-sigs/cluster-api/issues/160).

// GetIP returns IP address of the machine in the cluster.
func (a *Actuator) GetIP(cluster *clusterv1.Cluster, machine *clusterv1.Machine) (string, error) {
	machineKey := types.NamespacedName{
		Namespace: machine.Namespace,
		Name:      machine.Name,
	}
	log.Printf("Getting IP of machine %v for cluster %v.", machineKey, cluster.Name)
	return "", fmt.Errorf("TODO: Not yet implemented")
}

// GetKubeConfig gets a kubeconfig from the master.
func (a *Actuator) GetKubeConfig(cluster *clusterv1.Cluster, master *clusterv1.Machine) (string, error) {
	log.Printf("Getting IP of machine %v for cluster %v.", master, cluster.Name)
	return "", fmt.Errorf("TODO: Not yet implemented")
}
