/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ns

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/intelsdi-x/snap-plugin-utilities/str"
	"github.com/oleiade/reflections"
)

type FlagFunc (func(nsPath string, itemKind reflect.Type) bool)

// FlagFunc is invoked to determine flag value at position in namespace

type OptionFunc (func() (int, FlagFunc))

// OptionFunc is used to attach flag functions for specific flags

//notAllowedChars array with not allowed characters in namespace
var notAllowedChars = map[string][]string{
	"brackets":     {"(", ")", "[", "]", "{", "}"},
	"spaces":       {" "},
	"punctuations": {".", ",", ";", "?", "!"},
	"slashes":      {"|", "\\", "/"},
	"carets":       {"^"},
	"quotations":   {"\"", "`", "'"},
}

// FromMap constructs list of namespaces from multilevel map using map keys as namespace entries.
// 'Current' value is prefixed to all namespace elements.
// It returns nil in case of success or error if building namespaces failed.
func FromMap(m map[string]interface{}, current string, namespace *[]string) error {
	return FromCompositeObject(m, current, namespace,
		InspectNilPointers(AlwaysFalse),
		InspectEmptyContainers(AlwaysFalse),
		ExportJsonFieldNames(AlwaysFalse))
}

// FromJSON constructs list of namespaces from json document using json literals as namespace entries.
// 'Current' value is prefixed to all namespace elements.
// It returns nil in case of success or error if building namespaces failed.
func FromJSON(data *[]byte, current string, namespace *[]string) error {

	var m map[string]interface{}
	err := json.Unmarshal(*data, &m)

	if err != nil {
		return err
	}

	return FromMap(m, current, namespace)
}

// FromComposition constructs list of namespaces from multilevel struct compositions using field names as namespace entries.
// 'Current' value is prefixed to all namespace elements.
// It returns nil in case of success or error if building namespaces failed.
func FromComposition(object interface{}, current string, namespace *[]string) error {
	return FromCompositeObject(object, current, namespace,
		InspectNilPointers(AlwaysFalse),
		InspectEmptyContainers(AlwaysFalse),
		ExportJsonFieldNames(AlwaysFalse))
}

// FromCompositionTags constructs list of namespaces from multilevel struct composition using field tags as namespace entries.
// 'Current' value is prefixed to all namespace elements.
// It returns nil in case of success or error if building namespaces failed.
func FromCompositionTags(object interface{}, current string, namespace *[]string) error {
	return FromCompositeObject(object, current, namespace,
		InspectNilPointers(AlwaysFalse),
		InspectEmptyContainers(AlwaysFalse))
}

// GetValueByNamespace returns value stored in nested map, array or struct composition.
// TODO Implementation of handling complex compositions with map/slice
func GetValueByNamespace(object interface{}, ns []string) interface{} {

	if len(ns) == 0 {
		fmt.Fprintf(os.Stderr, "Namespace length equal to zero\n")
		return nil
	}

	if object == nil {
		fmt.Fprintf(os.Stderr, "First parameter cannot be nil!\n")
		return nil
	}

	// current level of namespace
	current := ns[0]

	switch reflect.TypeOf(object).Kind() {
	case reflect.Map:
		if m, ok := object.(map[string]interface{}); ok {
			if val, ok := m[current]; ok {
				if len(ns) == 1 {
					return val
				}
				return GetValueByNamespace(val, ns[1:])
			}
			fmt.Fprintf(os.Stderr, "Key does not exist in map {key %s}\n", current)
			return nil
		}
	case reflect.Slice:
		curr, err := strconv.Atoi(current)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cannot convert index to integer {idx %v}\n", current)
			return nil
		}

		if a, ok := object.([]interface{}); ok {
			if curr >= len(a) {
				fmt.Fprintf(os.Stderr, "Index out of range {idx %v}\n", current)
				return nil
			}
			if len(ns) == 1 {
				return a[curr]
			}
			return GetValueByNamespace(a[curr], ns[1:])
		}
	case reflect.Ptr:
		// TODO Implementation of handling pointers to objects other than structs
		return GetStructValueByNamespace(object, ns)
	case reflect.Struct:
		return GetStructValueByNamespace(object, ns)
	default:
		fmt.Fprintf(os.Stderr, "Unsupported object type {object %v}", object)
	}

	return nil
}

