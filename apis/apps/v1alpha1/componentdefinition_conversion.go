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

package v1alpha1

import (
	appsv1 "github.com/apecloud/kubeblocks/apis/apps/v1"
	"github.com/jinzhu/copier"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this ComponentDefinition to the Hub version (v1).
func (r *ComponentDefinition) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*appsv1.ComponentDefinition)

	// objectMeta
	dst.ObjectMeta = r.ObjectMeta

	// spec
	copier.Copy(&dst.Spec, &r.Spec) // TODO(v1.0): changed fields

	// status
	copier.Copy(&dst.Status, &r.Status)

	return nil
}

// ConvertFrom converts from the Hub version (v1) to this version.
func (r *ComponentDefinition) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*appsv1.ComponentDefinition)

	// objectMeta
	r.ObjectMeta = src.ObjectMeta

	// spec
	copier.Copy(&r.Spec, &src.Spec) // TODO(v1.0): changed fields

	// status
	copier.Copy(&r.Status, &src.Status)

	return nil
}
