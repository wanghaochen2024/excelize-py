// Copyright 2024 The excelize Authors. All rights reserved. Use of this source
// code is governed by a BSD-style license that can be found in the LICENSE file.
//
// Package excelize-py is a Python port of Go Excelize library, providing a set
// of functions that allow you to write and read from XLAM / XLSM / XLSX / XLTM
// / XLTX files. Supports reading and writing spreadsheet documents generated
// by Microsoft Excel™ 2007 and later. Supports complex components by high
// compatibility, and provided streaming API for generating or reading data from
// a worksheet with huge amounts of data. This library needs Python version 3.9
// or later.

package main

/*
#include <types_c.h>
*/
import "C"

import (
	"errors"
	"reflect"
	"sync"
	"time"
	"unsafe"

	"github.com/xuri/excelize/v2"
)

const (
	Nil     C.int = 0
	Int     C.int = 1
	String  C.int = 2
	Float   C.int = 3
	Boolean C.int = 4
	Time    C.int = 5
)

var (
	files      = sync.Map{}
	errNil     string
	errFilePtr = "can not find file pointer"
	errArgType = errors.New("invalid argument data type")

	// goBaseTypes defines Go's basic data types.
	goBaseTypes = map[reflect.Kind]bool{
		reflect.Bool:    true,
		reflect.Int:     true,
		reflect.Int8:    true,
		reflect.Int16:   true,
		reflect.Int32:   true,
		reflect.Int64:   true,
		reflect.Uint:    true,
		reflect.Uint8:   true,
		reflect.Uint16:  true,
		reflect.Uint32:  true,
		reflect.Uint64:  true,
		reflect.Uintptr: true,
		reflect.Float32: true,
		reflect.Float64: true,
		reflect.Map:     true,
		reflect.String:  true,
	}
	// cToBaseGoTypeFuncs defined functions mapping for G to Go basic data types
	// convention.
	cToBaseGoTypeFuncs = map[reflect.Kind]func(cVal reflect.Value, kind reflect.Kind) (reflect.Value, error){
		reflect.Bool: func(cVal reflect.Value, kind reflect.Kind) (reflect.Value, error) {
			return reflect.ValueOf(cVal.Bool()), nil
		},
		reflect.Uint: func(cVal reflect.Value, kind reflect.Kind) (reflect.Value, error) {
			return reflect.ValueOf(uint(cVal.Interface().(C.uint))), nil
		},
		reflect.Uint8: func(cVal reflect.Value, kind reflect.Kind) (reflect.Value, error) {
			return reflect.ValueOf(uint8(cVal.Interface().(C.uchar))), nil
		},
		reflect.Uint64: func(cVal reflect.Value, kind reflect.Kind) (reflect.Value, error) {
			return reflect.ValueOf(uint64(cVal.Interface().(C.uint))), nil
		},
		reflect.Int: func(cVal reflect.Value, kind reflect.Kind) (reflect.Value, error) {
			return reflect.ValueOf(int(cVal.Int())), nil
		},
		reflect.Int64: func(cVal reflect.Value, kind reflect.Kind) (reflect.Value, error) {
			return reflect.ValueOf(cVal.Int()), nil
		},
		reflect.Float64: func(cVal reflect.Value, kind reflect.Kind) (reflect.Value, error) {
			return reflect.ValueOf(float64(cVal.Float())), nil
		},
		reflect.String: func(cVal reflect.Value, kind reflect.Kind) (reflect.Value, error) {
			if cVal.Elem().CanAddr() {
				return reflect.ValueOf(C.GoString(cVal.Interface().(*C.char))), nil
			}
			return reflect.ValueOf(""), nil
		},
	}
	// goBaseValueToCFuncs defined functions mapping for Go basic data types
	// value to C convention.
	goBaseValueToCFuncs = map[reflect.Kind]func(goVal reflect.Value, kind reflect.Kind) (reflect.Value, error){
		reflect.Bool: func(goVal reflect.Value, kind reflect.Kind) (reflect.Value, error) {
			return reflect.ValueOf(C._Bool(goVal.Bool())), nil
		},
		reflect.Uint: func(goVal reflect.Value, kind reflect.Kind) (reflect.Value, error) {
			return reflect.ValueOf(C.uint(uint32(goVal.Uint()))), nil
		},
		reflect.Uint8: func(goVal reflect.Value, kind reflect.Kind) (reflect.Value, error) {
			return reflect.ValueOf(C.uchar(int8(goVal.Uint()))), nil
		},
		reflect.Uint32: func(goVal reflect.Value, kind reflect.Kind) (reflect.Value, error) {
			return reflect.ValueOf(C.uint(uint32(goVal.Uint()))), nil
		},
		reflect.Uint64: func(goVal reflect.Value, kind reflect.Kind) (reflect.Value, error) {
			return reflect.ValueOf(C.ulong(goVal.Uint())), nil
		},
		reflect.Int: func(goVal reflect.Value, kind reflect.Kind) (reflect.Value, error) {
			return reflect.ValueOf(C.int(int32(goVal.Int()))), nil
		},
		reflect.Int32: func(goVal reflect.Value, kind reflect.Kind) (reflect.Value, error) {
			return reflect.ValueOf(C.long(int32(goVal.Int()))), nil
		},
		reflect.Int64: func(goVal reflect.Value, kind reflect.Kind) (reflect.Value, error) {
			return reflect.ValueOf(C.longlong(int64(goVal.Int()))), nil
		},
		reflect.Float64: func(goVal reflect.Value, kind reflect.Kind) (reflect.Value, error) {
			return reflect.ValueOf(C.double(goVal.Float())), nil
		},
		reflect.String: func(goVal reflect.Value, kind reflect.Kind) (reflect.Value, error) {
			return reflect.ValueOf(C.CString(goVal.String())), nil
		},
	}
)

