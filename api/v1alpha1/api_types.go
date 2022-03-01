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
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RateLimitConf struct {
	// +optional
	Daily *int32 `json:"daily,omitempty"`
	// +optional
	Monthly *int32 `json:"monthly,omitempty"`
	// +optional
	Eternity *int32 `json:"eternity,omitempty"`
}

func (r *RateLimitConf) IsEmpty() bool {
	return r.Daily == nil && r.Monthly == nil && r.Eternity == nil
}

type UnAuthRateLimitConf struct {
	// +optional
	Global *RateLimitConf `json:"global,omitempty"`
	// +optional
	RemoteIP *RateLimitConf `json:"remoteIP,omitempty"`
	// +optional
	Operations map[string]*RateLimitConf `json:"operations,omitempty"`
}

func (u *UnAuthRateLimitConf) GetGlobal() *RateLimitConf {
	if u.Global == nil {
		u.Global = &RateLimitConf{}
	}
	return u.Global
}

func (u *UnAuthRateLimitConf) GetRemoteIP() *RateLimitConf {
	if u.RemoteIP == nil {
		u.RemoteIP = &RateLimitConf{}
	}
	return u.RemoteIP
}

func (u *UnAuthRateLimitConf) GetOperation(operationID string) *RateLimitConf {
	if u.Operations == nil {
		u.Operations = map[string]*RateLimitConf{}
	}

	if _, ok := u.Operations[operationID]; !ok {
		u.Operations[operationID] = &RateLimitConf{}
	}

	return u.Operations[operationID]
}

type ApiPlan struct {
	Description string `json:"description"`
	// +optional
	Global *RateLimitConf `json:"global,omitempty"`
	// +optional
	Operations map[string]*RateLimitConf `json:"operations,omitempty"`
}

func (a *ApiPlan) GetGlobal() *RateLimitConf {
	if a.Global == nil {
		a.Global = &RateLimitConf{}
	}
	return a.Global
}

func (a *ApiPlan) GetOperation(operationID string) *RateLimitConf {
	if a.Operations == nil {
		a.Operations = map[string]*RateLimitConf{}
	}

	if _, ok := a.Operations[operationID]; !ok {
		a.Operations[operationID] = &RateLimitConf{}
	}

	return a.Operations[operationID]
}

func (a *ApiPlan) SetAuthGlobalDaily(val int32) {
	a.GetGlobal().Daily = &val
}

func (a *ApiPlan) SetAuthGlobalMonthly(val int32) {
	a.GetGlobal().Monthly = &val
}

func (a *ApiPlan) SetAuthGlobalEternity(val int32) {
	a.GetGlobal().Eternity = &val
}

func (a *ApiPlan) SetAuthOperationDaily(val int32, operationID string) {
	a.GetOperation(operationID).Daily = &val
}

func (a *ApiPlan) SetAuthOperationMonthly(val int32, operationID string) {
	a.GetOperation(operationID).Monthly = &val
}

func (a *ApiPlan) SetAuthOperationEternity(val int32, operationID string) {
	a.GetOperation(operationID).Eternity = &val
}

func (a *ApiPlan) IsEmpty() bool {
	for _, rateLimitConf := range a.Operations {
		if !rateLimitConf.IsEmpty() {
			return false
		}
	}

	return a.GetGlobal().IsEmpty()
}

type UserInfo struct {
	// +optional
	Plan *string `json:"plan,omitempty"`
	// +optional
	APIKey *string `json:"apiKey,omitempty"`
}

// ApiSpec defines the desired state of Api
type ApiSpec struct {
	Description   string `json:"description"`
	PublicDomain  string `json:"publicdomain"`
	OAS           string `json:"oas"`
	PathMatchType string `json:"pathmatchtype"`
	ServiceName   string `json:"servicename"`

	// +optional
	UnAuthRateLimit *UnAuthRateLimitConf `json:"unauthratelimit,omitempty"`
	// +optional
	Plans map[string]*ApiPlan `json:"plans,omitempty"`
	// UserPlan userID -> planID
	// +optional
	Users map[string]*UserInfo `json:"users,omitempty"`
}

func (a *ApiSpec) GetUnAuthRateLimit() *UnAuthRateLimitConf {
	if a.UnAuthRateLimit == nil {
		a.UnAuthRateLimit = &UnAuthRateLimitConf{}
	}
	return a.UnAuthRateLimit
}

func (a *ApiSpec) SetUnAuthGlobalDaily(val int32) {
	a.GetUnAuthRateLimit().GetGlobal().Daily = &val
}

func (a *ApiSpec) SetUnAuthGlobalMonthly(val int32) {
	a.GetUnAuthRateLimit().GetGlobal().Monthly = &val
}

func (a *ApiSpec) SetUnAuthGlobalEternity(val int32) {
	a.GetUnAuthRateLimit().GetGlobal().Eternity = &val
}

func (a *ApiSpec) SetUnAuthRemoteIPDaily(val int32) {
	a.GetUnAuthRateLimit().GetRemoteIP().Daily = &val
}

func (a *ApiSpec) SetUnAuthRemoteIPMonthly(val int32) {
	a.GetUnAuthRateLimit().GetRemoteIP().Monthly = &val
}

func (a *ApiSpec) SetUnAuthRemoteIPEternity(val int32) {
	a.GetUnAuthRateLimit().GetRemoteIP().Eternity = &val
}

func (a *ApiSpec) SetUnAuthOperationEternity(val int32, operationID string) {
	a.GetUnAuthRateLimit().GetOperation(operationID).Eternity = &val
}

func (a *ApiSpec) SetUnAuthOperationDaily(val int32, operationID string) {
	a.GetUnAuthRateLimit().GetOperation(operationID).Daily = &val
}

func (a *ApiSpec) SetUnAuthOperationMonthly(val int32, operationID string) {
	a.GetUnAuthRateLimit().GetOperation(operationID).Monthly = &val
}

func (a *ApiSpec) HasAnyRateLimitOnOperation(operationID string) bool {
	for _, userInfo := range a.Users {
		if userInfo.Plan != nil {
			apiPlan := a.Plans[*userInfo.Plan]
			if apiPlan == nil {
				panic(fmt.Sprintf("plan does not exist %s", *userInfo.Plan))
			}

			if !apiPlan.GetOperation(operationID).IsEmpty() {
				return true
			}
		}
	}

	return false
}
func (a *ApiSpec) HasAnyAuthRateLimit() bool {
	for _, userInfo := range a.Users {
		if userInfo.Plan != nil {
			apiPlan := a.Plans[*userInfo.Plan]
			if apiPlan == nil {
				panic(fmt.Sprintf("plan does not exist %s", *userInfo.Plan))
			}

			if !apiPlan.IsEmpty() {
				return true
			}
		}
	}
	return false
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
