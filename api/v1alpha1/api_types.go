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
	Daily *int32 `json:"daily,omitempty"`
	// +optional
	Monthly *int32 `json:"monthly,omitempty"`
	// +optional
	Eternity *int32 `json:"eternity,omitempty"`
}

type AuthRateLimitPlan struct {
	// +optional
	Global *RateLimitPlan `json:"global,omitempty"`
	// +optional
	Operations map[string]*RateLimitPlan `json:"operations,omitempty"`
}

func (a *AuthRateLimitPlan) GetGlobal() *RateLimitPlan {
	if a.Global == nil {
		a.Global = &RateLimitPlan{}
	}
	return a.Global
}

func (a *AuthRateLimitPlan) GetOperation(operationID string) *RateLimitPlan {
	if a.Operations == nil {
		a.Operations = map[string]*RateLimitPlan{}
	}

	if _, ok := a.Operations[operationID]; !ok {
		a.Operations[operationID] = &RateLimitPlan{}
	}

	return a.Operations[operationID]
}

type UnAuthRateLimitPlan struct {
	// +optional
	Global *RateLimitPlan `json:"global,omitempty"`
	// +optional
	RemoteIP *RateLimitPlan `json:"remoteIP,omitempty"`
	// +optional
	Operations map[string]*RateLimitPlan `json:"operations,omitempty"`
}

func (u *UnAuthRateLimitPlan) GetGlobal() *RateLimitPlan {
	if u.Global == nil {
		u.Global = &RateLimitPlan{}
	}
	return u.Global
}

func (u *UnAuthRateLimitPlan) GetRemoteIP() *RateLimitPlan {
	if u.RemoteIP == nil {
		u.RemoteIP = &RateLimitPlan{}
	}
	return u.RemoteIP
}

func (u *UnAuthRateLimitPlan) GetOperation(operationID string) *RateLimitPlan {
	if u.Operations == nil {
		u.Operations = map[string]*RateLimitPlan{}
	}

	if _, ok := u.Operations[operationID]; !ok {
		u.Operations[operationID] = &RateLimitPlan{}
	}

	return u.Operations[operationID]
}

type ApiPlan struct {
	Description string `json:"description"`
	// +optional
	Auth *AuthRateLimitPlan `json:"auth,omitempty"`
	// +optional
	UnAuth *UnAuthRateLimitPlan `json:"unauth,omitempty"`
}

func (a *ApiPlan) GetAuth() *AuthRateLimitPlan {
	if a.Auth == nil {
		a.Auth = &AuthRateLimitPlan{}
	}
	return a.Auth
}

func (a *ApiPlan) GetUnAuth() *UnAuthRateLimitPlan {
	if a.UnAuth == nil {
		a.UnAuth = &UnAuthRateLimitPlan{}
	}
	return a.UnAuth
}

func (a *ApiPlan) SetUnAuthGlobalDaily(val int32) {
	a.GetUnAuth().GetGlobal().Daily = &val
}

func (a *ApiPlan) SetUnAuthGlobalMonthly(val int32) {
	a.GetUnAuth().GetGlobal().Monthly = &val
}

func (a *ApiPlan) SetUnAuthGlobalEternity(val int32) {
	a.GetUnAuth().GetGlobal().Eternity = &val
}

func (a *ApiPlan) SetUnAuthRemoteIPDaily(val int32) {
	a.GetUnAuth().GetRemoteIP().Daily = &val
}

func (a *ApiPlan) SetUnAuthRemoteIPMonthly(val int32) {
	a.GetUnAuth().GetRemoteIP().Monthly = &val
}

func (a *ApiPlan) SetUnAuthRemoteIPEternity(val int32) {
	a.GetUnAuth().GetRemoteIP().Eternity = &val
}

func (a *ApiPlan) SetUnAuthOperationEternity(val int32, operationID string) {
	a.GetUnAuth().GetOperation(operationID).Eternity = &val
}

func (a *ApiPlan) SetUnAuthOperationDaily(val int32, operationID string) {
	a.GetUnAuth().GetOperation(operationID).Daily = &val
}

func (a *ApiPlan) SetUnAuthOperationMonthly(val int32, operationID string) {
	a.GetUnAuth().GetOperation(operationID).Monthly = &val
}

func (a *ApiPlan) SetAuthGlobalDaily(val int32) {
	a.GetAuth().GetGlobal().Daily = &val
}

func (a *ApiPlan) SetAuthGlobalMonthly(val int32) {
	a.GetAuth().GetGlobal().Monthly = &val
}

func (a *ApiPlan) SetAuthGlobalEternity(val int32) {
	a.GetAuth().GetGlobal().Eternity = &val
}

func (a *ApiPlan) SetAuthOperationDaily(val int32, operationID string) {
	a.GetAuth().GetOperation(operationID).Daily = &val
}

func (a *ApiPlan) SetAuthOperationMonthly(val int32, operationID string) {
	a.GetAuth().GetOperation(operationID).Monthly = &val
}

func (a *ApiPlan) SetAuthOperationEternity(val int32, operationID string) {
	a.GetAuth().GetOperation(operationID).Eternity = &val
}

// ApiSpec defines the desired state of Api
type ApiSpec struct {
	Description   string `json:"description"`
	PublicDomain  string `json:"publicdomain"`
	OAS           string `json:"oas"`
	PathMatchType string `json:"pathmatchtype"`
	ServiceName   string `json:"servicename"`

	// +optional
	Plans map[string]*ApiPlan `json:"plans,omitempty"`
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
