//go:build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConnectionCredentialAuth) DeepCopyInto(out *ConnectionCredentialAuth) {
	*out = *in
	if in.Username != nil {
		in, out := &in.Username, &out.Username
		*out = new(CredentialVar)
		(*in).DeepCopyInto(*out)
	}
	if in.Password != nil {
		in, out := &in.Password, &out.Password
		*out = new(CredentialVar)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConnectionCredentialAuth.
func (in *ConnectionCredentialAuth) DeepCopy() *ConnectionCredentialAuth {
	if in == nil {
		return nil
	}
	out := new(ConnectionCredentialAuth)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CredentialVar) DeepCopyInto(out *CredentialVar) {
	*out = *in
	if in.ValueFrom != nil {
		in, out := &in.ValueFrom, &out.ValueFrom
		*out = new(corev1.EnvVarSource)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CredentialVar.
func (in *CredentialVar) DeepCopy() *CredentialVar {
	if in == nil {
		return nil
	}
	out := new(CredentialVar)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceDescriptor) DeepCopyInto(out *ServiceDescriptor) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceDescriptor.
func (in *ServiceDescriptor) DeepCopy() *ServiceDescriptor {
	if in == nil {
		return nil
	}
	out := new(ServiceDescriptor)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServiceDescriptor) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceDescriptorList) DeepCopyInto(out *ServiceDescriptorList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ServiceDescriptor, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceDescriptorList.
func (in *ServiceDescriptorList) DeepCopy() *ServiceDescriptorList {
	if in == nil {
		return nil
	}
	out := new(ServiceDescriptorList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServiceDescriptorList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceDescriptorSpec) DeepCopyInto(out *ServiceDescriptorSpec) {
	*out = *in
	if in.Endpoint != nil {
		in, out := &in.Endpoint, &out.Endpoint
		*out = new(CredentialVar)
		(*in).DeepCopyInto(*out)
	}
	if in.Host != nil {
		in, out := &in.Host, &out.Host
		*out = new(CredentialVar)
		(*in).DeepCopyInto(*out)
	}
	if in.Port != nil {
		in, out := &in.Port, &out.Port
		*out = new(CredentialVar)
		(*in).DeepCopyInto(*out)
	}
	if in.Auth != nil {
		in, out := &in.Auth, &out.Auth
		*out = new(ConnectionCredentialAuth)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceDescriptorSpec.
func (in *ServiceDescriptorSpec) DeepCopy() *ServiceDescriptorSpec {
	if in == nil {
		return nil
	}
	out := new(ServiceDescriptorSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceDescriptorStatus) DeepCopyInto(out *ServiceDescriptorStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceDescriptorStatus.
func (in *ServiceDescriptorStatus) DeepCopy() *ServiceDescriptorStatus {
	if in == nil {
		return nil
	}
	out := new(ServiceDescriptorStatus)
	in.DeepCopyInto(out)
	return out
}
