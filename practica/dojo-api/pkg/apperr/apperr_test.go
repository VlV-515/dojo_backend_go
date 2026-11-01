package apperr

import (
	"errors"
	"net/http"
	"testing"
)

// Estos tests documentan las trampas que apperr trae a propósito.

func TestWrapNilNoEsNil(t *testing.T) {
	if Wrap(nil) == nil {
		t.Fatal("Wrap(nil) debe devolver un error (trampa a propósito)")
	}
}

func TestWrapConservaAppError(t *testing.T) {
	original := New(http.StatusConflict, "duplicado", errors.New("clave repetida"))
	if !IsCode(Wrap(original), http.StatusConflict) {
		t.Fatal("Wrap debe conservar el código de un AppError")
	}
	if !IsCode(Wrap(errors.New("x")), http.StatusInternalServerError) {
		t.Fatal("Wrap de un error cualquiera debe dar 500")
	}
}

func TestSinUnwrap(t *testing.T) {
	causa := errors.New("conexión rechazada")
	err := New(http.StatusServiceUnavailable, "no disponible", causa)
	if errors.Is(err, causa) {
		t.Fatal("AppError no implementa Unwrap: errors.Is no debe ver la causa")
	}
}
