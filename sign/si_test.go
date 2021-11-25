package sign

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"golang.org/x/crypto/sha3"
	"math/big"
	"testing"
	"fmt"
)

var EncryptData string
var CipherKey string

func TestSignGatewayEvmChain(t *testing.T) {
	siReq := SigReqData{
		To:    "a929022c9107643515f5c777ce9a910f0d1e490c",
		ToTag: "dd9b86c1000000000000000000000000a71edc38d189767582c38a3145b5873052c3e47a0000000000000000000000000000000000000000000422ca8b0a00a425000000000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000000286139633032303462313062626131306666636534383864636536666666663163616364626262313000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000036574680000000000000000000000000000000000000000000000000000000000",
		Nonce: 13,
		Decimal: 18,
		From:     "606288c605942f3c84a7794c0b3257b56487263c",
		//GasLimit 2000000
		FeeStep: "2000000",
		//GasPrice 1.2*suggestGasprice, or 150Gwei by default
		FeePrice: "15000000000",
		Amount:   "0",
		TaskType: "withdraw",
	}
	auReq := BusData{
		Chain: "heco",
		Quantity: "101000000000000000000",
		ToAddress: "a9c0204b10bba10ffce488dce6ffff1cacdbbb10",
		ToTag: "dd9b86c1000000000000000000000000a71edc38d189767582c38a3145b5873052c3e47a0000000000000000000000000000000000000000000422ca8b0a00a425000000000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000000286139633032303462313062626131306666636534383864636536666666663163616364626262313000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000036574680000000000000000000000000000000000000000000000000000000000",
	}

	appId := "rebal-si-gateway"

	signReq := SignReq{
		SiReq: siReq,
		AuReq: auReq,
	}

	resp, err := SignGatewayEvmChain(signReq, appId)
	if err != nil{
		t.Error(err)
	}

	t.Log(resp.Data.EncryptData)
	t.Log(resp.Data.Extra.Cipher)

	EncryptData  = resp.Data.EncryptData
	CipherKey = resp.Data.Extra.Cipher

	fmt.Println(resp)

	fmt.Println("EncryptData")
	fmt.Println(EncryptData)

	fmt.Println("CipherKey")
	fmt.Println(CipherKey)
}


func TestPostAuditInfo(t *testing.T) {
	req := AuditRequest{
		BusType:"starsHecoBridgeWithdraw",
		BusStep: 1,
		BusId: "15",
		BusData: BusData{
			Chain: "heco",
			Quantity: "101000000000000000000",
			ToAddress: "a9c0204b10bba10ffce488dce6ffff1cacdbbb10",
			ToTag: "dd9b86c1000000000000000000000000a71edc38d189767582c38a3145b5873052c3e47a0000000000000000000000000000000000000000000422ca8b0a00a425000000000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000000286139633032303462313062626131306666636534383864636536666666663163616364626262313000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000036574680000000000000000000000000000000000000000000000000000000000",
		},
		Result: 1,
	}

	appId := "rebal-si-gateway"

	resp, err := PostAuditInfo(req, appId)
	if err != nil {
		t.Error(err)
	}
	t.Log(resp.Code)
	fmt.Println(resp)
}


func TestValidator(t *testing.T) {
	appId := "rebal-si-gateway"
	req := ValidReq{
		Id: 15,
		Platform: "starshecobridge",
		Chain: "ht2",
		EncryptData: EncryptData,
		CipherKey: CipherKey,
	}

	fmt.Println(req)
	fmt.Println(appId)

	resp, err := Validator(req, appId)
	if err != nil{
		t.Error(err)
	}
	t.Log(resp.OK)
	fmt.Println(resp)
}

