/*
Copyright 2022.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RateLimitPlan struct {
	// +optional
	Daily *int32 `json:"global,omitempty"`
	// +optional
	Monthly *int32 `json:"global,omitempty"`
	// +optional
	Eternity *int32 `json:"global,omitempty"`
}

type AuthRateLimitPlan struct {
	// +optional
	Global *RateLimitPlan `json:"global,omitempty"`
	// +optional
	Operations map[string]RateLimitPlan `json:"operations,omitempty"`
}

type UnAuthRateLimitPlan struct {
	// +optional
	Global *RateLimitPlan `json:"global,omitempty"`
	// +optional
	RemoteIP *RateLimitPlan `json:"remoteIP,omitempty"`
	// +optional
	Operations map[string]RateLimitPlan `json:"operations,omitempty"`
}

type ApiPlan struct {
	Description string `json:"description"`
	// +optional
	Auth *AuthRateLimitPlan `json:"auth,omitempty"`
	// +optional
	UnAuth *UnAuthRateLimitPlan `json:"unauth,omitempty"`
}

// ApiSpec defines the desired state of Api
type ApiSpec struct {
	Description   string `json:"description"`
	PublicDomain  string `json:"publicdomain"`
	OAS           string `json:"oas"`
	PathMatchType string `json:"pathmatchtype"`
	ServiceName   string `json:"servicename"`

	// +optional
	Plans map[string]ApiPlan `json:"plans,omitempty"`
	// +optional
	Gateway *string `json:"gateway,omitempty"`
}

// ApiStatus defines the observed state of Api
type ApiStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Api is the Schema for the apis API
type Api struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApiSpec   `json:"spec,omitempty"`
	Status ApiStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ApiList contains a list of Api
type ApiList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Api `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Api{}, &ApiList{})
}