// cToGoBaseType convert JavaScript value to Go basic data type variable.
func cToGoBaseType(cVal reflect.Value, kind reflect.Kind) (reflect.Value, error) {
	fn, ok := cToBaseGoTypeFuncs[kind]
	if !ok {
		return reflect.ValueOf(nil), errArgType
	}
	return fn(cVal, kind)
}

// cToGoArray convert C language array to Go array base on the given Go	structure
// types.
func cToGoArray(cArray reflect.Value, cArrayLen int) reflect.Value {
	switch cArray.Elem().Type().String() {
	case "*main._Ctype_char":
		return reflect.ValueOf(append([]*C.char{}, unsafe.Slice(cArray.Interface().(**C.char), cArrayLen)...))
	case "*main._Ctype_struct_Options": // []*excelize.Options
		val := cArray.Interface().(**C.struct_Options)
		arr := unsafe.Slice(val, cArrayLen)
		return reflect.ValueOf(arr)
	case "main._Ctype_struct_Border":
		val := cArray.Interface().(*C.struct_Border)
		arr := unsafe.Slice(val, cArrayLen)
		return reflect.ValueOf(arr)
	}
	return cArray
}

// cValueToGo convert C language object to Go variable base on the given Go
// structure types, this function extract each fields of the structure from
// object recursively.
func cValueToGo(cVal reflect.Value, goType reflect.Type) (reflect.Value, error) {
	result := reflect.New(goType)
	s := result.Elem()
	for resultFieldIdx := 0; resultFieldIdx < s.NumField(); resultFieldIdx++ {
		field := goType.Field(resultFieldIdx)
		if goBaseTypes[field.Type.Kind()] {
			cBaseVal := cVal.FieldByName(field.Name)
			goBaseVal, err := cToGoBaseType(cBaseVal, field.Type.Kind())
			if err != nil {
				return result, err
			}
			s.Field(resultFieldIdx).Set(goBaseVal.Convert(s.Field(resultFieldIdx).Type()))
			continue
		}
		switch field.Type.Kind() {
		case reflect.Ptr:
			// Pointer of the Go data type, for example: *excelize.Options or *string
			ptrType := field.Type.Elem()
			if !goBaseTypes[ptrType.Kind()] {
				// Pointer of the Go struct, for example: *excelize.Options
				cObjVal := cVal.FieldByName(field.Name)
				if cObjVal.Elem().CanAddr() {
					v, err := cValueToGo(cObjVal.Elem(), ptrType)
					if err != nil {
						return result, err
					}
					s.Field(resultFieldIdx).Set(v)
				}
			}
			if goBaseTypes[ptrType.Kind()] {
				// Pointer of the Go basic data type, for example: *string
				cBaseVal := cVal.FieldByName(field.Name)
				if !cBaseVal.IsNil() {
					v, err := cToGoBaseType(cBaseVal.Elem(), ptrType.Kind())
					if err != nil {
						return result, err
					}
					x := reflect.New(ptrType)
					x.Elem().Set(v)
					s.Field(resultFieldIdx).Set(x.Elem().Addr())
				}
			}
		case reflect.Struct:
			// The Go struct, for example: excelize.Options, convert sub fields recursively
			structType := field.Type
			cObjVal := cVal.FieldByName(field.Name)
			v, err := cValueToGo(cObjVal, structType)
			if err != nil {
				return result, err
			}
			s.Field(resultFieldIdx).Set(v.Elem())
		case reflect.Slice:
			// The Go data type array, for example:
			// []*excelize.Options, []excelize.Options, []string, []*string
			ele := field.Type.Elem()
			cArray := cVal.FieldByName(field.Name)
			if cArray.IsZero() {
				continue
			}
			if ele.Kind() == reflect.Ptr {
				// Pointer array of the Go data type, for example: []*excelize.Options or []*string
				subEle := ele.Elem()
				cArrayLen := int(cVal.FieldByName(field.Name + "Len").Int())
				cArray = cToGoArray(cArray, cArrayLen)
				for i := 0; i < cArray.Len(); i++ {
					if goBaseTypes[subEle.Kind()] {
						// Pointer array of the Go basic data type, for example: []*string
						v, err := cToGoBaseType(cArray.Index(i), subEle.Kind())
						if err != nil {
							return result, err
						}
						x := reflect.New(subEle)
						x.Elem().Set(v)
						s.Field(resultFieldIdx).Set(reflect.Append(s.Field(resultFieldIdx), x.Elem().Addr()))
					} else {
						// Pointer array of the Go struct, for example: []*excelize.Options
						v, err := cValueToGo(cArray.Index(i).Elem(), subEle)
						if err != nil {
							return result, err
						}
						x := reflect.New(subEle)
						x.Elem().Set(v.Elem())
						s.Field(resultFieldIdx).Set(reflect.Append(s.Field(resultFieldIdx), x.Elem().Addr()))
					}
				}
			} else {
				// The Go data type array, for example: []excelize.Options or []string
				subEle := ele
				cArrayLen := int(cVal.FieldByName(field.Name + "Len").Int())
				cArray = cToGoArray(cArray, cArrayLen)
				for i := 0; i < cArray.Len(); i++ {
					if subEle.Kind() == reflect.Uint8 { // []byte
						break
					}
					if goBaseTypes[subEle.Kind()] {
						// The Go basic data type array, for example: []string
						v, err := cToGoBaseType(cArray.Index(i), subEle.Kind())
						if err != nil {
							return result, err
						}

						s.Field(resultFieldIdx).Set(reflect.Append(s.Field(resultFieldIdx), v))
					} else {
						// The Go struct array, for example: []excelize.Options
						v, err := cValueToGo(cArray.Index(i), subEle)
						if err != nil {
							return result, err
						}
						s.Field(resultFieldIdx).Set(reflect.Append(s.Field(resultFieldIdx), v.Elem()))
					}
				}
			}
		}
	}
	return result, nil
}