//发起跨链交易到签名机
func TestSendToBridge(t *testing.T) {
	//1.sign
    fromAddress := "606288c605942f3c84a7794c0b3257b56487263c"
	toAddress := common.HexToAddress("a9c0204b10bba10ffce488dce6ffff1cacdbbb10")  // 和receiver一致
	receiverAddress := "a9c0204b10bba10ffce488dce6ffff1cacdbbb10"  //brigde address

	//首先构造 inputdata
	sendToBridgeSignature := []byte("sendToBridge(address,uint256，uint256)")

	hash := sha3.NewLegacyKeccak256()
	hash.Write(sendToBridgeSignature)
	methodID := hash.Sum(nil)[:4]
	fmt.Println(hexutil.Encode(methodID))//0x0f75dc8c

	paddedAddress := common.LeftPadBytes(toAddress.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAddress)) // 0x000000000000000000000000a929022c9107643515f5c777ce9a910f0d1e490c

	amount := new(big.Int)
	amount.SetString("1000000000000000000000", 10) // sets the value to 1000 tokens, in the token denomination

	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAmount)) // 0x00000000000000000000000000000000000000000000003635c9adc5dea00000

	taskId := new(big.Int)
	taskId.SetString("19", 10) // sets the value to 1000 tokens, in the token denomination

	paddlesId := common.LeftPadBytes(taskId.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddlesId)) //0x0000000000000000000000000000000000000000000000000000000000000010

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)
	data = append(data, paddlesId...)

	str := hexutil.Encode(data)
	fmt.Println(str)

	//然后将data的0x去除
	toTag := str[2:]

	fmt.Println(toTag)

	siReq := SigReqData{
		To:    receiverAddress,
		ToTag: toTag,
		Decimal: 18,
		From:     fromAddress,
		FeeStep: "2000000",
		FeePrice: "15000000000",
		Amount:   "0",
		TaskType: "withdraw",
	}
	auReq := BusData{
		Chain: "heco",
		Quantity: "1000000000000000000000",
		ToAddress: receiverAddress,
		ToTag: toTag,
	}

	appId := "rebal-si-gateway"

	signReq := SignReq{
		SiReq: siReq,
		AuReq: auReq,
	}

	signResp, err := SignGatewayEvmChain(signReq, appId)
	if err != nil{
		t.Error(err)
	}

	t.Log(signResp.Data.EncryptData)
	t.Log(signResp.Data.Extra.Cipher)

	fmt.Println(signResp)

	//2.audit
	auditReq := AuditRequest{
		BusType:"starsHecoBridgeWithdraw",
		BusStep: 1,
		BusId: "20",
		BusData: BusData{
			Chain: "heco",
			Quantity: "1000000000000000000000",
			ToAddress: "a9c0204b10bba10ffce488dce6ffff1cacdbbb10",
			ToTag: toTag,
		},
		Result: 1,
	}

	auditResp, err := PostAuditInfo(auditReq, appId)
	if err != nil {
		t.Error(err)
	}
	t.Log(auditResp.Code)
	fmt.Println(auditResp)

	//3.validator
	vReq := ValidReq{
		Id: 20,
		Platform: "starshecobridge",
		Chain: "ht2",
		EncryptData: signResp.Data.EncryptData,
		CipherKey: signResp.Data.Extra.Cipher,
	}

	fmt.Println(vReq)

	valResp, err := Validator(vReq, appId)
	if err != nil{
		t.Error(err)
	}
	t.Log(valResp.OK)
	fmt.Println(valResp)
}

//接收跨链资产接口
func TestReceiveFromBridge(t *testing.T) {
	//1.sign
	fromAddress := "606288c605942f3c84a7794c0b3257b56487263c"
	receiverAddress := "a9c0204b10bba10ffce488dce6ffff1cacdbbb10"  //brigde address

	//首先构造 inputdata
	receiveFromBridgeSignature := []byte("receiveFromBridge(uint256，uint256)")

	hash := sha3.NewLegacyKeccak256()
	hash.Write(receiveFromBridgeSignature)
	methodID := hash.Sum(nil)[:4]
	fmt.Println(hexutil.Encode(methodID)) //0xecb56f69

	amount := new(big.Int)
	amount.SetString("1000000000000000000000", 10) // sets the value to 1000 tokens, in the token denomination

	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddedAmount)) //0x00000000000000000000000000000000000000000000003635c9adc5dea00000

	taskId := new(big.Int)
	taskId.SetString("17", 10) // sets the value to 1000 tokens, in the token denomination

	paddlesId := common.LeftPadBytes(taskId.Bytes(), 32)
	fmt.Println(hexutil.Encode(paddlesId)) //0x0000000000000000000000000000000000000000000000000000000000000011

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAmount...)
	data = append(data, paddlesId...)

	str := hexutil.Encode(data)
	fmt.Println(str)

	//然后将data的0x去除
	toTag := str[2:]

	fmt.Println(toTag)

	siReq := SigReqData{
		To:    receiverAddress,
		ToTag: toTag,
		Decimal: 18,
		From:     fromAddress,
		FeeStep: "2000000",
		FeePrice: "15000000000",
		Amount:   "0",
		TaskType: "withdraw",
	}
	auReq := BusData{
		Chain: "heco",
		Quantity: "1000000000000000000000",
		ToAddress: receiverAddress,
		ToTag: toTag,
	}

	appId := "rebal-si-gateway"

	signReq := SignReq{
		SiReq: siReq,
		AuReq: auReq,
	}

	signResp, err := SignGatewayEvmChain(signReq, appId)
	if err != nil{
		t.Error(err)
	}

	t.Log(signResp.Data.EncryptData)
	t.Log(signResp.Data.Extra.Cipher)

	fmt.Println(signResp)

	//2.audit
	auditReq := AuditRequest{
		BusType:"starsHecoBridgeWithdraw",
		BusStep: 1,
		BusId: "18",
		BusData: BusData{
			Chain: "heco",
			Quantity: "1000000000000000000000",
			ToAddress: "a9c0204b10bba10ffce488dce6ffff1cacdbbb10",
			ToTag: toTag,
		},
		Result: 1,
	}

	auditResp, err := PostAuditInfo(auditReq, appId)
	if err != nil {
		t.Error(err)
	}
	t.Log(auditResp.Code)
	fmt.Println(auditResp)

	//3.validator
	vReq := ValidReq{
		Id: 18,
		Platform: "starshecobridge",
		Chain: "ht2",
		EncryptData: signResp.Data.EncryptData,
		CipherKey: signResp.Data.Extra.Cipher,
	}

	fmt.Println(vReq)

	valResp, err := Validator(vReq, appId)
	if err != nil{
		t.Error(err)
	}
	t.Log(valResp.OK)
	fmt.Println(valResp)
}






