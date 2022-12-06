/*
Copyright 2022 The Crossplane Authors.

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

package v1alpha1

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// EngineParameters are the configurable fields of a Engine.
type EngineParameters struct {
	ConfigurableField string `json:"configurableField"`
}

// EngineObservation are the observable fields of a Engine.
type EngineObservation struct {
	ObservableField string `json:"observableField,omitempty"`
}

// A EngineSpec defines the desired state of a Engine.
type EngineSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       EngineParameters `json:"forProvider"`
}

// A EngineStatus represents the observed state of a Engine.
type EngineStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          EngineObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Engine is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,secret}
type Engine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EngineSpec   `json:"spec"`
	Status EngineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// EngineList contains a list of Engine
type EngineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Engine `json:"items"`
}

// Engine type metadata.
var (
	EngineKind             = reflect.TypeOf(Engine{}).Name()
	EngineGroupKind        = schema.GroupKind{Group: Group, Kind: EngineKind}.String()
	EngineKindAPIVersion   = EngineKind + "." + SchemeGroupVersion.String()
	EngineGroupVersionKind = SchemeGroupVersion.WithKind(EngineKind)
)

func init() {
	SchemeBuilder.Register(&Engine{}, &EngineList{})
}
