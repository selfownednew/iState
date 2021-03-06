/*
	Copyright 2020 Motoreq Infotech Pvt Ltd

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

package istate

import (
	"encoding/json"
	"fmt"
	"github.com/bluele/gcache"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"hash/crc64"
	"reflect"
	"sync"
)

type iState struct {
	mux               sync.Mutex
	structRef         interface{}
	fieldJSONIndexMap map[string]int
	jsonFieldKindMap  map[string]reflect.Kind
	mapKeyKindMap     map[string]reflect.Kind
	depthKindMap      map[string]reflect.Kind
	primaryIndex      int
	docsCounter       map[string]int
	istateJSONMap     map[string]string

	CompactionSize int

	// Cache
	kvCache   gcache.Cache
	hashTable *crc64.Table

	currentStub *shim.ChaincodeStubInterface
}

// Options is an input parameter for NewiState function.
// CacheSize determines the max no. of records allowed in memory.
// DefaultCompactionSize determines how many index records to compact at one go.
// Since Compaction process takes lot of time, it is useful to set a batchsize and
// Compaction can be run multiple times, where each time, n records gets compacted.
type Options struct {
	CacheSize             int
	DefaultCompactionSize int
}

// NewiState function is used to create an iState interface for a struct type.
// Further CRUD operation on the struct type can be performed using the interface returned.
// It takes an empty struct value, Options as input and returns istate interface.
func NewiState(object interface{}, opt Options) (iStateInterface Interface, iStateErr Error) {
	iStateLogger.Infof("Inside NewiState")
	defer iStateLogger.Infof("Exiting NewiState")

	filledRef := fillZeroValue(object)
	// A map of JSON fieldname & it's position in the struct
	fieldJSONIndexMap := make(map[string]int)
	jsonFieldKindMap := make(map[string]reflect.Kind)
	mapKeyKindMap := make(map[string]reflect.Kind)
	depthKindMap := make(map[string]reflect.Kind)
	iStateErr = generateFieldJSONIndexMap(filledRef, fieldJSONIndexMap)
	if iStateErr != nil {
		return
	}
	iStateErr = generatejsonFieldKindMap(filledRef, jsonFieldKindMap, mapKeyKindMap)
	if iStateErr != nil {
		return
	}
	iStateErr = generateDepthKindMap(filledRef, depthKindMap)
	if iStateErr != nil {
		return
	}
	var i int
	i, iStateErr = getPrimaryFieldIndex(filledRef)
	if iStateErr != nil {
		return
	}

	var istateJSONMap map[string]string
	istateJSONMap, iStateErr = generateistateJSONMap(object, fieldJSONIndexMap)
	if iStateErr != nil {
		return
	}

	docsCounter := make(map[string]int)
	iStateIns := &iState{
		structRef:         filledRef,
		fieldJSONIndexMap: fieldJSONIndexMap,
		jsonFieldKindMap:  jsonFieldKindMap,
		mapKeyKindMap:     mapKeyKindMap,
		depthKindMap:      depthKindMap,
		primaryIndex:      i,
		docsCounter:       docsCounter,
		CompactionSize:    opt.DefaultCompactionSize,
		hashTable:         crc64.MakeTable(crc64.ISO),
		istateJSONMap:     istateJSONMap,
	}

	// Cache
	kvCache := gcache.New(opt.CacheSize).
		ARC().
		LoaderFunc(iStateIns.loader).
		Build()

	iStateIns.kvCache = kvCache
	iStateInterface = iStateIns

	fmt.Println("=============================================================")
	fmt.Println("depthKindMap", depthKindMap)
	fmt.Println("=============================================================")
	fmt.Println("istateJSONMAP: ", istateJSONMap)
	fmt.Println("=============================================================")
	return
}

// CreateState function is used to create a new state in the state db.
// It is a function in iState Interface and must be called via the Interface
// returned by NewiState function.
// It takes chaincode stub and the actual structure value as input params.
// Note: This function does not do state validations such as, checking whether state exists or not
// before performing the operation.
func (iState *iState) CreateState(stub shim.ChaincodeStubInterface, object interface{}) (iStateErr Error) {
	iStateLogger.Infof("Inside CreateState")
	defer iStateLogger.Infof("Exiting CreateState")

	iState.setStub(&stub)

	if reflect.TypeOf(object) != reflect.TypeOf(iState.structRef) {
		iStateErr = newError(nil, 1014, reflect.TypeOf(iState.structRef), reflect.TypeOf(object))
		return
	}

	// Find primary key
	keyref := iState.getPrimaryKey(object)

	mo, err := json.Marshal(object)
	if err != nil {
		iStateErr = newError(err, 1002)
		return
	}

	var oMap map[string]interface{}
	err = json.Unmarshal(mo, &oMap)
	if err != nil {
		iStateErr = newError(err, 1003)
		return
	}

	hashString := iState.hash(mo)
	encodedKeyValPairs, docsCounter, _, iStateErr := iState.encodeState(oMap, keyref, hashString)
	if iStateErr != nil {
		return
	}

	// Main key - value
	encodedKeyValPairs[keyref] = mo

	for key, val := range encodedKeyValPairs {
		err = stub.PutState(key, val)
		if err != nil {
			iStateErr = newError(err, 1001)
			return
		}
	}

	for index, count := range docsCounter {
		iState.incDocsCounter(index, count)
	}

	iStateErr = iState.setCache(keyref, object, mo, hashString)
	if iStateErr != nil {
		return
	}

	return nil
}

// ReadState function is used to read a state from state db.
// It is a function in iState Interface and must be called via the Interface
// returned by NewiState function.
// It takes chaincode stub and value of primary key as input params.
// It returns the actual structure value as an interface{}. The returned value can
// be type asserted to the actual struct type before using.
// Note: This function does not do state validations such as, checking whether state exists or not
// before performing the operation.
func (iState *iState) ReadState(stub shim.ChaincodeStubInterface, primaryKey interface{}) (uObj interface{}, iStateErr Error) {
	iStateLogger.Infof("Inside ReadState")
	defer iStateLogger.Infof("Exiting ReadState")

	iState.setStub(&stub)

	primaryKeyString := fmt.Sprintf("%v", primaryKey)

	stateBytes, err := stub.GetState(primaryKeyString)
	if err != nil {
		iStateErr = newError(err, 1005)
	}

	if stateBytes == nil {
		return
	}

	uObjReflect, iStateErr := iState.unmarshalToStruct(stateBytes)
	if iStateErr != nil {
		return
	}

	uObj = uObjReflect.Interface()

	return
}

// UpdateState function is used to update a state from statedb.
// It is a function in iState Interface and must be called via the Interface
// returned by NewiState function.
// It takes chaincode stub and the actual structure value as input params.
// Note: This function does not do state validations such as, checking whether state exists or not
// before performing the operation.
func (iState *iState) UpdateState(stub shim.ChaincodeStubInterface, object interface{}) (iStateErr Error) {
	iStateLogger.Infof("Inside UpdateState")
	defer iStateLogger.Infof("Exiting UpdateState")

	iState.setStub(&stub)

	if reflect.TypeOf(object) != reflect.TypeOf(iState.structRef) {
		iStateErr = newError(nil, 1015, reflect.TypeOf(iState.structRef), reflect.TypeOf(object))
		return
	}

	// Find primary key
	keyref := iState.getPrimaryKey(object)

	stateBytes, err := stub.GetState(keyref)
	if err != nil {
		iStateErr = newError(err, 1017)
	}

	if stateBytes == nil {
		return iState.CreateState(stub, object)
	}

	var source map[string]interface{}
	err = json.Unmarshal(stateBytes, &source)
	if err != nil {
		iStateErr = newError(err, 1007)
		return
	}

	mo, err := json.Marshal(object)
	if err != nil {
		iStateErr = newError(err, 1006)
		return
	}

	var target map[string]interface{}
	err = json.Unmarshal(mo, &target)
	if err != nil {
		iStateErr = newError(err, 1007)
		return
	}

	appendOrModifyMap, deleteMap, iStateErr := iState.findDifference(source, target)
	if iStateErr != nil {
		return
	}

	if len(appendOrModifyMap) == 0 && len(deleteMap) == 0 {
		iStateErr = newError(nil, 1004)
		return
	}

	hashString := iState.hash(mo)
	// Delete first, so that TestStruct_aMap_user1 (map key) does not get deleted.
	// When updating same key of a map, the above key gets over-writted at first,
	// then when deleting, the over-written key gets deleted.
	deleteEncodedKeyValPairs, delDocsCounter, _, iStateErr := iState.encodeState(deleteMap, keyref, hashString)
	if iStateErr != nil {
		return
	}
	appendEncodedKeyValPairs, addDocsCounter, _, iStateErr := iState.encodeState(target, keyref, hashString)
	if iStateErr != nil {
		return
	}
	// Main key - value
	appendEncodedKeyValPairs[keyref] = mo

	for key := range deleteEncodedKeyValPairs {
		err = stub.DelState(key)
		if err != nil {
			iStateErr = newError(err, 1008)
			return
		}
	}
	for key, val := range appendEncodedKeyValPairs {
		err = stub.PutState(key, val)
		if err != nil {
			iStateErr = newError(err, 1009)
			return
		}
	}

	for index, count := range delDocsCounter {
		iState.decDocsCounter(index, count)
	}
	for index, count := range addDocsCounter {
		iState.incDocsCounter(index, count)
	}

	iStateErr = iState.setCache(keyref, object, mo, hashString)
	if iStateErr != nil {
		return
	}

	return nil
}

// DeleteState function is used to delete a state from state db.
// It is a function in iState Interface and must be called via the Interface
// returned by NewiState function.
// It takes chaincode stub and value of primary key as input params.
// Note: This function does not do state validations such as, checking whether state exists or not
// before performing the operation.
func (iState *iState) DeleteState(stub shim.ChaincodeStubInterface, primaryKey interface{}) (iStateErr Error) {
	iStateLogger.Infof("Inside DeleteState")
	defer iStateLogger.Infof("Exiting DeleteState")

	iState.setStub(&stub)

	keyref := fmt.Sprintf("%v", primaryKey)

	stateBytes, err := stub.GetState(keyref)
	if err != nil {
		iStateErr = newError(err, 1018)
	}

	if stateBytes == nil {
		// If state does not exist, return silently
		//iStateErr = newError(nil, 1013, keyref)
		return
	}

	var source map[string]interface{}
	err = json.Unmarshal(stateBytes, &source)
	if err != nil {
		iStateErr = newError(err, 1011)
		return
	}

	// Delete first, so that TestStruct_aMap_user1 (map key) does not get deleted.
	// When updating same key of a map, the above key gets over-writted at first,
	// then when deleting, the over-written key gets deleted.
	encodedKeyValPairs, delDocsCounter, _, iStateErr := iState.encodeState(source, keyref, "")
	if iStateErr != nil {
		return
	}
	// Main key - value
	encodedKeyValPairs[keyref] = stateBytes

	for key := range encodedKeyValPairs {
		err = stub.DelState(key)
		if err != nil {
			iStateErr = newError(err, 1012)
			return
		}
	}

	for index, count := range delDocsCounter {
		iState.decDocsCounter(index, count)
	}

	return nil
}

//
func (iState *iState) CompactIndex(stub shim.ChaincodeStubInterface) (iStateErr Error) {
	iStateLogger.Infof("Inside CompactIndex")
	defer iStateLogger.Infof("Exiting CompactIndex")

	iState.setStub(&stub)

	uObj, iStateErr := convertObjToMap(iState.structRef)
	if iStateErr != nil {
		return
	}

	keyRef := ""
	encodedKV, _, _, iStateErr := iState.encodeState(uObj, keyRef, "", 2) // separate key & value = 2, query = false
	if iStateErr != nil {
		return
	}

	compactedIndexMap := make(map[string]compactIndexV) // compacted index

	for index := range encodedKV {
		startKey := index
		endKey := index + asciiLast

		var kvMap map[string][]byte // original index
		kvMap, iStateErr = getKeyByRange(stub, startKey, endKey, iState.CompactionSize)
		if iStateErr != nil {
			return
		}

		alreadyFetched := make(map[string]struct{})
		for origIndexK, hashBytes := range kvMap {
			compactIndex, keyRef := generateCIndexKey(origIndexK)
			if compactIndex == "" {
				continue
			}
			var cIndexVal compactIndexV
			var oldCIndexVal compactIndexV
			switch val, ok := compactedIndexMap[compactIndex]; !ok {
			case true:
				if _, ok := alreadyFetched[compactIndex]; !ok {
					cIndexVal, iStateErr = fetchCompactIndex(stub, compactIndex)
					alreadyFetched[compactIndex] = struct{}{}
					oldCIndexVal = cIndexVal
				}
			default:
				cIndexVal = val
			}
			if len(cIndexVal) == 0 {
				cIndexVal = make(compactIndexV)
			}
			cIndexVal[keyRef] = string(hashBytes)
			if !reflect.DeepEqual(oldCIndexVal, cIndexVal) {
				compactedIndexMap[compactIndex] = cIndexVal
				// Delete original index key
				err := stub.DelState(origIndexK)
				if err != nil {
					iStateErr = newError(err, 1016)
					return
				}
			}
		}

	}

	iStateErr = putCompactIndex(stub, compactedIndexMap)
	if iStateErr != nil {
		return
	}

	return
}
