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

package cluster

import (
	"context"
	"log"
	"reflect"

	errors "golang.org/x/xerrors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/cluster-api-provider-aws/pkg/cloud/aws/actuators"
	"sigs.k8s.io/cluster-api-provider-aws/pkg/deployer"
	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	packet "github.com/kkohtaka/cluster-api-provider-packet/pkg/cloud/packet/client"
	"github.com/kkohtaka/cluster-api-provider-packet/pkg/cloud/packet/util"
)

// +kubebuilder:rbac:groups=cluster.k8s.io,resources=machines;machines/status;machinedeployments;machinedeployments/status;machinesets;machinesets/status;machineclasses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.k8s.io,resources=clusters;clusters/status,verbs=get;list;watch

// Actuator is responsible for performing cluster reconciliation
type Actuator struct {
	*deployer.Deployer

	client client.Client
}

// ActuatorParams holds parameter information for Actuator
type ActuatorParams struct {
	Client client.Client
}

// NewActuator creates a new Actuator
func NewActuator(params ActuatorParams) (*Actuator, error) {
	return &Actuator{
		Deployer: deployer.New(deployer.Params{ScopeGetter: actuators.DefaultScopeGetter}),
		client:   params.Client,
	}, nil
}

// Reconcile reconciles a cluster and is invoked by the Cluster Controller
func (a *Actuator) Reconcile(cluster *clusterv1.Cluster) error {
	log.Printf("Reconciling cluster %v.", cluster.Name)

	spec, err := util.ToClusterProviderSpec(&cluster.Spec)
	if err != nil {
		return errors.Errorf("decode cluster provider spec for cluster %v/%v: %w",
			cluster.Namespace, cluster.Name, err)
	}

	status, err := util.ToClusterProviderStatus(&cluster.Status)
	if err != nil {
		return errors.Errorf("decode cluster provider status for cluster %v/%v: %w",
			cluster.Namespace, cluster.Name, err)
	}

	var secret corev1.Secret
	objKey := types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      spec.SecretRef,
	}
	err = a.client.Get(context.TODO(), objKey, &secret)
	if err != nil {
		return errors.Errorf("get secret %v: %w", objKey, err)
	}
	c, err := packet.NewClient(&secret)
	if err != nil {
		return errors.Errorf("create Packet API client: %w", err)
	}

	projectID, err := c.GetProjectID(spec.Project)
	if err != nil {
		return errors.Errorf("get project ID: %w", err)
	}

	newStatus := status.DeepCopy()
	newStatus.ProjectID = projectID

	if !reflect.DeepEqual(newStatus, status) {
		raw, err := util.ToRaw(newStatus)
		if err != nil {
			return errors.Errorf("generate raw data of cluster provider status for cluster %v/%v: %w",
				cluster.Namespace, cluster.Name, err)
		}
		newCluster := cluster.DeepCopy()
		newCluster.Status.ProviderStatus = &runtime.RawExtension{Raw: raw}

		err = a.client.Status().Update(context.TODO(), newCluster)
		if err != nil {
			return errors.Errorf("update status of cluster %v/%v", cluster.Namespace, cluster.Name, err)
		}
	}
	return nil
}

// Delete deletes a cluster and is invoked by the Cluster Controller
func (a *Actuator) Delete(cluster *clusterv1.Cluster) error {
	log.Printf("Deleting cluster %v.", cluster.Name)

	status, err := util.ToClusterProviderStatus(&cluster.Status)
	if err != nil {
		return errors.Errorf("decode cluster provider status for cluster %v/%v: %w",
			cluster.Namespace, cluster.Name, err)
	}

	newStatus := status.DeepCopy()
	newStatus.ProjectID = ""

	if !reflect.DeepEqual(newStatus, status) {
		raw, err := util.ToRaw(newStatus)
		if err != nil {
			return errors.Errorf("generate raw data of cluster provider status for cluster %v/%v: %w",
				cluster.Namespace, cluster.Name, err)
		}
		newCluster := cluster.DeepCopy()
		newCluster.Status.ProviderStatus = &runtime.RawExtension{Raw: raw}

		err = a.client.Status().Update(context.TODO(), newCluster)
		if err != nil {
			return errors.Errorf("update status of cluster %v/%v", cluster.Namespace, cluster.Name, err)
		}
	}
	return nil
}
