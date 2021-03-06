/*
 * Keyturner api
 *
 * Keyturner api
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package api

import (
	"errors"
)

// OfficialApiService is a service that implents the logic for the OfficialApiServicer
// This service should implement the business logic for every endpoint for the OfficialApi API. 
// Include any external packages or services that will be required by this service.
type OfficialApiService struct {
}

// NewOfficialApiService creates a default api service
func NewOfficialApiService() OfficialApiServicer {
	return &OfficialApiService{}
}

// CallbackAddGet - Registers a new callback url
func (s *OfficialApiService) CallbackAddGet(url string) (interface{}, error) {
	// TODO - update CallbackAddGet with the required logic for this service method.
	// Add api_official_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.
	return nil, errors.New("service method 'CallbackAddGet' not implemented")
}

// CallbackListGet - Returns all registered url callbacks
func (s *OfficialApiService) CallbackListGet() (interface{}, error) {
	// TODO - update CallbackListGet with the required logic for this service method.
	// Add api_official_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.
	return nil, errors.New("service method 'CallbackListGet' not implemented")
}

// CallbackRemoveGet - Removes a previously added callback
func (s *OfficialApiService) CallbackRemoveGet(id string) (interface{}, error) {
	// TODO - update CallbackRemoveGet with the required logic for this service method.
	// Add api_official_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.
	return nil, errors.New("service method 'CallbackRemoveGet' not implemented")
}

// ListGet - 
func (s *OfficialApiService) ListGet() (interface{}, error) {
	// TODO - update ListGet with the required logic for this service method.
	// Add api_official_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.
	return nil, errors.New("service method 'ListGet' not implemented")
}

// LockActionGet - Performs a lock operation on the given Smart Lock
func (s *OfficialApiService) LockActionGet(nukiId string, action string, noWait string) (interface{}, error) {
	// TODO - update LockActionGet with the required logic for this service method.
	// Add api_official_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.
	return nil, errors.New("service method 'LockActionGet' not implemented")
}

// LockStateGet - 
func (s *OfficialApiService) LockStateGet(nukiId string) (interface{}, error) {
	// TODO - update LockStateGet with the required logic for this service method.
	// Add api_official_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.
	return nil, errors.New("service method 'LockStateGet' not implemented")
}
