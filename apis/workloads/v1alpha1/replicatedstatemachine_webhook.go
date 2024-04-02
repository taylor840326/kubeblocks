/*
Copyright (C) 2022-2024 ApeCloud Co., Ltd

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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var replicatedstatemachinelog = logf.Log.WithName("replicatedstatemachine-resource")

func (r *ReplicatedStateMachine) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-workloads-kubeblocks-io-v1alpha1-replicatedstatemachine,mutating=true,failurePolicy=fail,sideEffects=None,groups=workloads.kubeblocks.io,resources=replicatedstatemachines,verbs=create;update,versions=v1alpha1,name=mreplicatedstatemachine.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &ReplicatedStateMachine{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ReplicatedStateMachine) Default() {
	replicatedstatemachinelog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-workloads-kubeblocks-io-v1alpha1-replicatedstatemachine,mutating=false,failurePolicy=fail,sideEffects=None,groups=workloads.kubeblocks.io,resources=replicatedstatemachines,verbs=create;update,versions=v1alpha1,name=vreplicatedstatemachine.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &ReplicatedStateMachine{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ReplicatedStateMachine) ValidateCreate() (admission.Warnings, error) {
	replicatedstatemachinelog.Info("validate create", "name", r.Name)

	return nil, r.validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ReplicatedStateMachine) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	replicatedstatemachinelog.Info("validate update", "name", r.Name)

	return nil, r.validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ReplicatedStateMachine) ValidateDelete() (admission.Warnings, error) {
	replicatedstatemachinelog.Info("validate delete", "name", r.Name)

	return nil, r.validate()
}

func (r *ReplicatedStateMachine) validate() error {
	var allErrs field.ErrorList

	if len(r.Spec.Roles) > 0 {
		// Leader is required
		hasHeader := false
		for _, role := range r.Spec.Roles {
			if role.IsLeader && len(role.Name) > 0 {
				hasHeader = true
				break
			}
		}
		if !hasHeader {
			allErrs = append(allErrs,
				field.Required(field.NewPath("spec.roles"),
					"leader is required"))
		}
	}

	// servicePort must provide if spec.service is not nil
	if r.Spec.Service != nil && len(r.Spec.Service.Spec.Ports) == 0 {
		allErrs = append(allErrs,
			field.Required(field.NewPath("spec.service.ports"),
				"servicePort must provide"))
	}

	if len(allErrs) > 0 {
		return apierrors.NewInvalid(
			schema.GroupKind{
				Group: "workloads.kubeblocks.io/v1alpha1",
				Kind:  "ReplicatedStateMachine",
			},
			r.Name, allErrs)
	}

	return nil
}