// goBaseTypeToC convert Go basic data type value to C variable.
func goBaseTypeToC(goVal reflect.Value, kind reflect.Kind) (reflect.Value, error) {
	fn, ok := goBaseValueToCFuncs[kind]
	if !ok {
		return reflect.ValueOf(nil), errors.New("invalid argument data type" + kind.String())
	}
	return fn(goVal, kind)
}

// goValueToC convert Go variable to C object base on the given Go structure
// types, this function extract each fields of the structure from structure
// variable recursively.
func goValueToC(goVal, cVal reflect.Value) (reflect.Value, error) {
	result := cVal
	c := result.Elem()
	for i := 0; i < goVal.Type().NumField(); i++ {
		cField, _ := c.Type().FieldByName(goVal.Type().Field(i).Name)
		field := goVal.Type().Field(i)
		if goBaseTypes[field.Type.Kind()] {
			goBaseVal := goVal.FieldByName(field.Name)
			cBaseVal, err := goBaseTypeToC(goBaseVal, goBaseVal.Type().Kind())
			if err != nil {
				return result, err
			}
			c.FieldByName(field.Name).Set(cBaseVal.Convert(cField.Type))
			continue
		}
		switch goVal.Type().Field(i).Type.Kind() {
		case reflect.Ptr:
			// Pointer of the Go data type, for example: *excelize.Options or *string
			ptrType := field.Type.Elem()
			if !goBaseTypes[ptrType.Kind()] {
				// Pointer of the Go struct, for example: *excelize.Options
				goStructVal := goVal.Field(i)
				if !goStructVal.IsNil() {
					cPtr := C.malloc(C.size_t(cField.Type.Elem().Size()))
					cStructPtr := reflect.NewAt(cField.Type.Elem(), cPtr)
					v, err := goValueToC(goStructVal.Elem(), cStructPtr)
					if err != nil {
						return result, err
					}
					c.FieldByName(field.Name).Set(v)
				}
			}
			if goBaseTypes[ptrType.Kind()] {
				// Pointer of the Go basic data type, for example: *string
				goBaseVal := goVal.Field(i)
				if !goBaseVal.IsNil() {
					v, err := goBaseTypeToC(goBaseVal.Elem(), ptrType.Kind())
					if err != nil {
						return result, err
					}
					cValPtr := C.malloc(C.size_t(unsafe.Sizeof(cField.Type.Elem().Size())))
					ptrVal := reflect.NewAt(v.Type(), cValPtr).Elem()
					ptrVal.Set(v)
					c.FieldByName(field.Name).Set(ptrVal.Addr())
				}
			}
		case reflect.Struct:
			// The Go struct, for example: excelize.Options, convert sub fields recursively
			goStructVal := goVal.Field(i)
			v, err := goValueToC(goStructVal, reflect.New(cField.Type))
			if err != nil {
				return result, err
			}
			c.FieldByName(field.Name).Set(v.Elem())
		case reflect.Slice:
			// The Go data type array, for example:
			// []*excelize.Options, []excelize.Options, []string, []*string
			goSlice := goVal.Field(i)
			ele := goSlice.Type().Elem()
			l, err := goBaseTypeToC(reflect.ValueOf(goSlice.Len()), reflect.Int)
			if err != nil {
				return result, err
			}
			c.FieldByName(field.Name + "Len").Set(l)
			cArray := C.malloc(C.size_t(goSlice.Len()) * C.size_t(cField.Type.Elem().Size()))
			for j := 0; j < goSlice.Len(); j++ {
				if goBaseTypes[ele.Kind()] {
					// The Go basic data type array, for example: []string
					cBaseVal, err := goBaseTypeToC(goSlice.Index(j), goSlice.Index(j).Type().Kind())
					if err != nil {
						return result, err
					}
					elePtr := unsafe.Pointer(uintptr(cArray) + uintptr(j)*cBaseVal.Type().Size())
					ele := reflect.NewAt(cBaseVal.Type(), elePtr).Elem()
					ele.Set(cBaseVal)
				} else {
					// The Go struct array, for example: []excelize.Options
					cPtr := C.malloc(C.size_t(cField.Type.Elem().Size()))
					cStructPtr := reflect.NewAt(cField.Type.Elem(), cPtr)
					v, err := goValueToC(goSlice.Index(j), cStructPtr)
					if err != nil {
						return result, err
					}
					elePtr := unsafe.Pointer(uintptr(cArray) + uintptr(j)*cField.Type.Elem().Size())
					ele := reflect.NewAt(cField.Type.Elem(), elePtr).Elem()
					ele.Set(reflect.NewAt(cField.Type.Elem(), unsafe.Pointer(v.Pointer())).Elem())
				}
			}
			c.FieldByName(field.Name).Set(reflect.NewAt(cField.Type.Elem(), cArray))
		}
	}
	return result, nil
}

