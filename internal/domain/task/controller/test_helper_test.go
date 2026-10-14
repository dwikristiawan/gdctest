package controller_test

import (
	"bytes"
	"net/http/httptest"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type echoValidator struct {
	v *validator.Validate
}

func (ev *echoValidator) Validate(i interface{}) error {
	return ev.v.Struct(i)
}

func setupEchoTest(method, target string, body []byte) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	e.Validator = &echoValidator{v: validator.New()}
	req := httptest.NewRequest(method, target, bytes.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}
