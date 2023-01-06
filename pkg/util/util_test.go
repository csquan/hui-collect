package util

import (
	"encoding/base64"
	"fmt"
	"testing"
	"user/pkg/conf"
)

func Test_Hashed(t *testing.T) {
	conf.Conf.AesSalt = "EUETM5UVAVOMKTQHMBO74BBVJHUYTMVL"
	txt := "EUETM5UVAVOMKTQHMBO74BBVJHUYTMVL"
	hashed, err := HashSecret(txt)
	if err != nil {
		t.Error(err)
	}
	length := len(hashed)
	if length < 20 {
		t.Error("lll")
	}
	restore, err := DecodeSecret(hashed)
	if err != nil {
		t.Error(err)
	}
	if restore != txt {
		t.Error("no eq")
	}
	fmt.Println(restore)
}

func TestBcrypt(t *testing.T) {
	txt := "EUETM5UVAVOMKTQHMBO74BBVJHUYTMVasdfasdfL"
	hashed, err := HashPassword(txt)
	if err != nil {
		t.Error(err)
	}
	length := len(hashed)
	if length != 60 {
		t.Error("not 60 long")
	}
	fmt.Println(length)
	fmt.Println(hashed)
	b64 := base64.StdEncoding.EncodeToString([]byte(txt))
	fmt.Println(b64)
	plainByte, er := base64.StdEncoding.DecodeString(b64)
	if er != nil {
		t.Error(err)
	}
	if err := VerifyPassword(hashed, string(plainByte)); err != nil {
		t.Error(err)
	}

}

func TestDecodeSecret(t *testing.T) {
	sec := "B+9iSyA0dKnQvrNqMRh62QJF3nDAeLOlxmOh8eUjrVzlqfoz1Ayz4+rmoxHNBP6iY6tq3jaxDHaDVBm9"
	decodedStoredSecret, err := base64.StdEncoding.DecodeString(sec)
	if err != nil {
		t.Fatal(err)
	}
	AesSalt := "EUETM5UVAVOMKTQHMBO74BBVJHUYTMVL"
	k := CreateSha2(AesSalt)

	secretBytes, er := AesGcmDec(k[:], decodedStoredSecret)
	if er != nil {
		t.Fatal(er)
	}
	fmt.Println(string(secretBytes))

}
