package search_v3

import (
	"fmt"
	"github.com/fatih/structtag"
	"github.com/stretchr/testify/require"
	"go-playground/protobuf/utils"
	"reflect"
	"sort"
	"testing"
)

func requireDeepEqual(t *testing.T, leftV, rightV reflect.Value) {
	require.Equal(t, leftV.IsValid(), rightV.IsValid())
	if !leftV.IsValid() {
		return
	}

	leftT := leftV.Type()
	rightT := rightV.Type()

	if leftT.Kind() != rightT.Kind() {
		if leftT.Kind() == reflect.Ptr {
			requireDeepEqual(t, leftV.Elem(), rightV)
			return
		}
		if rightT.Kind() == reflect.Ptr {
			requireDeepEqual(t, leftV, rightV.Elem())
			return
		}
	}

	// Features of proto formats
	if leftT.Kind() == reflect.Map && rightT.Kind() == reflect.Struct && rightT.Name() == "MapStringString" {
		requireDeepEqual(t, leftV, rightV.FieldByName("Map"))
		return
	}
	if leftT.Kind() == reflect.Slice && rightT.Kind() == reflect.Struct && rightT.NumField() == 4 {
		requireDeepEqual(t, leftV, findFirstExportedField(rightV))
		return
	}

	// Time can be represented as int64
	if leftT.Kind() == reflect.Struct && rightT.Kind() == reflect.Int64 && leftT.Name() == "Time" {
		return
	}

	// DateTime can be represented as string
	if leftT.Kind() == reflect.Struct && rightT.Kind() == reflect.String && leftT.Name() == "DateTime" {
		return
	}

	// Some exceptional cases
	if leftT.Kind() == reflect.Struct && leftT.Name() == "Code" && rightT.Kind() == reflect.Int32 {
		return
	}
	if leftT.Kind() == reflect.Struct && leftT.Name() == "PointerBool" && rightT.Kind() == reflect.Struct && rightT.Name() == "OptBool" {
		return
	}

	switch leftT.Kind() {
	case reflect.Slice, reflect.Array:
		require.Equal(t, leftT.Kind().String(), rightT.Kind().String())
		requireDeepEqualSlice(t, leftV, rightV)
	case reflect.Struct:
		require.Equal(t, leftT.Kind().String(), rightT.Kind().String())
		requireDeepEqualStruct(t, leftV, rightV)
	case reflect.Ptr:
		require.Equal(t, leftT.Kind().String(), rightT.Kind().String())
		requireDeepEqual(t, leftV.Elem(), rightV.Elem())
	case reflect.Map:
		require.Equal(t, leftT.Kind().String(), rightT.Kind().String())
		requireDeepEqualMap(t, leftV, rightV)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		require.Equal(t, leftV.Int(), rightV.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		require.Equal(t, leftV.Uint(), rightV.Uint())
	case reflect.String:
		require.Equal(t, leftV.String(), rightV.String())
	case reflect.Float64, reflect.Float32:
		require.Equal(t, leftV.Float(), rightV.Float())
	case reflect.Bool:
		require.Equal(t, leftT.Kind().String(), rightT.Kind().String())
		require.Equal(t, leftV.Bool(), rightV.Bool())
	default:
		require.FailNow(t, "Unsupported kind %s", leftT.Kind())
	}
}

func requireDeepEqualSlice(t *testing.T, left, right reflect.Value) {
	require.Equal(t, left.Len(), right.Len())

	for i := 0; i < left.Len(); i++ {
		leftV := left.Index(i)
		rightV := right.Index(i)
		requireDeepEqual(t, leftV, rightV)
	}
}

func requireDeepEqualMap(t *testing.T, left, right reflect.Value) {
	require.Equal(t, left.Len(), right.Len())

	leftKeys := left.MapKeys()
	rightKeys := right.MapKeys()
	sortReflect(leftKeys)
	sortReflect(rightKeys)

	for i, leftKey := range leftKeys {
		rightKey := rightKeys[i]

		leftMapValue := left.MapIndex(leftKey)
		rightMapValue := right.MapIndex(rightKeys[i])
		fmt.Printf("Checking map value (%v/%v) for equality\n", leftKey.Interface(), rightKey.Interface())
		requireDeepEqual(t, leftMapValue, rightMapValue)
	}
}

func sortReflect(keys []reflect.Value) {
	if len(keys) == 0 {
		return
	}
	v := keys[0]
	t := v.Type()
	var sortFunc func(i, j int) bool

	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		sortFunc = func(i, j int) bool {
			iVal := keys[i].Int()
			jVal := keys[j].Int()
			return iVal < jVal
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		sortFunc = func(i, j int) bool {
			iVal := keys[i].Uint()
			jVal := keys[j].Uint()
			return iVal < jVal
		}
	case reflect.String:
		sortFunc = func(i, j int) bool {
			iVal := keys[i].String()
			jVal := keys[j].String()
			return iVal < jVal
		}
	case reflect.Float64, reflect.Float32:
		sortFunc = func(i, j int) bool {
			iVal := keys[i].Float()
			jVal := keys[j].Float()
			return iVal < jVal
		}
	case reflect.Struct:
		if _, ok := v.Interface().(fmt.Stringer); ok {
			sortFunc = func(i, j int) bool {
				iVal := keys[i].Interface().(fmt.Stringer).String()
				jVal := keys[j].Interface().(fmt.Stringer).String()
				return iVal < jVal
			}
		} else {
			panic(fmt.Errorf("unsortable type %s", t.String()))
		}
	default:
		panic(fmt.Errorf("unsortable type %s", t.String()))
	}

	sort.SliceStable(keys, sortFunc)
}

func requireDeepEqualStruct(t *testing.T, left, right reflect.Value) {
	leftFieldsByTags := fieldsByJsonTags(left)
	rightFieldsByTags := fieldsByJsonTags(right)
	fmt.Printf("Checking struct %s for equality, found %d fields\n", left.Type().Name(), len(leftFieldsByTags))

	require.Equal(t, len(leftFieldsByTags), len(rightFieldsByTags), "wrong len for %s", left.Type().Name())
	for key, lValue := range leftFieldsByTags {
		rValue := rightFieldsByTags[key]

		fmt.Printf("Checking field %s for equality\n", key)
		requireDeepEqual(t, lValue, rValue)
	}
}

func fieldsByJsonTags(obj reflect.Value) map[string]reflect.Value {
	result := make(map[string]reflect.Value)
	objT := obj.Type()
	for i := 0; i < objT.NumField(); i++ {
		leftTField := objT.Field(i)
		tags := utils.Must2(structtag.Parse(string(leftTField.Tag)))
		if jsonTag, err := tags.Get("json"); err == nil && jsonTag.Name != "" {
			result[jsonTag.Name] = obj.Field(i)
		}
	}
	return result
}

func findFirstExportedField(obj reflect.Value) reflect.Value {
	for i := 0; i < obj.NumField(); i++ {
		f := obj.Field(i)
		if f.CanSet() {
			return f
		}
	}
	panic(fmt.Errorf("unable to find exported field of %s", obj.Type().Name()))
}
