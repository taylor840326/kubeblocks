/*
Copyright (C) 2022-2024 ApeCloud Co., Ltd

This file is part of KubeBlocks project

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var clusterdefinitionlog = logf.Log.WithName("clusterdefinition-resource")

func (r *ClusterDefinition) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-apps-kubeblocks-io-v1-clusterdefinition,mutating=true,failurePolicy=fail,sideEffects=None,groups=apps.kubeblocks.io,resources=clusterdefinitions,verbs=create;update,versions=v1,name=mclusterdefinition.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &ClusterDefinition{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ClusterDefinition) Default() {
	clusterdefinitionlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-apps-kubeblocks-io-v1-clusterdefinition,mutating=false,failurePolicy=fail,sideEffects=None,groups=apps.kubeblocks.io,resources=clusterdefinitions,verbs=create;update,versions=v1,name=vclusterdefinition.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &ClusterDefinition{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterDefinition) ValidateCreate() (warnings admission.Warnings, err error) {
	clusterdefinitionlog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil, nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterDefinition) ValidateUpdate(old runtime.Object) (warnings admission.Warnings, err error) {
	clusterdefinitionlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterDefinition) ValidateDelete() (warnings admission.Warnings, err error) {
	clusterdefinitionlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
