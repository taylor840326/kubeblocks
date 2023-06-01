/*
Copyright (C) 2022-2023 ApeCloud Co., Ltd

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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ComponentClassDefinitionSpec defines the desired state of ComponentClassDefinition
type ComponentClassDefinitionSpec struct {
	// group defines a list of class series that conform to the same constraint.
	// +optional
	Groups []ComponentClassGroup `json:"groups,omitempty"`
}

type ComponentClassGroup struct {
	// resourceConstraintRef reference to the resource constraint object name, indicates that the series
	// defined below all conform to the constraint.
	// +kubebuilder:validation:Required
	ResourceConstraintRef string `json:"resourceConstraintRef"`

	// template is a class definition template that uses the Go template syntax and allows for variable declaration.
	// When defining a class in Series, specifying the variable's value is sufficient, as the complete class
	// definition will be generated through rendering the template.
	//
	// For example:
	//	template: |
	//	  cpu: "{{ or .cpu 1 }}"
	//	  memory: "{{ or .memory 4 }}Gi"
	//
	// +optional
	Template string `json:"template,omitempty"`

	// vars defines the variables declared in the template and will be used to generating the complete class definition by
	// render the template.
	// +listType=set
	// +optional
	Vars []string `json:"vars,omitempty"`

	// series is a series of class definitions.
	// +optional
	Series []ComponentClassSeries `json:"series,omitempty"`
}

type ComponentClassSeries struct {
	// namingTemplate is a template that uses the Go template syntax and allows for referencing variables defined
	// in ComponentClassGroup.Template. This enables dynamic generation of class names.
	// For example:
	// name: "general-{{ .cpu }}c{{ .memory }}g"
	// +optional
	NamingTemplate string `json:"namingTemplate,omitempty"`

	// classes are definitions of classes that come in two forms. In the first form, only ComponentClass.Args
	// need to be defined, and the complete class definition is generated by rendering the ComponentClassGroup.Template
	// and Name. In the second form, the Name, CPU and Memory must be defined.
	// +optional
	Classes []ComponentClass `json:"classes,omitempty"`
}

type ComponentClass struct {
	// name is the class name
	// +optional
	Name string `json:"name,omitempty"`

	// args are variable's value
	// +optional
	Args []string `json:"args,omitempty"`

	// the CPU of the class
	// +optional
	CPU resource.Quantity `json:"cpu,omitempty"`

	// the memory of the class
	// +optional
	Memory resource.Quantity `json:"memory,omitempty"`
}

// ComponentClassDefinitionStatus defines the observed state of ComponentClassDefinition
type ComponentClassDefinitionStatus struct {
	// observedGeneration is the most recent generation observed for this
	// ComponentClassDefinition. It corresponds to the ComponentClassDefinition's generation, which is
	// updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// classes is the list of classes that have been observed for this ComponentClassDefinition
	Classes []ComponentClassInstance `json:"classes,omitempty"`
}

type ComponentClassInstance struct {
	ComponentClass `json:",inline"`

	// resourceConstraintRef reference to the resource constraint object name.
	ResourceConstraintRef string `json:"resourceConstraintRef,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories={kubeblocks},scope=Cluster,shortName=ccd

// ComponentClassDefinition is the Schema for the componentclassdefinitions API
type ComponentClassDefinition struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ComponentClassDefinitionSpec   `json:"spec,omitempty"`
	Status ComponentClassDefinitionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ComponentClassDefinitionList contains a list of ComponentClassDefinition
type ComponentClassDefinitionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ComponentClassDefinition `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ComponentClassDefinition{}, &ComponentClassDefinitionList{})
}

func (r *ComponentClass) ToResourceRequirements() corev1.ResourceRequirements {
	requests := corev1.ResourceList{
		corev1.ResourceCPU:    r.CPU,
		corev1.ResourceMemory: r.Memory,
	}
	return corev1.ResourceRequirements{Requests: requests, Limits: requests}
}
