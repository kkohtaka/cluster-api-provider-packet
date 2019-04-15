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

	errors "golang.org/x/xerrors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	clusterclient "sigs.k8s.io/cluster-api/pkg/client/clientset_generated/clientset/typed/cluster/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	packet "github.com/kkohtaka/cluster-api-provider-packet/pkg/cloud/packet/client"
	"github.com/kkohtaka/cluster-api-provider-packet/pkg/cloud/packet/util"
)

const (
	ProviderName = "packet"
)

// Actuator is responsible for performing machine reconciliation
type Actuator struct {
	machinesGetter clusterclient.MachinesGetter
	client         client.Client
}

// ActuatorParams holds parameter information for Actuator
type ActuatorParams struct {
	MachinesGetter clusterclient.MachinesGetter
	Client         client.Client
}

// NewActuator creates a new Actuator
func NewActuator(params ActuatorParams) (*Actuator, error) {
	return &Actuator{
		machinesGetter: params.MachinesGetter,
		client:         params.Client,
	}, nil
}

// Create creates a machine and is invoked by the Machine Controller
func (a *Actuator) Create(ctx context.Context, cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	log.Printf("Creating machine %v for cluster %v.", machine.Name, cluster.Name)
	return fmt.Errorf("TODO: Not yet implemented")
}

// Delete deletes a machine and is invoked by the Machine Controller
func (a *Actuator) Delete(ctx context.Context, cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	log.Printf("Deleting machine %v for cluster %v.", machine.Name, cluster.Name)
	return fmt.Errorf("TODO: Not yet implemented")
}

// Update updates a machine and is invoked by the Machine Controller
func (a *Actuator) Update(ctx context.Context, cluster *clusterv1.Cluster, machine *clusterv1.Machine) error {
	log.Printf("Updating machine %v for cluster %v.", machine.Name, cluster.Name)
	return fmt.Errorf("TODO: Not yet implemented")
}

// Exists test for the existance of a machine and is invoked by the Machine Controller
func (a *Actuator) Exists(ctx context.Context, cluster *clusterv1.Cluster, machine *clusterv1.Machine) (bool, error) {
	log.Printf("Checking if machine %v for cluster %v exists.", machine.Name, cluster.Name)
	if cluster == nil {
		return false, errors.Errorf("missing cluster for machine %v/%v", machine.Namespace, machine.Name)
	}
	clusterSpec, err := util.ToClusterProviderSpec(&cluster.Spec)
	if err != nil {
		return false, errors.Errorf("decode machine provider clusterSpec for machine %v/%v: %w",
			machine.Namespace, machine.Name, err)
	}

	// TODO(kkohtaka): Should check missing client?
	if a.client == nil {
		return false, errors.Errorf("missing clinet")
	}
	var secret corev1.Secret
	objKey := types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      clusterSpec.SecretRef,
	}
	err = a.client.Get(ctx, objKey, &secret)
	if err != nil {
		return false, errors.Errorf("get secret %v: %w", objKey, err)
	}

	status, err := util.ToMachineProviderStatus(&machine.Status)
	if err != nil {
		return false, errors.Errorf("decode machine provider status for machine %v/%v: %w",
			machine.Namespace, machine.Name, err)
	}
	if status.ID == "" {
		log.Printf("Machine %v/%v isn't created yet", machine.Namespace, machine.Name)
		return false, nil
	}

	c, err := packet.NewClient(&secret)
	if err != nil {
		return false, errors.Errorf("create Packet API client: %w", err)
	}
	exist, err := c.DoesDeviceExist(status.ID)
	if err != nil {
		return false, errors.Errorf("check Device existence for machine %v/%v: %w",
			machine.Namespace, machine.Name, err)
	}
	return exist, nil
}

// The Machine Actuator interface must implement GetIP and GetKubeConfig functions as a workaround for issues
// cluster-api#158 (https://github.com/kubernetes-sigs/cluster-api/issues/158) and cluster-api#160
// (https://github.com/kubernetes-sigs/cluster-api/issues/160).

// GetIP returns IP address of the machine in the cluster.
func (a *Actuator) GetIP(cluster *clusterv1.Cluster, machine *clusterv1.Machine) (string, error) {
	log.Printf("Getting IP of machine %v for cluster %v.", machine.Name, cluster.Name)
	return "", fmt.Errorf("TODO: Not yet implemented")
}

// GetKubeConfig gets a kubeconfig from the master.
func (a *Actuator) GetKubeConfig(cluster *clusterv1.Cluster, master *clusterv1.Machine) (string, error) {
	log.Printf("Getting IP of machine %v for cluster %v.", master.Name, cluster.Name)
	return "", fmt.Errorf("TODO: Not yet implemented")
}
