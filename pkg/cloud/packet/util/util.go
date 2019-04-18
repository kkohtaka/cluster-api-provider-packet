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
	errors "golang.org/x/xerrors"

	"k8s.io/apimachinery/pkg/util/json"

	clusterv1 "sigs.k8s.io/cluster-api/pkg/apis/cluster/v1alpha1"
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
