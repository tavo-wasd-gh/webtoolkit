package forms

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

func ParseForm(r *http.Request, dst interface{}) error {
	err := r.ParseForm()
	if err != nil {
		return err
	}

	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("dst must be a pointer to a struct")
	}

	v = v.Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)

		formKey := structField.Tag.Get("form")
		if formKey == "" {
			continue
		}

		required := structField.Tag.Get("req") == "1"

		formValue, found := r.Form[formKey]
		if !found || len(formValue[0]) == 0 {
			if required {
				return fmt.Errorf("missing required form field: %s", formKey)
			}
			continue // skip if not required
		}

		value := formValue[0]

		if formKey != "password" {
			value = strings.TrimSpace(value)
		}

		switch formKey {
		case "email":
			value = strings.ToLower(value)
		}

		if !field.CanSet() {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(value)
		case reflect.Int, reflect.Int64:
			intVal, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid int for field '%s': %v", formKey, err)
			}
			field.SetInt(intVal)
		case reflect.Float64:
			floatVal, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return fmt.Errorf("invalid float for field '%s': %v", formKey, err)
			}
			field.SetFloat(floatVal)
		case reflect.Bool:
			boolVal, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("invalid bool for field '%s': %v", formKey, err)
			}
			field.SetBool(boolVal)
		default:
			return fmt.Errorf("unsupported field type: %s", field.Kind())
		}
	}

	return nil
}