// GetStructValueByNamespace returns value stored in struct composition.
// It requires filed tags on each struct field which may be represented as namespace component.
// It iterates over fields recursively, checks tags until it finds leaf value.
func GetStructValueByNamespace(object interface{}, ns []string) interface{} {
	// current level of namespace
	current := ns[0]

	fields, err := reflections.Fields(object)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not return fields for object{%v}\n", object)
		return nil
	}

	for _, field := range fields {
		tag, err := reflections.GetFieldTag(object, field, "json")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not find tag for field{%s}\n", field)
			return nil
		}

		// remove omitempty from tag
		tag = strings.Split(tag, ",")[0]

		if tag == current {
			val, err := reflections.GetField(object, field)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not retrieve field{%s}\n", field)
				return nil
			}
			// handling of special cases for slice and map
			switch reflect.TypeOf(val).Kind() {
			case reflect.Slice:
				idx, _ := strconv.Atoi(ns[1])
				val := reflect.ValueOf(val)
				if val.Index(idx).Kind() == reflect.Struct {
					return GetStructValueByNamespace(val.Index(idx).Interface(), ns[2:])
				} else {
					return val.Index(idx).Interface()
				}
			case reflect.Map:
				key := ns[1]

				if vi, ok := val.(map[string]uint64); ok {
					return vi[key]
				}

				val := reflect.ValueOf(val)
				kval := reflect.ValueOf(key)
				if reflect.TypeOf(val.MapIndex(kval).Interface()).Kind() == reflect.Struct {
					return GetStructValueByNamespace(val.MapIndex(kval).Interface(), ns[2:])
				}
			case reflect.Ptr:
				v := reflect.Indirect(reflect.ValueOf(val))
				if !v.IsValid() || !v.CanInterface() {
					// If reflect can't call Interface() on v, we can't go deeper even if
					// len(ns) > 1. Therefore, we should just return nil here.
					return nil
				}

				if len(ns) == 1 {
					return v.Interface()
				} else {
					return GetStructValueByNamespace(v.Interface(), ns[1:])
				}
			default:
				// last ns, return value found
				if len(ns) == 1 {
					return val
				} else {
					// or go deeper
					return GetStructValueByNamespace(val, ns[1:])
				}
			}
		}
	}
	return nil
}

//ReplaceNotAllowedCharsInNamespacePart replaces not allowed characters in namespace part  by '_'
func ReplaceNotAllowedCharsInNamespacePart(ns string) string {
	for _, chars := range notAllowedChars {
		for _, ch := range chars {
			ns = strings.Replace(ns, ch, "_", -1)
			ns = strings.Replace(ns, "__", "_", -1)
		}
	}
	ns = strings.Trim(ns, "_")
	return ns
}

//ValidateMetricNamespacePart checks if namespace part contains not allowed characters
func ValidateMetricNamespacePart(ns string) error {
	for _, chars := range notAllowedChars {
		for _, ch := range chars {
			if strings.ContainsAny(ns, ch) {
				return fmt.Errorf("Namespace contains not allowed chars, namespace part: %s, not allowed char: %s", ns, ch)
			}
		}
	}
	return nil
}

// getJsonFieldName gets an object's field name compatible with json tag on;
// details can be found in docs for pkg  json. If json tag is empty, or no
// tag is present then original  fieldName will be reported. If json tag hides
// a field, function will return dash "-" as name.
func getJsonFieldName(object interface{}, fieldName string) (string, error) {
	jsonTag, err := reflections.GetFieldTag(object, fieldName, "json")
	if err != nil {
		return "", err
	} else if jsonTag == "" {
		return fieldName, nil
	}
	i := strings.Index(jsonTag, ",")
	if i == -1 {
		return jsonTag, nil
	}
	if tag := jsonTag[:i]; tag == "-" {
		return "-", nil
	} else if tag == "" {
		return fieldName, nil
	} else {
		return tag, nil
	}
}

