// +build !ignore_autogenerated

// Code generated by operator-sdk. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ChaosPod) DeepCopyInto(out *ChaosPod) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ChaosPod.
func (in *ChaosPod) DeepCopy() *ChaosPod {
	if in == nil {
		return nil
	}
	out := new(ChaosPod)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ChaosPod) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ChaosPodList) DeepCopyInto(out *ChaosPodList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ChaosPod, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ChaosPodList.
func (in *ChaosPodList) DeepCopy() *ChaosPodList {
	if in == nil {
		return nil
	}
	out := new(ChaosPodList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ChaosPodList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ChaosPodSpec) DeepCopyInto(out *ChaosPodSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ChaosPodSpec.
func (in *ChaosPodSpec) DeepCopy() *ChaosPodSpec {
	if in == nil {
		return nil
	}
	out := new(ChaosPodSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ChaosPodStatus) DeepCopyInto(out *ChaosPodStatus) {
	*out = *in
	if in.KilledPodNames != nil {
		in, out := &in.KilledPodNames, &out.KilledPodNames
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ChaosPodStatus.
func (in *ChaosPodStatus) DeepCopy() *ChaosPodStatus {
	if in == nil {
		return nil
	}
	out := new(ChaosPodStatus)
	in.DeepCopyInto(out)
	return out
}
