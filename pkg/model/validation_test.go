package model

import (
	"fmt"
	"testing"
)

func TestMobileInput_Validate(t *testing.T) {
	mi := MobileInput{"+1 978 355 8888"}
	err := mi.Validate()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(mi.Number())
	mi = MobileInput{"+86 188 88 88 8 88 8"}
	err = mi.Validate()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(mi.Number())
}

func TestSubStruct(t *testing.T) {
	vmi := VerifyMobileInput{
		MobileInput: MobileInput{Mobile: "123"},
		Code:        "123123",
	}
	mi := vmi.MobileInput
	if mi.Mobile != vmi.Mobile {
		t.Error("not eq")
	}
}
