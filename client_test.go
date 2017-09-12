package ovo

import (
    "crypto/hmac"
    "crypto/sha256"
    "fmt"
    "regexp"
    "testing"
)

var TestDomainMap = map[string]string{
    "customer_profile":               "/customers/:customer_id",                                                   //GET
    "calculate_points":               "/customers/:customer_id/points",                                            //PUT
    "pushtopay_transaction":          "/customers/:customer_id/transactions",                                      //POST
    "pushtopay_transaction_status":   "/customers/:customer_id/transactions/:transaction_id",                      //GET
    "pushtopay_void_transaction":     "/customers/:customer_id/transactions/:transaction_id",                      //PUT
    "customer_profile_qr":            "/merchants/:merchant_id/stores/:store_id/terminals/:terminal_id/customers", //GET
    "customer_linkage":               "/customers/:customer_id",                                                   //POST
    "customer_authentication":        "/authentications",                                                          // POST
    "customer_authentication_status": "/authentications/:authentication_id",                                       //GET
}

//TestGetURL : Testing domain map availability
func TestGetURL(t *testing.T) {
    client := New("", "", "", "")
    randNumber := "123456"

    for k := range TestDomainMap {
        re := regexp.MustCompile(":[a-zA-Z0-9_]+")
        vars := re.FindAllString(TestDomainMap[k], -1)

        switch len(vars) {
        case 1:
            if _, err := client.getURL(k, randNumber); err != nil {
                t.Errorf(k + " domain map is missing")
            }
            if _, err := client.getURL(k, randNumber, randNumber); err.Error() != TErr("ovo_unidentified_request", client.LocaleID).Error() {
                t.Errorf(k + "params number must be invalid")
            }
        case 2:
            if _, err := client.getURL(k, randNumber, randNumber); err != nil {
                t.Errorf(k + " domain map is missing")
            }
            if _, err := client.getURL(k, randNumber); err.Error() != TErr("ovo_unidentified_request", client.LocaleID).Error() {
                t.Errorf(k + "params number must be invalid")
            }
        case 3:
            if _, err := client.getURL(k, randNumber, randNumber, randNumber); err != nil {
                t.Errorf(k + " domain map is missing")
            }
            if _, err := client.getURL(k, randNumber, randNumber); err.Error() != TErr("ovo_unidentified_request", client.LocaleID).Error() {
                t.Errorf(k + "params number must be invalid")
            }
        }
    }
}

func TestAuthorizationKey(t *testing.T) {
    client := &Client{
        APIKey: "56789",
        AppID:  "01234",
        Random: "POKOKNYA_INI_RANDOM",
    }
    client.setAuthorizationKey()

    stringToSign := []byte(client.AppID + client.Random)
    h := hmac.New(sha256.New, []byte(client.APIKey))
    h.Write(stringToSign)
    TestHmac := fmt.Sprintf("%x", h.Sum(nil))

    if TestHmac != client.Hmac {
        t.Errorf("Invalid Authorization Key Type")
    }
}

func TestNewRequest(t *testing.T) {
    client := new(Client)
    methods := []string{"POST", "PUT"}
    for _, v := range methods {
        httpReq, err := client.newRequest(v, "http://testing.com", nil)
        if err != nil {
            t.Errorf(err.Error())
        }
        if httpReq.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
            t.Errorf("Invalid content type")
        }
    }

    httpReq, err := client.newRequest("GET", "http://testing.com", nil)
    if err != nil {
        t.Errorf(err.Error())
    }

    header := httpReq.Header

    if _, ok := header["App-Id"]; !ok {
        t.Errorf("app-id not found")
    }
    if _, ok := header["Random"]; !ok {
        t.Errorf("random not found")
    }
    if _, ok := header["Hmac"]; !ok {
        t.Errorf("hmac not found")
    }

}

func TestGetResponse(t *testing.T) {
    client := new(Client)
    _, err := client.getResponse([]byte("[]"))
    if err != nil {
        t.Errorf("this should not be error")
    }
}
