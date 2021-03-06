# Copyright 2020 Motoreq Infotech Pvt Ltd

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

#     http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


export GOPATH=/media/sf_Blockchain/medisot/gitrepo/gopath
rm vendor/ -rf
mkdir -p vendor/github.com/prasanths96/iState/
# cp ../../vendor . -rf
cp ../../*.go vendor/github.com/prasanths96/iState
rm $GOPATH/src/github.com/prasanths96/iState -rf
mkdir -p $GOPATH/src/github.com/prasanths96/iState
cp vendor/github.com/prasanths96/iState/* $GOPATH/src/github.com/prasanths96/iState/. -rf
cp $GOPATH/src/github.com/emirpasic/ vendor/github.com/. -rf
cp $GOPATH/src/github.com/bluele/ vendor/github.com/. -rf
cp $GOPATH/src/github.com/prasanths96/gostack vendor/github.com/prasanths96/. -rf
# rm $GOPATH/src/github.com/prasanths96/hyperledger/easycompositestate -rf
# mkdir -p $GOPATH/src/github.com/prasanths96/hyperledger/easycompositestate
# cp vendor/github.com/prasanths96/hyperledger/easycompositestate/* $GOPATH/src/github.com/prasanths96/hyperledger/easycompositestate/. -rf
# rm $GOPATH/src/github.com/prasanths96/hyperledger/querylib -rf
# mkdir -p $GOPATH/src/github.com/prasanths96/hyperledger/querylib
# cp vendor/github.com/prasanths96/hyperledger/querylib/* $GOPATH/src/github.com/prasanths96/hyperledger/querylib/. -rf
go build