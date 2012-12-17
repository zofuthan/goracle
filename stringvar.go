// Copyright 2012-2013 Tamás Gulácsi
// See LICENSE.txt
// Translated from cx_Oracle ((c) Anthony Tuininga) by Tamás Gulácsi
package goracle

/*
#cgo CFLAGS: -I/usr/include/oracle/11.2/client64
#cgo LDFLAGS: -lclntsh -L/usr/lib/oracle/11.2/client64/lib

#include <stdlib.h>
#include <oci.h>
*/
import "C"

import (
	// "unsafe"
	"errors"
	"fmt"
)

// Initialize the variable.
func stringVar_Initialize(v *Variable, cur *Cursor) error {
	v.actualLength = make([]C.ub2, v.allocatedElements)
	return nil
}

// Set the value of the variable.
func stringVar_SetValue(v *Variable, pos uint, value interface{}) error {
	var (
		text   string
		buf    []byte
		ok     bool
		length int
	)
	if text, ok = value.(string); !ok {
		if buf, ok = value.([]byte); !ok {
			return fmt.Errorf("string or []byte required, got %T", value)
		} else {
			if v.typ.isCharData {
				text = string(buf)
				length = len(text)
			} else {
				length = len(buf)
			}
		}
	} else {
		if v.typ.isCharData {
			length = len(text)
		} else {
			length = len(buf)
		}
		buf = []byte(text)
	}
	if v.typ.isCharData && length > MAX_STRING_CHARS {
		return errors.New("string data too large")
	} else if !v.typ.isCharData && length > MAX_BINARY_BYTES {
		return errors.New("binary data too large")
	}

	// ensure that the buffer is large enough
	if length > v.bufferSize {
		if err := v.resize(uint(length)); err != nil {
			return err
		}
	}

	// keep a copy of the string
	v.actualLength[pos] = C.ub2(length)
	if length > 0 {
		copy(v.data[v.bufferSize*int(pos):], buf)
	}

	return nil
}

// Returns the value stored at the given array position.
func stringVar_GetValue(v *Variable, pos uint) (interface{}, error) {
	buf := v.data[v.bufferSize*int(pos) : v.bufferSize*int(pos)+int(v.actualLength[pos])]
	if v.typ == BinaryVarType {
		return buf, nil
	}
	return v.environment.FromEncodedString(buf), nil
	/*
		#if PY_MAJOR_VERSION < 3
		    if (var->type == &vt_FixedNationalChar
		            || var->type == &vt_NationalCharString)
		        return PyUnicode_Decode(data, var->actualLength[pos],
		                var->environment->nencoding, NULL);
		#endif
	*/
}

/*
#if PY_MAJOR_VERSION < 3
//-----------------------------------------------------------------------------
// StringVar_PostDefine()
//   Set the character set information when values are fetched from this
// variable.
//-----------------------------------------------------------------------------
static int StringVar_PostDefine(
    udt_StringVar *var)                 // variable to initialize
{
    sword status;

    status = OCIAttrSet(var->defineHandle, OCI_HTYPE_DEFINE,
            &var->type->charsetForm, 0, OCI_ATTR_CHARSET_FORM,
            var->environment->errorHandle);
    if (Environment_CheckForError(var->environment, status,
            "StringVar_PostDefine(): setting charset form") < 0)
        return -1;

    return 0;
}
#endif
*/

// Returns the buffer size to use for the variable.
func stringVar_GetBufferSize(v *Variable) int {
	if v.typ.isCharData {
		return v.size * v.environment.maxBytesPerCharacter
	}
	return v.size
}

var StringVarType = &VariableType{Id: 0,
	isVariableLength: true,
	initialize:       stringVar_Initialize,
	setValue:         stringVar_SetValue,
	getValue:         stringVar_GetValue,
	getBufferSize:    stringVar_GetBufferSize,
	oracleType:       C.SQLT_CHR,       // Oracle type
	charsetForm:      C.SQLCS_IMPLICIT, // charset form
	size:             MAX_STRING_CHARS, // element length (default)
	isCharData:       true,             // is character data
	canBeCopied:      true,             // can be copied
	canBeInArray:     true,             // can be in array
}

var FixedCharVarType = &VariableType{
	initialize:       stringVar_Initialize,
	setValue:         stringVar_SetValue,
	getValue:         stringVar_GetValue,
	getBufferSize:    stringVar_GetBufferSize,
	oracleType:       C.SQLT_AFC,       // Oracle type
	charsetForm:      C.SQLCS_IMPLICIT, // charset form
	size:             2000,             // element length (default)
	isCharData:       true,             // is character data
	isVariableLength: true,             // is variable length
	canBeCopied:      true,             // can be copied
	canBeInArray:     true,             // can be in array
}

var RowidVarType = &VariableType{
	initialize:       stringVar_Initialize,
	setValue:         stringVar_SetValue,
	getValue:         stringVar_GetValue,
	getBufferSize:    stringVar_GetBufferSize,
	oracleType:       C.SQLT_CHR,       // Oracle type
	charsetForm:      C.SQLCS_IMPLICIT, // charset form
	size:             18,               // element length (default)
	isCharData:       true,             // is character data
	isVariableLength: false,            // is variable length
	canBeCopied:      true,             // can be copied
	canBeInArray:     true,             // can be in array
}

var BinaryVarType = &VariableType{
	initialize:       stringVar_Initialize,
	setValue:         stringVar_SetValue,
	getValue:         stringVar_GetValue,
	oracleType:       C.SQLT_BIN,       // Oracle type
	charsetForm:      C.SQLCS_IMPLICIT, // charset form
	size:             MAX_BINARY_BYTES, // element length (default)
	isCharData:       false,            // is character data
	isVariableLength: true,             // is variable length
	canBeCopied:      true,             // can be copied
	canBeInArray:     true,             // can be in array
}