func AlwaysTrue(_ string, _ reflect.Type) bool {
	return true
}

func AlwaysFalse(_ string, _ reflect.Type) bool {
	return false
}

// InspectNilPointers option controls NS expansion for nil pointers.
//
// If InspectNilPointers returns  true for nil pointer at some path in ns,
// ns will be expanded for pointer's original type
func InspectNilPointers(flagFunc FlagFunc) OptionFunc {
	return func() (int, FlagFunc) {
		return inspectNilPointers, flagFunc
	}
}

// InspectEmptyContainers option controls NS expansion for empty maps.
//
// If InspectEmptyContainers returns  true for map at some path in ns, ns will
// be expanded for map's value type
func InspectEmptyContainers(flagFunc FlagFunc) OptionFunc {
	return func() (int, FlagFunc) {
		return inspectEmptyContainers, flagFunc
	}
}

// EntryForContainersRoot controls inserting entries for containers themselves.
//
// If EntryForContainersRoot returns  true for container at some path in ns,
// ns will contain entry for a container itself.
func EntryForContainersRoot(flagFunc FlagFunc) OptionFunc {
	return func() (int, FlagFunc) {
		return entryForContainersRoot, flagFunc
	}
}

// ExportJsonFieldNames option controls naming fields of struct where field is
// annotated with json tag.
//
// If ExportJsonFieldNames returns  true for struct at some path in ns, json
// name for all fields will be used rather than fields' own names. If json tag
// contains dash ('-') for any field, field won't be exported.
func ExportJsonFieldNames(flagFunc FlagFunc) OptionFunc {
	return func() (int, FlagFunc) {
		return exportJsonFieldNames, flagFunc
	}
}

// WildcardEntryInContainer controls adding a wildcard for inspected non-empty
// containers.
//
// If WildcardEntryInContainer returns  true for non-empty container
// (map/ slice/ array) at some path in ns, a wildcard entry will be reported
// for container and the containers' value type will be inspected in depth.
//
// See also: InspectEmptyContainers
func WildcardEntryInContainer(flagFunc FlagFunc) OptionFunc {
	return func() (int, FlagFunc) {
		return wildcardEntryInContainer, flagFunc
	}
}

// SanitizeNamespaceParts controls sanitizing namespace parts in generated
// namespace.
//
// If SanitizeNamespaceParts returns  true for object at some path in ns, all
// child keys of that object (fields, map keys) will have their names
// sanitized, ie. all the invalid characters removed.
//
// See also: ns.ReplaceNotAllowedCharsInNamespacePart
func SanitizeNamespaceParts(flagFunc FlagFunc) OptionFunc {
	return func() (int, FlagFunc) {
		return sanitizeNamespaceParts, flagFunc
	}
}

const (
	inspectNilPointers = iota
	inspectEmptyContainers
	entryForContainersRoot
	exportJsonFieldNames
	wildcardEntryInContainer
	sanitizeNamespaceParts
)

// CompositeObjectToNs inspects an object to construct a list of paths to data.
//
// Operation of this method is controlled by options, which default to:
// 	InspectEmptyContainers(AlwaysTrue),
// 	InspectNilPointers(AlwaysTrue),
// 	EntryForContainersRoot(AlwaysFalse),
// 	ExportJsonFieldNames(AlwaysTrue),
// 	WildcardEntryInContainer(AlwaysFalse),
//	SanitizeNamespaceParts(AlwaysTrue).
// Different options may be specified to implement selective and
// context-sensitive inspection.
func FromCompositeObject(object interface{}, current string, namespace *[]string, options ...OptionFunc) error {
	flags := map[int]FlagFunc{}
	options = append([]OptionFunc{
		InspectEmptyContainers(AlwaysTrue),
		InspectNilPointers(AlwaysTrue),
		EntryForContainersRoot(AlwaysFalse),
		ExportJsonFieldNames(AlwaysTrue),
		WildcardEntryInContainer(AlwaysFalse),
		SanitizeNamespaceParts(AlwaysTrue)}, options...)
	for _, option := range options {
		key, filter := option()
		flags[key] = filter
	}
	return fromCompositeObject(object, current, namespace, flags)
}

