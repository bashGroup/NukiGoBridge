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
	"net/http"
	"strings"
)

// A OfficialApiController binds http requests to an api service and writes the service results to the http response
type OfficialApiController struct {
	service OfficialApiServicer
}

// NewOfficialApiController creates a default api controller
func NewOfficialApiController(s OfficialApiServicer) Router {
	return &OfficialApiController{service: s}
}

// Routes returns all of the api route for the OfficialApiController
func (c *OfficialApiController) Routes() Routes {
	return Routes{
		{
			"CallbackAddGet",
			strings.ToUpper("Get"),
			"/api/v1/callback/add",
			c.CallbackAddGet,
		},
		{
			"CallbackListGet",
			strings.ToUpper("Get"),
			"/api/v1/callback/list",
			c.CallbackListGet,
		},
		{
			"CallbackRemoveGet",
			strings.ToUpper("Get"),
			"/api/v1/callback/remove",
			c.CallbackRemoveGet,
		},
		{
			"ListGet",
			strings.ToUpper("Get"),
			"/api/v1/list",
			c.ListGet,
		},
		{
			"LockActionGet",
			strings.ToUpper("Get"),
			"/api/v1/lockAction",
			c.LockActionGet,
		},
		{
			"LockStateGet",
			strings.ToUpper("Get"),
			"/api/v1/lockState",
			c.LockStateGet,
		},
	}
}

// CallbackAddGet - Registers a new callback url
func (c *OfficialApiController) CallbackAddGet(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	url := query.Get("url")
	result, err := c.service.CallbackAddGet(url)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	EncodeJSONResponse(result, nil, w)
}

// CallbackListGet - Returns all registered url callbacks
func (c *OfficialApiController) CallbackListGet(w http.ResponseWriter, r *http.Request) {
	result, err := c.service.CallbackListGet()
	if err != nil {
		w.WriteHeader(500)
		return
	}

	EncodeJSONResponse(result, nil, w)
}

// CallbackRemoveGet - Removes a previously added callback
func (c *OfficialApiController) CallbackRemoveGet(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	id := query.Get("id")
	result, err := c.service.CallbackRemoveGet(id)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	EncodeJSONResponse(result, nil, w)
}

// ListGet -
func (c *OfficialApiController) ListGet(w http.ResponseWriter, r *http.Request) {
	result, err := c.service.ListGet()
	if err != nil {
		w.WriteHeader(500)
		return
	}

	EncodeJSONResponse(result, nil, w)
}

// LockActionGet - Performs a lock operation on the given Smart Lock
func (c *OfficialApiController) LockActionGet(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	nukiId := query.Get("nukiId")
	action := query.Get("action")
	noWait := query.Get("noWait")
	result, err := c.service.LockActionGet(nukiId, action, noWait)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	EncodeJSONResponse(result, nil, w)
}

// LockStateGet -
func (c *OfficialApiController) LockStateGet(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	nukiId := query.Get("nukiId")
	result, err := c.service.LockStateGet(nukiId)
	if err != nil {
		w.WriteHeader(500)
		return
	}

	EncodeJSONResponse(result, nil, w)
}