// cInterfaceToGo convert C interface to Go interface data type value.
func cInterfaceToGo(val C.struct_Interface) interface{} {
	switch val.Type {
	case Int:
		return int(val.Integer)
	case String:
		return C.GoString(val.String)
	case Float:
		return float64(val.Float64)
	case Boolean:
		return bool(val.Boolean)
	case Time:
		return time.Unix(int64(val.Integer), 0)
	default:
		return nil
	}
}

// Close closes and cleanup the open temporary file for the spreadsheet.
//
//export Close
func Close(idx int) *C.char {
	f, ok := files.Load(idx)
	if !ok {
		return C.CString(errFilePtr)
	}
	defer files.Delete(idx)
	if err := f.(*excelize.File).Close(); err != nil {
		return C.CString(err.Error())
	}
	return C.CString(errNil)
}

//export CopySheet
func CopySheet(idx, from, to int) *C.char {
	f, ok := files.Load(idx)
	if !ok {
		return C.CString(errFilePtr)
	}
	if err := f.(*excelize.File).CopySheet(from, to); err != nil {
		return C.CString(err.Error())
	}
	return C.CString(errNil)
}

//export DeleteChart
func DeleteChart(idx int, sheet, cell *C.char) *C.char {
	f, ok := files.Load(idx)
	if !ok {
		return C.CString(errFilePtr)
	}
	if err := f.(*excelize.File).DeleteChart(C.GoString(sheet), C.GoString(cell)); err != nil {
		return C.CString(err.Error())
	}
	return C.CString(errNil)
}

