/*
Deepfence ThreatMapper

Deepfence Runtime API provides programmatic control over Deepfence microservice securing your container, kubernetes and cloud deployments. The API abstracts away underlying infrastructure details like cloud provider,  container distros, container orchestrator and type of deployment. This is one uniform API to manage and control security alerts, policies and response to alerts for microservices running anywhere i.e. managed pure greenfield container deployments or a mix of containers, VMs and serverless paradigms like AWS Fargate.

API version: 2.0.0
Contact: community@deepfence.io
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package deepfence_server_client

import (
	"encoding/json"
)

// checks if the ModelScanInfo type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ModelScanInfo{}

// ModelScanInfo struct for ModelScanInfo
type ModelScanInfo struct {
	ScanId string `json:"scan_id"`
	Status string `json:"status"`
	UpdatedAt int32 `json:"updated_at"`
}

// NewModelScanInfo instantiates a new ModelScanInfo object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewModelScanInfo(scanId string, status string, updatedAt int32) *ModelScanInfo {
	this := ModelScanInfo{}
	this.ScanId = scanId
	this.Status = status
	this.UpdatedAt = updatedAt
	return &this
}

// NewModelScanInfoWithDefaults instantiates a new ModelScanInfo object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewModelScanInfoWithDefaults() *ModelScanInfo {
	this := ModelScanInfo{}
	return &this
}

// GetScanId returns the ScanId field value
func (o *ModelScanInfo) GetScanId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.ScanId
}

// GetScanIdOk returns a tuple with the ScanId field value
// and a boolean to check if the value has been set.
func (o *ModelScanInfo) GetScanIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.ScanId, true
}

// SetScanId sets field value
func (o *ModelScanInfo) SetScanId(v string) {
	o.ScanId = v
}

// GetStatus returns the Status field value
func (o *ModelScanInfo) GetStatus() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Status
}

// GetStatusOk returns a tuple with the Status field value
// and a boolean to check if the value has been set.
func (o *ModelScanInfo) GetStatusOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Status, true
}

// SetStatus sets field value
func (o *ModelScanInfo) SetStatus(v string) {
	o.Status = v
}

// GetUpdatedAt returns the UpdatedAt field value
func (o *ModelScanInfo) GetUpdatedAt() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.UpdatedAt
}

// GetUpdatedAtOk returns a tuple with the UpdatedAt field value
// and a boolean to check if the value has been set.
func (o *ModelScanInfo) GetUpdatedAtOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.UpdatedAt, true
}

// SetUpdatedAt sets field value
func (o *ModelScanInfo) SetUpdatedAt(v int32) {
	o.UpdatedAt = v
}

func (o ModelScanInfo) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o ModelScanInfo) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["scan_id"] = o.ScanId
	toSerialize["status"] = o.Status
	toSerialize["updated_at"] = o.UpdatedAt
	return toSerialize, nil
}

type NullableModelScanInfo struct {
	value *ModelScanInfo
	isSet bool
}

func (v NullableModelScanInfo) Get() *ModelScanInfo {
	return v.value
}

func (v *NullableModelScanInfo) Set(val *ModelScanInfo) {
	v.value = val
	v.isSet = true
}

func (v NullableModelScanInfo) IsSet() bool {
	return v.isSet
}

func (v *NullableModelScanInfo) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableModelScanInfo(val *ModelScanInfo) *NullableModelScanInfo {
	return &NullableModelScanInfo{value: val, isSet: true}
}

func (v NullableModelScanInfo) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableModelScanInfo) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

