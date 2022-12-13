package v1alpha1

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// SecretPathParameters are the configurable fields of a SecretPath.
type SecretPathParameters struct {
	Path   string `json:"path"`
	Engine string `json:"engine"`
}

// SecretPathObservation are the observable fields of a SecretPath.
type SecretPathObservation struct {
	ObservableField string `json:"observableField,omitempty"`
}

// A SecretPathSpec defines the desired state of a SecretPath.
type SecretPathSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       SecretPathParameters `json:"forProvider"`
}

// A SecretPathStatus represents the observed state of a SecretPath.
type SecretPathStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          SecretPathObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A SecretPath is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,secret}
type SecretPath struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecretPathSpec   `json:"spec"`
	Status SecretPathStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SecretPathList contains a list of SecretPath
type SecretPathList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecretPath `json:"items"`
}

// SecretPath type metadata.
var (
	SecretPathKind             = reflect.TypeOf(SecretPath{}).Name()
	SecretPathGroupKind        = schema.GroupKind{Group: Group, Kind: SecretPathKind}.String()
	SecretPathKindAPIVersion   = SecretPathKind + "." + SchemeGroupVersion.String()
	SecretPathGroupVersionKind = SchemeGroupVersion.WithKind(SecretPathKind)
)

func init() {
	SchemeBuilder.Register(&SecretPath{}, &SecretPathList{})
}
