# Demo for User-Managed Access on Hyperledger Fabric

デモの使い方について説明する．
登場主体は以下の5種類である．
* Authorization Blockchain - Authorization Serverに相当
* FL-Server - Resource Serverに相当
* FL-Submitter - Resource Ownerに相当
* FL-Requestor - Requesting Partyに相当
* FL-Client - Clientに相当

本デモでは，Authorization Blockchain，FL-Server，及びFL-Clientの3種のエンティティをそれぞれ異なるWebサーバ上に実装する．
FL-Submitter及びFL-RequestorはWebブラウザ使用者が担当する．（一人でもできるし，複数人でもできる．）

# Authorization Blockchain

バージョン情報
* Hyperledger Fabric 2.1.0
* Go 1.13.10
* DOcker Engine 19.03.12
* python 3.8.5
* flask 1.1.2

Hyperledger Fabricネットワークを立ち上げ，各チェーンコードをインストールする．
Hyperledger Fabricのサンプルディレクトリ fabric-samples は ~/bcauth-authz_server/ に配置する．

```bash
# ネットワーク立ち上げ
cd ~/bcauth-authz_server/fabric-samples/test-network;
./network.sh down;
./network.sh up createChannel;

# path を設定
export PATH=${PWD}/../bin:$PATH;
export FABRIC_CFG_PATH=$PWD/../config/;

# GO 関連の設定（意味は知らない）
cd ~/bcauth-authz_server/fabric-samples/chaincode/uma/v2_timestamp/timestamp;
GO111MODULE=on go mod vendor;
cd ~/bcauth-authz_server/fabric-samples/chaincode/uma/v2_timestamp/pat;
GO111MODULE=on go mod vendor;
cd ~/bcauth-authz_server/fabric-samples/chaincode/uma/v2_timestamp/rreg;
GO111MODULE=on go mod vendor;
cd ~/bcauth-authz_server/fabric-samples/chaincode/uma/v2_timestamp/policy;
GO111MODULE=on go mod vendor;
cd ~/bcauth-authz_server/fabric-samples/chaincode/uma/v2_timestamp/perm;
GO111MODULE=on go mod vendor;
cd ~/bcauth-authz_server/fabric-samples/chaincode/uma/v2_timestamp/token;
GO111MODULE=on go mod vendor;
cd ~/bcauth-authz_server/fabric-samples/chaincode/uma/v2_timestamp/claim;
GO111MODULE=on go mod vendor;
cd ~/bcauth-authz_server/fabric-samples/chaincode/uma/v2_timestamp/intro;
GO111MODULE=on go mod vendor;

# package
cd ~/bcauth-authz_server/fabric-samples/test-network;
export PACKAGE_CC_NAME="timestamp";
peer lifecycle chaincode package $PACKAGE_CC_NAME.tar.gz --path ../chaincode/uma/v2_timestamp/$PACKAGE_CC_NAME/ --lang golang --label ${PACKAGE_CC_NAME}_1.0;
export PACKAGE_CC_NAME="pat";
peer lifecycle chaincode package $PACKAGE_CC_NAME.tar.gz --path ../chaincode/uma/v2_timestamp/$PACKAGE_CC_NAME/ --lang golang --label ${PACKAGE_CC_NAME}_1.0;
export PACKAGE_CC_NAME="rreg";
peer lifecycle chaincode package $PACKAGE_CC_NAME.tar.gz --path ../chaincode/uma/v2_timestamp/$PACKAGE_CC_NAME/ --lang golang --label ${PACKAGE_CC_NAME}_1.0;
export PACKAGE_CC_NAME="policy";
peer lifecycle chaincode package $PACKAGE_CC_NAME.tar.gz --path ../chaincode/uma/v2_timestamp/$PACKAGE_CC_NAME/ --lang golang --label ${PACKAGE_CC_NAME}_1.0;
export PACKAGE_CC_NAME="perm";
peer lifecycle chaincode package $PACKAGE_CC_NAME.tar.gz --path ../chaincode/uma/v2_timestamp/$PACKAGE_CC_NAME/ --lang golang --label ${PACKAGE_CC_NAME}_1.0;
export PACKAGE_CC_NAME="token";
peer lifecycle chaincode package $PACKAGE_CC_NAME.tar.gz --path ../chaincode/uma/v2_timestamp/$PACKAGE_CC_NAME/ --lang golang --label ${PACKAGE_CC_NAME}_1.0;
export PACKAGE_CC_NAME="claim";
peer lifecycle chaincode package $PACKAGE_CC_NAME.tar.gz --path ../chaincode/uma/v2_timestamp/$PACKAGE_CC_NAME/ --lang golang --label ${PACKAGE_CC_NAME}_1.0;
export PACKAGE_CC_NAME="intro";
peer lifecycle chaincode package $PACKAGE_CC_NAME.tar.gz --path ../chaincode/uma/v2_timestamp/$PACKAGE_CC_NAME/ --lang golang --label ${PACKAGE_CC_NAME}_1.0;

# コマンド操作を Org1MSP に設定
export CORE_PEER_TLS_ENABLED=true;
export CORE_PEER_LOCALMSPID="Org1MSP";
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt;
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp;
export CORE_PEER_ADDRESS=localhost:7051;

# instgall (Org1MSP)
cd ~/bcauth-authz_server/fabric-samples/test-network;
export INSTALL_CC_NAME="timestamp";
peer lifecycle chaincode install $INSTALL_CC_NAME.tar.gz;
export INSTALL_CC_NAME="pat";
peer lifecycle chaincode install $INSTALL_CC_NAME.tar.gz;
export INSTALL_CC_NAME="rreg";
peer lifecycle chaincode install $INSTALL_CC_NAME.tar.gz;
export INSTALL_CC_NAME="policy";
peer lifecycle chaincode install $INSTALL_CC_NAME.tar.gz;
export INSTALL_CC_NAME="perm";
peer lifecycle chaincode install $INSTALL_CC_NAME.tar.gz;
export INSTALL_CC_NAME="token";
peer lifecycle chaincode install $INSTALL_CC_NAME.tar.gz;
export INSTALL_CC_NAME="claim";
peer lifecycle chaincode install $INSTALL_CC_NAME.tar.gz;
export INSTALL_CC_NAME="intro";
peer lifecycle chaincode install $INSTALL_CC_NAME.tar.gz;

# コマンド操作を Org2MSP に設定
export CORE_PEER_LOCALMSPID="Org2MSP";
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt;
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp;
export CORE_PEER_ADDRESS=localhost:9051;

# install (Org2MSP)
export INSTALL_CC_NAME="timestamp";
peer lifecycle chaincode install $INSTALL_CC_NAME.tar.gz;
export INSTALL_CC_NAME="pat";
peer lifecycle chaincode install $INSTALL_CC_NAME.tar.gz;
export INSTALL_CC_NAME="rreg";
peer lifecycle chaincode install $INSTALL_CC_NAME.tar.gz;
export INSTALL_CC_NAME="policy";
peer lifecycle chaincode install $INSTALL_CC_NAME.tar.gz;
export INSTALL_CC_NAME="perm";
peer lifecycle chaincode install $INSTALL_CC_NAME.tar.gz;
export INSTALL_CC_NAME="token";
peer lifecycle chaincode install $INSTALL_CC_NAME.tar.gz;
export INSTALL_CC_NAME="claim";
peer lifecycle chaincode install $INSTALL_CC_NAME.tar.gz;
export INSTALL_CC_NAME="intro";
peer lifecycle chaincode install $INSTALL_CC_NAME.tar.gz;

# queryinstalled - package_id を確認
peer lifecycle chaincode queryinstalled;

# approveformyorg (Org2MSP)
export CC_PACKAGE_NAME="timestamp";
export CC_PACKAGE_ID=timestamp_1.0:c6c6f738635d7b373d58ff0431ee266af9c54af61b9abd5b882230d42d70ac25;
peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name $CC_PACKAGE_NAME --version 1.0 --package-id $CC_PACKAGE_ID --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem;
export CC_PACKAGE_NAME="pat";
export CC_PACKAGE_ID=pat_1.0:481bfe4fad1e33e54cd97cc5ff24526bd8d0ebd02365c37d337580c0dcdb28ed;
peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name $CC_PACKAGE_NAME --version 1.0 --package-id $CC_PACKAGE_ID --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem;
export CC_PACKAGE_NAME="rreg";
export CC_PACKAGE_ID=rreg_1.0:b2654a24039036b0b9caa1373f471a89101903b6987dd92f2a67222b49406eb5;
peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name $CC_PACKAGE_NAME --version 1.0 --package-id $CC_PACKAGE_ID --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem;
export CC_PACKAGE_NAME="policy";
export CC_PACKAGE_ID=policy_1.0:8be6bbe519998f5b4e6e64ad8df1c08d5eeead7f6b6cd126c4c011656b6e2384;
peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name $CC_PACKAGE_NAME --version 1.0 --package-id $CC_PACKAGE_ID --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem;
export CC_PACKAGE_NAME="perm";
export CC_PACKAGE_ID=perm_1.0:2fda651b8521e6aa6b97681cd406d654ce018bf4c3edad480a6dd79ce613e6c0;
peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name $CC_PACKAGE_NAME --version 1.0 --package-id $CC_PACKAGE_ID --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem;
export CC_PACKAGE_NAME="token";
export CC_PACKAGE_ID=token_1.0:11deaa41a6f5a0ab7181169d1615f5ec9a9111e1e843a552cddc69455c68dd2b;
peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name $CC_PACKAGE_NAME --version 1.0 --package-id $CC_PACKAGE_ID --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem;
export CC_PACKAGE_NAME="claim";
export CC_PACKAGE_ID=claim_1.0:bfada39280d9e637af31211689b4e419187fd14d2788c837a308a0261fc27edc;
peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name $CC_PACKAGE_NAME --version 1.0 --package-id $CC_PACKAGE_ID --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem;
export CC_PACKAGE_NAME="intro";
export CC_PACKAGE_ID=intro_1.0:b194a401988b87d7f5dba803021779948b22785d3da3167482e2a5f3b8916ae6;
peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name $CC_PACKAGE_NAME --version 1.0 --package-id $CC_PACKAGE_ID --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem;

# コマンド操作を Org1MSP に設定
export CORE_PEER_LOCALMSPID="Org1MSP";
export CORE_PEER_MSPCONFIGPATH=${PWD}/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp;
export CORE_PEER_TLS_ROOTCERT_FILE=${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt;
export CORE_PEER_ADDRESS=localhost:7051;

# approveformyorg (Org1MSP)
export CC_PACKAGE_NAME="timestamp";
export CC_PACKAGE_ID=timestamp_1.0:c6c6f738635d7b373d58ff0431ee266af9c54af61b9abd5b882230d42d70ac25;
peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name $CC_PACKAGE_NAME --version 1.0 --package-id $CC_PACKAGE_ID --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem;
export CC_PACKAGE_NAME="pat";
export CC_PACKAGE_ID=pat_1.0:481bfe4fad1e33e54cd97cc5ff24526bd8d0ebd02365c37d337580c0dcdb28ed;
peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name $CC_PACKAGE_NAME --version 1.0 --package-id $CC_PACKAGE_ID --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem;
export CC_PACKAGE_NAME="rreg";
export CC_PACKAGE_ID=rreg_1.0:b2654a24039036b0b9caa1373f471a89101903b6987dd92f2a67222b49406eb5;
peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name $CC_PACKAGE_NAME --version 1.0 --package-id $CC_PACKAGE_ID --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem;
export CC_PACKAGE_NAME="policy";
export CC_PACKAGE_ID=policy_1.0:8be6bbe519998f5b4e6e64ad8df1c08d5eeead7f6b6cd126c4c011656b6e2384;
peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name $CC_PACKAGE_NAME --version 1.0 --package-id $CC_PACKAGE_ID --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem;
export CC_PACKAGE_NAME="perm";
export CC_PACKAGE_ID=perm_1.0:2fda651b8521e6aa6b97681cd406d654ce018bf4c3edad480a6dd79ce613e6c0;
peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name $CC_PACKAGE_NAME --version 1.0 --package-id $CC_PACKAGE_ID --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem;
export CC_PACKAGE_NAME="token";
export CC_PACKAGE_ID=token_1.0:11deaa41a6f5a0ab7181169d1615f5ec9a9111e1e843a552cddc69455c68dd2b;
peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name $CC_PACKAGE_NAME --version 1.0 --package-id $CC_PACKAGE_ID --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem;
export CC_PACKAGE_NAME="claim";
export CC_PACKAGE_ID=claim_1.0:bfada39280d9e637af31211689b4e419187fd14d2788c837a308a0261fc27edc;
peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name $CC_PACKAGE_NAME --version 1.0 --package-id $CC_PACKAGE_ID --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem;
export CC_PACKAGE_NAME="intro";
export CC_PACKAGE_ID=intro_1.0:b194a401988b87d7f5dba803021779948b22785d3da3167482e2a5f3b8916ae6;
peer lifecycle chaincode approveformyorg -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name $CC_PACKAGE_NAME --version 1.0 --package-id $CC_PACKAGE_ID --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem;

# checkcommitreadiness - チャネルメンバーの approve 状況を確認できる．
peer lifecycle chaincode checkcommitreadiness --channelID mychannel --name intro --version 1.0 --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --output json;

# commit
export COMMIT_CC_NAME="timestamp";
peer lifecycle chaincode commit -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name $COMMIT_CC_NAME --version 1.0 --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt;
export COMMIT_CC_NAME="pat";
peer lifecycle chaincode commit -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name $COMMIT_CC_NAME --version 1.0 --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt;
export COMMIT_CC_NAME="rreg";
peer lifecycle chaincode commit -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name $COMMIT_CC_NAME --version 1.0 --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt;
export COMMIT_CC_NAME="policy";
peer lifecycle chaincode commit -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name $COMMIT_CC_NAME --version 1.0 --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt;
export COMMIT_CC_NAME="perm";
peer lifecycle chaincode commit -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name $COMMIT_CC_NAME --version 1.0 --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt;
export COMMIT_CC_NAME="token";
peer lifecycle chaincode commit -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name $COMMIT_CC_NAME --version 1.0 --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt;
export COMMIT_CC_NAME="claim";
peer lifecycle chaincode commit -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name $COMMIT_CC_NAME --version 1.0 --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt;
export COMMIT_CC_NAME="intro";
peer lifecycle chaincode commit -o localhost:7050 --ordererTLSHostnameOverride orderer.example.com --channelID mychannel --name $COMMIT_CC_NAME --version 1.0 --sequence 1 --tls --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem --peerAddresses localhost:7051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt --peerAddresses localhost:9051 --tlsRootCertFiles ${PWD}/organizations/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt;

# queryinstalled
peer lifecycle chaincode querycommitted --channelID mychannel --name timestamp --cafile ${PWD}/organizations/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem;
```

上記の作業が終了したら，アプリケーションサーバを立ち上げる．
※さくらサーバ上では ~/bcauth-authz_server は ~/project-bcauth に対応する．

```bash
cd ~/bcauth-authz_server/flask
python app.py
```

# FL-Server

バージョン情報
* python 3.6.9
* flask 1.1.2

アプリケーションサーバを立ち上げる．
※さくらサーバ上では， ~/bcauth-fl_server は ~/tff/sample に対応する．
```bash
cd ~/bcauth-fl_server
python app.py
```

# FL-Client
バージョン情報
* python 3.8.5
* flask 1.1.2
* tensorflow 2.4.0
* tensorflow-federated-nightly 0.17.0.dev20201218

アプリケーションサーバを立ち上げる．
※さくらサーバ上では， ~/bcauth-client-backend は ~/project-bcauth に対応する．
```bash
cd ~/bcauth-client-backend/flask
python app.py
```
