/*
 * Keyturner api
 *
 * Keyturner api
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package api

type LastLockState struct {

	State int32 `json:"state"`

	StateName string `json:"stateName"`

	BatteryCritical bool `json:"batteryCritical"`

	Timestamp string `json:"timestamp"`
}
