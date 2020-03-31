// Copyright 2020 <>. All rights reserved.

package istate

import (
	"fmt"
)

const errorPrefix = "iState Error: "

// errorCodes stores a map of Error Code to Error message.
var errorCodes = map[int]string{
	0: "Error Test",

	1001: "CreateState: PutStateError",
	1002: "Marshal error at SaveState",
	1003: "Unmarshal error at SaveState",
	1004: "UpdateState: Found no change with update",
	1005: "ReadState: GetStateError",
	1006: "Marshal error at UpdateState",
	1007: "Unmarshal error at UpdateState",
	1008: "UpdateState: DelStateError",
	1009: "UpdateState: PutState",
	1010: "Marshal error at DeleteState",
	1011: "Unmarshal error at DeleteState",
	1012: "DeleteState: DelStateError",
	1013: "DeleteState: Key does not exist: %s",
	1014: "CreateState: Struct with incorrect type passed: Expected - %v, Received - %v",
	1015: "UpdateState: Struct with incorrect type passed: Expected - %v, Received - %v",

	// Util errors
	2001: "json tag is not set for struct field: %s of type: %v",
	2002: "primary tag is not set for any field in struct type: %v (or) primary field value is empty",
	2003: "Marshal error at encodeState",
	2004: "Unmarshal error at encodeState",
	2005: "encode: Unsupported kind: %v",
	2006: "Marshal error at generateRelationalTables",
	2007: "Unmarshal error at generateRelationalTables",
	2008: "generateRelationalTables: Pointer of structure expected. Received: %v",
	2009: "encode: Integer overflow. Number with digits: %d",
	2010: "findDifference: Unsupported kind received: %v",
	2011: "findSliceDifference: Source and Target type are not same, Received: %v %v",
	2012: "findSliceDifference: Only slice kind is expected, Received: %v %v",
	2013: "findMapDifference: Source and Target type are not same, Received: %v %v",
	2014: "findMapDifference: Only map kind is expected, Received: %v %v",
	2015: "traverseAndGenerateRelationalTable: currentdepth: %v, not found in depthKindMap",

	// Query
	3001: "Query: Marshal Error:",
	3002: "Query: Unmarshal Error:",
	3003: "parseAndEvalSingle: Syntax error: Field value in query must be of type string. Received: %v",
	3004: "parseAndEvalSingle: Syntax error: <Space> not found in %v",
	3005: "Syntax error: Unknown keyword: %v",
	3006: "Evaluate: GetStateByRange Error",
	3007: "Query Error: getStateByRange: Iterator.Next()",
	3008: "Query Error: getStateByRange: GetState()",
	3009: "convertToRightType: ParseBool Error",
	3010: "convertToRightType: Empty index received",
	3011: "convertToRightType: Atoi Error",
	3012: "convertToRightType: ParseInt Error",
	3013: "convertToRightType: ParseUint Error",
	3014: "convertToRightType: ParseFloat Error",
	3015: "convertToRightType: Indexdot overflow",
	3016: "convertToRightType: Field not found in jsonFieldKindMap: %v",
	3017: "convertToPrimitiveType: Invalid primitive type received: %v",
}

// Error is the interface of this package.
type Error interface {
	Error() string
	GetCode() int
}

// iStateError is the error type of this package.
type iStateError struct {
	Err  string
	Code int
}

// Error function is defined to let Error implement error interface.
func (err iStateError) Error() string {
	return err.Err
}

// Error function is defined to let Error implement error interface.
func (err iStateError) GetCode() int {
	return err.Code
}

// NewError function is to create Errors in a more readable way.
func NewError(err error, code int, params ...interface{}) Error {
	iStateLogger.Debugf("Inside NewError")
	defer iStateLogger.Debugf("Exiting NewError")
	msg := fmt.Sprintf("%d: ", code) + fmt.Sprintf(errorCodes[code], params...)
	if err != nil {
		msg = errorPrefix + msg + ": " + err.Error()
	}
	return iStateError{
		Err:  msg,
		Code: code,
	}
}