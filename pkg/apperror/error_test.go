package apperror

import (
	"errors"
	"net/http"
	"testing"
)

func TestAppError_Formatting(t *testing.T) {
	// 1. Text bawaan http murni (tanpa detail)
	errNoDetail := NotFound("")
	expected1 := "Not Found"
	if errNoDetail.Error() != expected1 {
		t.Errorf("expected %q, got %q", expected1, errNoDetail.Error())
	}
	if errNoDetail.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", errNoDetail.Code)
	}

	// 2. Text bawaan + detail tambahan
	errWithDetail := BadRequest("Email is invalid")
	expected2 := "Bad Request. Email is invalid"
	if errWithDetail.Error() != expected2 {
		t.Errorf("expected %q, got %q", expected2, errWithDetail.Error())
	}
	if errWithDetail.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", errWithDetail.Code)
	}

	// 3. Text bawaan + underlying root error
	rootErr := errors.New("connection timeout")
	errWithRoot := Internal("", rootErr)
	expected3 := "Internal Server Error. connection timeout"
	if errWithRoot.Error() != expected3 {
		t.Errorf("expected %q, got %q", expected3, errWithRoot.Error())
	}
	if !errors.Is(errWithRoot, rootErr) {
		t.Errorf("expected unwrapped error to match rootErr")
	}
}
