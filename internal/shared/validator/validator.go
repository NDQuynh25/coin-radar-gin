package validator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"coin-radar-gin/internal/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Init wires custom validation behavior into Gin's shared validator.
// Call once at startup (before serving). It makes field errors report the
// JSON field name (e.g. "email") instead of the Go struct field ("Email").
func Init() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// Bind parses and validates the JSON body into target. On failure it writes a
// 400 response with structured field errors and returns false, so callers do:
//
//	if !validator.Bind(c, &req) { return }
func Bind(c *gin.Context, target interface{}) bool {
	err := c.ShouldBindJSON(target)
	if err == nil {
		return true
	}
	c.JSON(http.StatusBadRequest, errorResponse(err))
	return false
}

// errorResponse translates a binding error into a client-friendly Response.
func errorResponse(err error) response.Response {
	// Field-level validation failures.
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		fields := make(map[string]string, len(ve))
		for _, fe := range ve {
			fields[fe.Field()] = message(fe)
		}
		return response.ValidationError("validation_failed", "one or more fields are invalid", fields)
	}

	// Wrong JSON type for a field, e.g. string where a number is expected.
	var ute *json.UnmarshalTypeError
	if errors.As(err, &ute) {
		return response.ValidationError(
			"invalid_request",
			"one or more fields have the wrong type",
			map[string]string{ute.Field: fmt.Sprintf("must be of type %s", ute.Type.String())},
		)
	}

	// Empty body.
	if errors.Is(err, io.EOF) {
		return response.Error("invalid_request", "request body is empty")
	}

	// Malformed JSON or anything else — don't leak internal parser details.
	return response.Error("invalid_request", "malformed JSON body")
}

// message renders a human-readable message for a single field error.
func message(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "gt":
		return fmt.Sprintf("must be greater than %s", fe.Param())
	default:
		return "invalid value"
	}
}
