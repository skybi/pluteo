package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

var (
	errRequestBodyInvalidJSON = func(err string) *Error {
		return &Error{
			Type:    "validation.requestBody.invalidJSON",
			Message: "Request body is not a valid JSON input.",
			Details: map[string]any{
				"error": err,
			},
		}
	}
	errRequestBodyParameterInvalidType = func(name, expectedType string) *Error {
		return &Error{
			Type:    "validation.requestBody.parameter.invalidType",
			Message: fmt.Sprintf("The request body parameter '%s' could not be assigned to the required type (%s).", name, expectedType),
			Details: map[string]any{
				"parameter":     name,
				"expected_type": expectedType,
			},
		}
	}
	errRequestBodyParameterMissing = func(name string) *Error {
		return &Error{
			Type:    "validation.requestBody.parameter.missing",
			Message: fmt.Sprintf("The request body parameter '%s' is required but was not present in the request.", name),
			Details: map[string]any{
				"parameter": name,
			},
		}
	}
	errRequestBodyParameterNumberOutOfRange = func(name string, value, min, max int64) *Error {
		comparison := ""
		if value < min {
			comparison = fmt.Sprintf("%d [given] < %d [min]", value, min)
		} else if value > max {
			comparison = fmt.Sprintf("%d [given] > %d [max]", value, max)
		}

		return &Error{
			Type:    "validation.requestBody.parameter.number.outOfRange",
			Message: fmt.Sprintf("The request body parameter '%s' is out of the required range (%s).", name, comparison),
			Details: map[string]any{
				"parameter": name,
				"value":     value,
				"min":       min,
				"max":       max,
			},
		}
	}
)

// UnmarshalBody parses and decodes a JSON request body and performs validations on it
func UnmarshalBody[T any](request *http.Request) (*T, []*Error, error) {
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, nil, err
	}

	target := new(T)
	if err := json.Unmarshal(body, target); err != nil {
		if typeErr, ok := err.(*json.UnmarshalTypeError); ok {
			return nil, []*Error{errRequestBodyParameterInvalidType(typeErr.Field, typeErr.Type.String())}, nil
		} else {
			return nil, []*Error{errRequestBodyInvalidJSON(err.Error())}, nil
		}
	}

	errs, err := validateStruct("", target)
	if err != nil {
		return nil, nil, err
	}
	return target, errs, nil
}

func validateStruct(fieldPrefix string, val any) ([]*Error, error) {
	typ := reflect.TypeOf(val)
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil, errors.New("illegal call to validateStruct with non-struct parameter")
	}
	ref := reflect.ValueOf(val)
	if ref.Kind() == reflect.Pointer {
		ref = ref.Elem()
	}

	var errs []*Error

	for i := 0; i < typ.NumField(); i++ {
		// Retrieve the validation requirements
		fieldDef := typ.Field(i)
		required := strings.EqualFold(fieldDef.Tag.Get("required"), "true")
		min, err := strconv.ParseInt(fieldDef.Tag.Get("min"), 10, 64)
		if err != nil {
			min = math.MinInt64
		}
		max, err := strconv.ParseInt(fieldDef.Tag.Get("max"), 10, 64)
		if err != nil {
			max = math.MaxInt64
		}

		fieldName := getFieldName(fieldDef)

		// Perform all validations on the field
		field := ref.Field(i)
		if required && field.IsNil() {
			errs = append(errs, errRequestBodyParameterMissing(fieldPrefix+fieldName))
		}
		if field.Kind() == reflect.Pointer {
			field = field.Elem()
		}
		if field.CanUint() {
			val := int64(field.Uint())
			if val < min || val > max {
				errs = append(errs, errRequestBodyParameterNumberOutOfRange(fieldPrefix+fieldName, val, min, max))
			}
		} else if field.CanInt() {
			val := field.Int()
			if val < min || val > max {
				errs = append(errs, errRequestBodyParameterNumberOutOfRange(fieldPrefix+fieldName, val, min, max))
			}
		} else if field.Kind() == reflect.Struct || (field.Kind() == reflect.Pointer && field.Elem().Kind() == reflect.Struct) {
			var val any
			if field.Kind() == reflect.Struct {
				val = field.Interface()
			} else {
				val = field.Elem().Interface()
			}

			subErrs, err := validateStruct(fieldPrefix+fieldName+".", val)
			if err != nil {
				return nil, err
			}
			errs = append(errs, subErrs...)
		}
	}

	return errs, nil
}

func getFieldName(def reflect.StructField) string {
	jsonVal, ok := def.Tag.Lookup("json")
	if !ok || jsonVal == "-" {
		return def.Name
	}
	name, _, _ := strings.Cut(jsonVal, ",")
	return name
}