func fromCompositeObject(object interface{}, current string, namespace *[]string, flags map[int]FlagFunc) error {
	val := reflect.Indirect(reflect.ValueOf(object))
	saneAppendNs := func() {
		if current != "" {
			if !str.Contains(*namespace, current) {
				*namespace = append(*namespace, current)
			}
		}
	}
	safeExtendNs := func(part string) string {
		if flags[sanitizeNamespaceParts](current, val.Type()) {
			return filepath.Join(current, ReplaceNotAllowedCharsInNamespacePart(part))
		}
		return filepath.Join(current, part)
	}
	regularExtendNs := func(part string) string {
		return filepath.Join(current, part)
	}

	switch val.Kind() {
	case reflect.Invalid:
		val = reflect.ValueOf(object)
		if val.Kind() != reflect.Ptr || !val.IsNil() {
			return nil
		}
		if false == flags[inspectNilPointers](current, val.Type()) {
			return nil
		}
		nuObj := reflect.Zero(val.Type().Elem())
		if err := fromCompositeObject(nuObj.Interface(), current, namespace, flags); err != nil {
			return err
		}
		return nil
	case reflect.Ptr:
		return fromCompositeObject(val.Interface(), current, namespace, flags)
	case reflect.Map:
		if true == flags[entryForContainersRoot](current, val.Type()) {
			saneAppendNs()
		}
		wildcardEntryInContainer := flags[wildcardEntryInContainer](current, val.Type())
		if val.Len() == 0 || wildcardEntryInContainer {
			if !wildcardEntryInContainer && false == flags[inspectEmptyContainers](current, val.Type()) {
				return nil
			}
			typ := reflect.TypeOf(object)
			nuObj := reflect.Zero(typ.Elem()).Interface()
			if err := fromCompositeObject(
				nuObj,
				regularExtendNs("*"),
				namespace,
				flags); err != nil {
				return err
			}
		}

		for _, mkey := range val.MapKeys() {
			if err := fromCompositeObject(
				val.MapIndex(mkey).Interface(),
				safeExtendNs(mkey.String()),
				namespace,
				flags); err != nil {
				return err
			}
		}
	case reflect.Array, reflect.Slice:
		if true == flags[entryForContainersRoot](current, val.Type()) {
			saneAppendNs()
		}
		wildcardEntryInContainer := flags[wildcardEntryInContainer](current, val.Type())
		if val.Len() == 0 || wildcardEntryInContainer {
			if !wildcardEntryInContainer && false == flags[inspectEmptyContainers](current, val.Type()) {
				return nil
			}
			typ := reflect.TypeOf(object)
			nuObj := reflect.Zero(typ.Elem()).Interface()
			if err := fromCompositeObject(
				nuObj,
				regularExtendNs("*"),
				namespace,
				flags); err != nil {
				return err
			}
		}
		for i := 0; i < val.Len(); i++ {
			if err := fromCompositeObject(val.Index(i).Interface(),
				regularExtendNs(strconv.Itoa(i)),
				namespace,
				flags); err != nil {
				return err
			}
		}
	case reflect.Struct:
		if true == flags[entryForContainersRoot](current, val.Type()) {
			saneAppendNs()
		}
		fields, err := reflections.Fields(object)
		if err != nil {
			return err
		}
		exportJsonFieldNamesHere := flags[exportJsonFieldNames](current, val.Type())
		for _, field := range fields {
			f, err := reflections.GetField(object, field)
			if err != nil {
				return err
			}
			fieldName := field
			if true == exportJsonFieldNamesHere {
				jsonField, err := getJsonFieldName(object, field)
				if err != nil {
					return err
				}
				// hidden field - skip it
				if jsonField == "-" {
					continue
				}
				fieldName = jsonField
			}
			nuCurrent := safeExtendNs(fieldName)
			if err := fromCompositeObject(f, nuCurrent, namespace, flags); err != nil {
				return err
			}
		}
	default:
		saneAppendNs()
	}
	if len(*namespace) == 0 {
		return fmt.Errorf("Namespace empty!")
	}
	return nil
}
