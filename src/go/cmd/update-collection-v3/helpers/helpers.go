package helpers

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
)

// CheckForConflictsInRest checks if Rest property contains structure properties keys
func CheckForConflictsInRest(obj interface{}) ([]string, error) {
	return CheckForConflictsInRestRecursive(reflect.ValueOf(obj), obj, "")
}

// Runs CheckForConflictsInRestRecursive recursively checks is Rest field contains duplication of any property
func CheckForConflictsInRestRecursive(v reflect.Value, obj interface{}, prefix string) ([]string, error) {
	knownProperties := getKnownProperties(obj)
	duplicated := []string{}
	var e error = nil

	// Handle pointers to structs
	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct {
		v = v.Elem()
	}

	// Skip if element is not a structure
	if v.Kind() != reflect.Struct {
		return duplicated, nil
	}

	t := v.Type()
	// Take rest field
	rest := v.FieldByName("Rest")
	if rest.IsValid() {
		for _, key := range rest.MapKeys() {
			keyStr := key.String()
			if slices.Contains(knownProperties, keyStr) {
				duplicated = append(duplicated, fmt.Sprintf("%s%s", prefix, keyStr))
			}
		}
	}

	for i := 0; i < t.NumField(); i++ {
		fieldValue := v.Field(i)
		structField := t.Field(i)

		if !fieldValue.CanInterface() {
			continue
		}

		d, _ := CheckForConflictsInRestRecursive(fieldValue, fieldValue.Interface(), fmt.Sprintf("%v%v.", prefix, getTagName(structField)))
		duplicated = append(duplicated, d...)
	}

	if len(duplicated) > 0 {
		e = fmt.Errorf("conflict between input and output values for the following keys: %s", strings.Join(duplicated, ", "))
	}
	return duplicated, e
}

// Helper function to get the known property names from the struct's field tags
func getKnownProperties(s interface{}) []string {
	var knownProperties []string
	v := reflect.ValueOf(s)

	// Handle pointers to structs
	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return []string{}
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := getTagName(field)

		if tag == "" {
			continue
		}
		knownProperties = append(knownProperties, tag)
	}
	return knownProperties
}

func getTagName(field reflect.StructField) string {
	tag := field.Tag.Get("yaml")
	parts := strings.Split(tag, ",")
	if len(parts) != 2 || parts[0] == "" || parts[0] == "-" {
		return ""
	}

	return parts[0]
}
