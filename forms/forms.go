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

		formValue := r.FormValue(formKey)

		if formKey != "password" {
			formValue = strings.TrimSpace(formValue)
		}
		switch formKey {
		case "email":
			formValue = strings.ToLower(formValue)
		case "name":
			if len(formValue) > 0 {
				formValue = strings.ToUpper(formValue[:1]) + formValue[1:]
			}
		}

		if !field.CanSet() {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(formValue)
		case reflect.Int, reflect.Int64:
			intVal, err := strconv.ParseInt(formValue, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid int for field '%s': %v", formKey, err)
			}
			field.SetInt(intVal)
		case reflect.Float64:
			floatVal, err := strconv.ParseFloat(formValue, 64)
			if err != nil {
				return fmt.Errorf("invalid float for field '%s': %v", formKey, err)
			}
			field.SetFloat(floatVal)
		case reflect.Bool:
			boolVal, err := strconv.ParseBool(formValue)
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
