package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MykindSpec defines the desired state of Mykind
type MykindSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ReplicasMySQL *int32          `json:"replicas_mysql"`
	ImageMySQL    string          `json:"image_mysql"`
	PortMySQL     int32           `json:"port_mysql"`
	NodeportMySQL int32           `json:"nodeport_mysql"`
	EnvsMySQL     []corev1.EnvVar `json:"envs_mysql,omitempty"`

	ReplicasCov *int32 `json:"replicas_cov"`
	ImageCov    string `json:"image_cov"`
	PortCov     int32  `json:"port_cov"`
	NodeportCov int32  `json:"nodeport_cov"`
}

// MykindStatus defines the observed state of Mykind
type MykindStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status string `json:"status"`
}

// +kubebuilder:subresource:status
// +kubebuilder:object:root=true

// Mykind is the Schema for the mykinds API
type Mykind struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MykindSpec   `json:"spec,omitempty"`
	Status MykindStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MykindList contains a list of Mykind
type MykindList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Mykind `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Mykind{}, &MykindList{})
}
