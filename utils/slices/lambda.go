package slices

import "reflect"

type Slice []interface{}

func (s Slice) Filter(predicate func(interface{}) bool) []interface{} {
	filtered := []interface{}{}
	for _, item := range s {
		if predicate(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// Filter filters elements in a slice of any built-in type based on the provided condition
func Filter(slice []interface{}, predicate func(interface{}) bool) []interface{} {
	filtered := []interface{}{}
	for _, item := range slice {
		if predicate(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// FilterStruct filters elements in a slice of any struct type based on the provided condition
func FilterStruct(slice interface{}, predicate interface{}) interface{} {
	sliceValue := reflect.ValueOf(slice)
	if sliceValue.Kind() != reflect.Slice {
		panic("FilterStructByCondition: not a slice")
	}

	conditionValue := reflect.ValueOf(predicate)
	if conditionValue.Kind() != reflect.Func {
		panic("FilterStructByCondition: not a function")
	}

	elemType := sliceValue.Type().Elem()
	resultSlice := reflect.MakeSlice(reflect.SliceOf(elemType), 0, 0)

	for i := 0; i < sliceValue.Len(); i++ {
		elem := sliceValue.Index(i)
		conditionResult := conditionValue.Call([]reflect.Value{elem})
		if len(conditionResult) > 0 && conditionResult[0].Kind() == reflect.Bool && conditionResult[0].Bool() {
			resultSlice = reflect.Append(resultSlice, elem)
		}
	}
	return resultSlice.Interface()
}
