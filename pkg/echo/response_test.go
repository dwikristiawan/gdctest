package echo_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gdctest/pkg/apperror"
	pkgEcho "gdctest/pkg/echo"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRespond_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := pkgEcho.Respond(c, map[string]string{"foo": "bar"}, nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"message":"Successful"`)
	assert.Contains(t, rec.Body.String(), `"foo":"bar"`)
}

func TestRespond_ClientError_4xxExposesDetail(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	clientErr := apperror.New(http.StatusBadRequest, "email is invalid")
	err := pkgEcho.Respond(c, nil, clientErr)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	// 4xx error message harus detail agar client tahu salahnya di mana
	assert.Contains(t, rec.Body.String(), `"status":false`)
	assert.Contains(t, rec.Body.String(), `"code":"BAD_REQUEST"`)
	assert.Contains(t, rec.Body.String(), `"timestamp":`)
	assert.Contains(t, rec.Body.String(), "email is invalid")
}

func TestRespond_InternalError_500MasksDetailInResponse(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Detail internal error (misal koneksi database bocor)
	internalErr := apperror.New(http.StatusInternalServerError, "failed to query user: pq: password authentication failed for user postgres")
	err := pkgEcho.Respond(c, nil, internalErr)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	// RESPONSE HARUS MASKING: menampilkan general error, BUKAN database credential/leak
	assert.Contains(t, rec.Body.String(), `"message":"Internal Server Error. General error"`)
	assert.Contains(t, rec.Body.String(), `"error":"General error"`)
	assert.NotContains(t, rec.Body.String(), "pq: password authentication failed")
}

func TestRespond_GenericUnwrappedError_DefaultsTo500AndMasks(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Raw unexpected Go error
	genericErr := errors.New("unexpected nil pointer dereference at runtime")
	err := pkgEcho.Respond(c, nil, genericErr)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	assert.Contains(t, rec.Body.String(), `"message":"Internal Server Error. General error"`)
	assert.NotContains(t, rec.Body.String(), "nil pointer dereference")
}
