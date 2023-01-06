package web

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/go-resty/resty/v2"
	"net/http"
	"testing"
	"time"
	"user/pkg/util/ecies"
)

const KycPubKey = "026a38b8aa47f2af2d2163253ff5385cd059f90476990c7e3ee84bca0ead322241"

func TestKycUpdate(t *testing.T) {
	pubKey, err := ecies.PublicFromString(KycPubKey)
	if err != nil {
		t.Fatal(err)
	}

	cli := resty.New()
	cli.SetBaseURL("http://localhost:8000")

	nowStr := time.Now().UTC().Format(http.TimeFormat)
	ct, err := ecies.Encrypt(rand.Reader, pubKey, []byte(nowStr), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	data := map[string]interface{}{
		"bid":       "569923450059",
		"firm-name": "一地鸡毛蒜皮小公司",
		"firm-type": 2,
		"country":   "+86",
		"verified":  hex.EncodeToString(ct),
	}
	var result HttpData
	resp, er := cli.R().SetBody(data).SetResult(&result).Post("/api/v1/pub/kyc-ok")
	if er != nil {
		t.Fatal(er)
	}
	if resp.StatusCode() != http.StatusOK {
		t.Fatal("not 200")
	}
	if result.Code != 0 {
		t.Fatal(result)
	}
}

func TestKycUserList(t *testing.T) {
	pubKey, err := ecies.PublicFromString(KycPubKey)
	if err != nil {
		t.Fatal(err)
	}

	cli := resty.New()
	cli.SetBaseURL("http://localhost:8000")

	nowStr := time.Now().UTC().Format(http.TimeFormat)
	ct, err := ecies.Encrypt(rand.Reader, pubKey, []byte(nowStr), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	data := map[string]interface{}{
		"verified": hex.EncodeToString(ct),
		"start":    time.Now().AddDate(0, 0, -100).Unix(),
		"end":      time.Now().Unix(),
		"page":     2,
		"limit":    11,
	}
	var result HttpData
	resp, er := cli.R().SetBody(data).SetResult(&result).Post("/api/v1/pub/kyc-user-list")
	if er != nil {
		t.Fatal(er)
	}
	if resp.StatusCode() != http.StatusOK {
		t.Fatal("not 200")
	}
	if result.Code != 0 {
		t.Fatal(result)
	}
	fmt.Println(result.Data)
}