//export DeleteComment
func DeleteComment(idx int, sheet, cell *C.char) *C.char {
	f, ok := files.Load(idx)
	if !ok {
		return C.CString(errFilePtr)
	}
	if err := f.(*excelize.File).DeleteComment(C.GoString(sheet), C.GoString(cell)); err != nil {
		return C.CString(err.Error())
	}
	return C.CString(errNil)
}

//export DeletePicture
func DeletePicture(idx int, sheet, cell *C.char) *C.char {
	f, ok := files.Load(idx)
	if !ok {
		return C.CString(errFilePtr)
	}
	if err := f.(*excelize.File).DeletePicture(C.GoString(sheet), C.GoString(cell)); err != nil {
		return C.CString(err.Error())
	}
	return C.CString(errNil)
}

//export DeleteSheet
func DeleteSheet(idx int, sheet *C.char) *C.char {
	f, ok := files.Load(idx)
	if !ok {
		return C.CString(errFilePtr)
	}
	if err := f.(*excelize.File).DeleteSheet(C.GoString(sheet)); err != nil {
		return C.CString(err.Error())
	}
	return C.CString(errNil)
}

//export DeleteSlicer
func DeleteSlicer(idx int, name *C.char) *C.char {
	f, ok := files.Load(idx)
	if !ok {
		return C.CString(errFilePtr)
	}
	if err := f.(*excelize.File).DeleteSlicer(C.GoString(name)); err != nil {
		return C.CString(err.Error())
	}
	return C.CString(errNil)
}

// NewFile provides a function to create new file by default template.
//
//export NewFile
func NewFile() int {
	f, idx := excelize.NewFile(), 0
	files.Range(func(_, _ interface{}) bool {
		idx++
		return true
	})
	idx++
	files.Store(idx, f)
	return idx
}

// OpenFile take the name of a spreadsheet file and returns a populated
// spreadsheet file struct for it.
//
//export OpenFile
func OpenFile(filename *C.char, opts *C.struct_Options) C.struct_OptionsResult {
	var options excelize.Options
	if opts != nil {
		goVal, err := cValueToGo(reflect.ValueOf(*opts), reflect.TypeOf(excelize.Options{}))
		if err != nil {
			return C.struct_OptionsResult{idx: C.int(-1), err: C.CString(err.Error())}
		}
		options = goVal.Elem().Interface().(excelize.Options)
	}
	f, err := excelize.OpenFile(C.GoString(filename), options)
	if err != nil {
		return C.struct_OptionsResult{idx: C.int(-1), err: C.CString(err.Error())}
	}
	var idx int
	files.Range(func(_, _ interface{}) bool {
		idx++
		return true
	})
	idx++
	files.Store(idx, f)
	return C.struct_OptionsResult{idx: C.int(idx), err: C.CString(errNil)}
}

// Save provides a function to override the spreadsheet with origin path.
//
//export Save
func Save(idx int, opts *C.struct_Options) *C.char {
	f, ok := files.Load(idx)
	if !ok {
		return C.CString(errFilePtr)
	}
	if opts != nil {
		var options excelize.Options
		goVal, err := cValueToGo(reflect.ValueOf(*opts), reflect.TypeOf(excelize.Options{}))
		if err != nil {
			return C.CString(err.Error())
		}
		options = goVal.Elem().Interface().(excelize.Options)
		if err := f.(*excelize.File).Save(options); err != nil {
			return C.CString(err.Error())
		}
		return C.CString(errNil)
	}
	if err := f.(*excelize.File).Save(); err != nil {
		return C.CString(err.Error())
	}
	return C.CString(errNil)
}

// SaveAs provides a function to create or update to a spreadsheet at the
// provided path.
//
//export SaveAs
func SaveAs(idx int, name *C.char, opts *C.struct_Options) *C.char {
	f, ok := files.Load(idx)
	if !ok {
		return C.CString("")
	}
	if opts != nil {
		var options excelize.Options
		goVal, err := cValueToGo(reflect.ValueOf(*opts), reflect.TypeOf(excelize.Options{}))
		if err != nil {
			return C.CString(err.Error())
		}
		options = goVal.Elem().Interface().(excelize.Options)
		if err := f.(*excelize.File).SaveAs(C.GoString(name), options); err != nil {
			return C.CString(err.Error())
		}
		return C.CString(errNil)
	}
	if err := f.(*excelize.File).SaveAs(C.GoString(name)); err != nil {
		return C.CString(err.Error())
	}
	return C.CString(errNil)
}

