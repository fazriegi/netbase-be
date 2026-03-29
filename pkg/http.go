package pkg

import (
	"errors"
	"net/http"
	"reflect"
	"strconv"
)

func MapQueryTags(v reflect.Value, result map[string]reflect.Value) {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := v.Field(i)

		// Handle Embedded Struct (Recursive)
		if field.Anonymous && fieldVal.Kind() == reflect.Struct {
			MapQueryTags(fieldVal, result)
			continue
		}

		tag := field.Tag.Get("query")
		if tag == "" || tag == "-" {
			continue
		}

		// Simpan reflect.Value-nya agar bisa di-set nilainya nanti
		result[tag] = fieldVal
	}
}

func ParseQueryParam(r *http.Request, dest interface{}) error {
	val := reflect.ValueOf(dest)

	for val.Kind() == reflect.Ptr && val.Elem().Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return errors.New("dest must be a pointer to a struct")
	}

	structVal := val.Elem()
	fieldMap := make(map[string]reflect.Value)

	MapQueryTags(structVal, fieldMap)

	queryParams := r.URL.Query()

	for tag, field := range fieldMap {
		valStr := queryParams.Get(tag)
		if valStr == "" {
			continue
		}

		// --- HANDLE POINTER FIELD ---
		isPtr := field.Kind() == reflect.Ptr
		var targetField reflect.Value
		var baseKind reflect.Kind

		if isPtr {
			// Jika pointer, alokasikan memori baru untuk tipe dasarnya (elem)
			targetField = reflect.New(field.Type().Elem()).Elem()
			baseKind = targetField.Kind()
		} else {
			// Jika bukan pointer, langsung gunakan field-nya
			targetField = field
			baseKind = targetField.Kind()
		}

		// Set nilai ke targetField berdasarkan tipe dasarnya
		switch baseKind {
		case reflect.String:
			targetField.SetString(valStr)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			i, err := strconv.ParseInt(valStr, 10, 64)
			if err == nil {
				targetField.SetInt(i)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			u, err := strconv.ParseUint(valStr, 10, 64)
			if err == nil {
				targetField.SetUint(u)
			}
		case reflect.Bool:
			b, err := strconv.ParseBool(valStr)
			if err == nil {
				targetField.SetBool(b)
			}
		case reflect.Float32, reflect.Float64:
			f, err := strconv.ParseFloat(valStr, 64)
			if err == nil {
				targetField.SetFloat(f)
			}
		}

		// Jika field di struct aslinya adalah pointer, kita set nilainya dengan address dari targetField
		if isPtr {
			field.Set(targetField.Addr())
		}
	}

	return nil
}
