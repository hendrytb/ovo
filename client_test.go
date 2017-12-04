package ovo

import (
    "crypto/hmac"
    "crypto/sha256"
    "fmt"
    "net/http"
    "regexp"
    "testing"
    "time"
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
    random := time.Now().Format("20060102150405")
    client := &Client{
        APIKey: "084b13ecac81e1a8caf1775ad02bd5fa40e7219c8956dba11429a497a0e4cd89",
        AppID:  "hypermart",
        Random: random,
    }

    client.setAuthorizationKey()

    stringToSign := []byte(client.AppID + client.Random)
    h := hmac.New(sha256.New, []byte(client.APIKey))
    h.Write(stringToSign)
    TestHmac := fmt.Sprintf("%x", h.Sum(nil))
    fmt.Println(random)
    fmt.Println(TestHmac)
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

func TestExecRequest503(t *testing.T) {

    client := new(Client)
    client.LocaleID = "en"
    client.httpHandler = func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusServiceUnavailable)
        //w.Write([]byte("<html><body>Hello World!</body></html>"))
        //io.WriteString(w, "<html><body>Hello World!</body></html>")
    }

    _, err := client.execRequest("GET", "http://apapunitu.com", nil)
    if err == nil {
        t.Errorf("Should error when service 503")
    }
    if err != nil && err.Error() != TErr("ovo_unavailable_service", client.LocaleID).Error() {
        t.Errorf("Should return Err: " + TErr("ovo_unavailable_service", client.LocaleID).Error())
    }

}
