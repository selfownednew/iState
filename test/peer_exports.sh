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

export CORE_VM_ENDPOINT=unix:///host/var/run/docker.sock
export CORE_VM_DOCKER_HOSTCONFIG_NETWORKMODE=${COMPOSE_PROJECT_NAME}_fabricnetwork
export CORE_LOGGING_LEVEL=DEBUG
export CORE_PEER_TLS_ENABLED=true
export CORE_PEER_ENDORSER_ENABLED=true
export CORE_PEER_GOSSIP_USELEADERELECTION=false
export CORE_PEER_GOSSIP_ORGLEADER=true
export CORE_PEER_PROFILE_ENABLED=true
export CORE_PEER_TLS_CERT_FILE=/etc/hyperledger/fabric/tls/server.crt
export CORE_PEER_TLS_KEY_FILE=/etc/hyperledger/fabric/tls/server.key
export CORE_PEER_TLS_ROOTCERT_FILE=/etc/hyperledger/fabric/tls/ca.crt
# export CORE_PEER_DISCOVERY_ENABLED=false
export CORE_CHAINCODE_EXECUTETIMEOUT=300s
export CORE_VM_DOCKER_HOSTCONFIG_MEMORY=5368709120‬
export FABRIC_SDK_DISCOVERY=false
export GODEBUG=netdns=go

export CORE_PEER_ID=peer0.users.medisot.com
export CORE_PEER_ADDRESS=peer0.users.medisot.com:7051
export CORE_PEER_CHAINCODELISTENADDRESS=peer0.users.medisot.com:7052
export CORE_PEER_GOSSIP_EXTERNALENDPOINT=peer0.users.medisot.com:7051
export CORE_PEER_LOCALMSPID=medisotUsersMSP
# export CORE_LEDGER_STATE_STATEDATABASE=CouchDB
# export CORE_LEDGER_STATE_COUCHDBCONFIG_COUCHDBADDRESS=couch0:5984
export CORE_PEER_GOSSIP_BOOTSTRAP=peer0.users.medisot.com:7051
export CORE_LEDGER_STATE_TOTALQUERYLIMIT=1000000
# export CORE_PEER_PROFILE_ENABLED=true
# export CORE_PEER_PROFILE_LISTENADDRESS=0.0.0.0:6060