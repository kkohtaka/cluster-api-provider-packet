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

package util

import (
	"context"
	"reflect"

	errors "golang.org/x/xerrors"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/util/retry"

	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	packetv1alpha1 "github.com/kkohtaka/cluster-api-provider-packet/pkg/apis/packet/v1alpha1"
)

func ToClusterProviderSpec(src *clusterv1.ClusterSpec) (*packetv1alpha1.PacketClusterProviderSpec, error) {
	var dst packetv1alpha1.PacketClusterProviderSpec
	if pspec := src.ProviderSpec; pspec.Value != nil {
		if err := yaml.Unmarshal(pspec.Value.Raw, &dst); err != nil {
			return nil, errors.Errorf("unmarshal cluster provider spec: %w", err)
		}
	}
	return &dst, nil
}

func ToClusterProviderStatus(src *clusterv1.ClusterStatus) (*packetv1alpha1.PacketClusterProviderStatus, error) {
	var dst packetv1alpha1.PacketClusterProviderStatus
	if pstatus := src.ProviderStatus; pstatus != nil {
		if err := yaml.Unmarshal(pstatus.Raw, &dst); err != nil {
			return nil, errors.Errorf("unmarshal cluster provider status: %w", err)
		}
	}
	return &dst, nil
}

func ToMachineProviderSpec(src *clusterv1.MachineSpec) (*packetv1alpha1.PacketMachineProviderSpec, error) {
	var dst packetv1alpha1.PacketMachineProviderSpec
	if pspec := src.ProviderSpec; pspec.Value != nil {
		if err := yaml.Unmarshal(pspec.Value.Raw, &dst); err != nil {
			return nil, errors.Errorf("unmarshal machine provider spec: %w", err)
		}
	}
	return &dst, nil
}

func ToMachineProviderStatus(src *clusterv1.MachineStatus) (*packetv1alpha1.PacketMachineProviderStatus, error) {
	var dst packetv1alpha1.PacketMachineProviderStatus
	if pstatus := src.ProviderStatus; pstatus != nil {
		if err := yaml.Unmarshal(pstatus.Raw, &dst); err != nil {
			return nil, errors.Errorf("unmarshal machine provider status: %w", err)
		}
	}
	return &dst, nil
}

func ToRaw(src interface{}) ([]byte, error) {
	data, err := json.Marshal(src)
	if err != nil {
		return nil, errors.Errorf("marshal to raw data: %w", err)
	}
	return data, nil
}

func UpdateClusterProviderSpec(
	c client.Client,
	clusterKey types.NamespacedName,
	newSpec *packetv1alpha1.PacketClusterProviderSpec,
) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		var cluster clusterv1.Cluster
		if err := c.Get(context.TODO(), clusterKey, &cluster); err != nil {
			return errors.Errorf("get latest cluster %v: %w", clusterKey, err)
		}

		raw, err := ToRaw(newSpec)
		if err != nil {
			return errors.Errorf("generate raw data of cluster provider spec for cluster %v: %w", clusterKey, err)
		}
		newCluster := cluster.DeepCopy()
		newCluster.Spec.ProviderSpec.Value = &runtime.RawExtension{Raw: raw}

		if reflect.DeepEqual(newCluster, cluster) {
			return nil
		}
		if err := c.Update(context.TODO(), newCluster); err != nil {
			return errors.Errorf("update cluster %v: %w", clusterKey, err)
		}
		return nil
	})
}

func UpdateClusterProviderStatus(
	c client.Client,
	clusterKey types.NamespacedName,
	newStatus *packetv1alpha1.PacketClusterProviderStatus,
) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		var cluster clusterv1.Cluster
		if err := c.Get(context.TODO(), clusterKey, &cluster); err != nil {
			return errors.Errorf("get latest cluster %v: %w", clusterKey, err)
		}

		raw, err := ToRaw(newStatus)
		if err != nil {
			return errors.Errorf("generate raw data of cluster provider status for cluster %v: %w", clusterKey, err)
		}
		newCluster := cluster.DeepCopy()
		newCluster.Status.ProviderStatus = &runtime.RawExtension{Raw: raw}

		if reflect.DeepEqual(newCluster, cluster) {
			return nil
		}
		if err := c.Status().Update(context.TODO(), newCluster); err != nil {
			return errors.Errorf("update cluster %v: %w", clusterKey, err)
		}
		return nil
	})
}

func UpdateMachineProviderSpec(
	c client.Client,
	machineKey types.NamespacedName,
	newSpec *packetv1alpha1.PacketMachineProviderSpec,
) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		var machine clusterv1.Machine
		if err := c.Get(context.TODO(), machineKey, &machine); err != nil {
			return errors.Errorf("get latest machine %v: %w", machineKey, err)
		}

		raw, err := ToRaw(newSpec)
		if err != nil {
			return errors.Errorf("generate raw data of machine provider spec for machine %v: %w", machineKey, err)
		}
		newMachine := machine.DeepCopy()
		newMachine.Spec.ProviderSpec.Value = &runtime.RawExtension{Raw: raw}

		if reflect.DeepEqual(newMachine, machine) {
			return nil
		}
		if err := c.Update(context.TODO(), newMachine); err != nil {
			return errors.Errorf("update machine %v: %w", machineKey, err)
		}
		return nil
	})
}

func UpdateMachineProviderStatus(
	c client.Client,
	machineKey types.NamespacedName,
	newStatus *packetv1alpha1.PacketMachineProviderStatus,
) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		var machine clusterv1.Machine
		if err := c.Get(context.TODO(), machineKey, &machine); err != nil {
			return errors.Errorf("get latest machine %v: %w", machineKey, err)
		}

		raw, err := ToRaw(newStatus)
		if err != nil {
			return errors.Errorf("generate raw data of machine provider status for machine %v: %w", machineKey, err)
		}
		newMachine := machine.DeepCopy()
		newMachine.Status.ProviderStatus = &runtime.RawExtension{Raw: raw}

		if reflect.DeepEqual(newMachine, machine) {
			return nil
		}
		if err := c.Status().Update(context.TODO(), newMachine); err != nil {
			return errors.Errorf("update machine %v: %w", machineKey, err)
		}
		return nil
	})
}