//export NewSheet
func NewSheet(idx int, sheet *C.char) C.struct_NewSheetResult {
	f, ok := files.Load(idx)
	if !ok {
		return C.struct_NewSheetResult{idx: C.int(-1), err: C.CString(errFilePtr)}
	}
	idx, err := f.(*excelize.File).NewSheet(C.GoString(sheet))
	if err != nil {
		return C.struct_NewSheetResult{idx: C.int(idx), err: C.CString(err.Error())}
	}
	return C.struct_NewSheetResult{idx: C.int(idx), err: C.CString(errNil)}
}

//export NewStyle
func NewStyle(idx int, style *C.struct_Style) C.struct_NewStyleResult {
	var s excelize.Style
	goVal, err := cValueToGo(reflect.ValueOf(*style), reflect.TypeOf(excelize.Style{}))
	if err != nil {
		return C.struct_NewStyleResult{style: C.int(0), err: C.CString(err.Error())}
	}
	s = goVal.Elem().Interface().(excelize.Style)
	f, ok := files.Load(idx)
	if !ok {
		return C.struct_NewStyleResult{style: C.int(0), err: C.CString(errFilePtr)}
	}
	styleID, err := f.(*excelize.File).NewStyle(&s)
	if err != nil {
		return C.struct_NewStyleResult{style: C.int(styleID), err: C.CString(err.Error())}
	}
	return C.struct_NewStyleResult{style: C.int(styleID), err: C.CString(errNil)}
}

//export SetActiveSheet
func SetActiveSheet(idx, index int) *C.char {
	f, ok := files.Load(idx)
	if !ok {
		return C.CString(errFilePtr)
	}
	f.(*excelize.File).SetActiveSheet(index)
	return C.CString(errNil)
}

//export SetCellStyle
func SetCellStyle(idx int, sheet, topLeftCell, bottomRightCell *C.char, styleID int) *C.char {
	f, ok := files.Load(idx)
	if !ok {
		return C.CString(errFilePtr)
	}
	if err := f.(*excelize.File).SetCellStyle(C.GoString(sheet), C.GoString(topLeftCell), C.GoString(bottomRightCell), styleID); err != nil {
		return C.CString(err.Error())
	}
	return C.CString(errNil)
}

//export SetCellValue
func SetCellValue(idx int, sheet, cell *C.char, value *C.struct_Interface) *C.char {
	f, ok := files.Load(idx)
	if !ok {
		return C.CString(errFilePtr)
	}
	if err := f.(*excelize.File).SetCellValue(C.GoString(sheet), C.GoString(cell), cInterfaceToGo(*value)); err != nil {
		return C.CString(err.Error())
	}
	return C.CString(errNil)
}

//export GetStyle
func GetStyle(idx, styleID int) C.struct_GetStyleResult {
	f, ok := files.Load(idx)
	if !ok {
		return C.struct_GetStyleResult{err: C.CString(errFilePtr)}
	}
	style, err := f.(*excelize.File).GetStyle(styleID)
	if err != nil {
		return C.struct_GetStyleResult{err: C.CString(err.Error())}
	}
	cVal, err := goValueToC(reflect.ValueOf(*style), reflect.ValueOf(&C.struct_Style{}))
	if err != nil {
		return C.struct_GetStyleResult{err: C.CString(err.Error())}
	}
	return C.struct_GetStyleResult{style: cVal.Elem().Interface().(C.struct_Style), err: C.CString(errNil)}
}

//export SetSheetBackgroundFromBytes
func SetSheetBackgroundFromBytes(idx int, sheet, extension *C.char, picture *C.uchar, pictureLen C.int) *C.char {
	f, ok := files.Load(idx)
	if !ok {
		return C.CString(errFilePtr)
	}
	buf := C.GoBytes(unsafe.Pointer(picture), pictureLen)
	if err := f.(*excelize.File).SetSheetBackgroundFromBytes(C.GoString(sheet), C.GoString(extension), buf); err != nil {
		C.CString(err.Error())
	}
	return C.CString(errNil)
}

func main() {
}
