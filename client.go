package ovo

import (
    "bytes"
    "crypto/hmac"
    "crypto/sha256"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/http/httptest"
    "net/url"
    "regexp"
    "strings"
    "time"
)

//New : Constructor for OVO Client / App
func New(baseURL, apiKey, appID, merchantID string) *Client {
    random := time.Now().Format("20060102150405")

    c := &Client{
        BaseURL:    baseURL,
        APIKey:     apiKey,
        AppID:      appID,
        MerchantID: merchantID,
        Random:     random,
        LocaleID:   "en",
    }

    c.setAuthorizationKey()

    return c

}

//SetLocale : Setting locale for translation
func (client *Client) SetLocale(localeID string) {
    client.LocaleID = localeID
}

func (client *Client) setErrMessage(data map[string]map[string]string) {
    ErrMessage = data
}

func (client *Client) setAuthorizationKey() {
    stringToSign := []byte(client.AppID + client.Random)
    h := hmac.New(sha256.New, []byte(client.APIKey))
    h.Write(stringToSign)
    client.Hmac = fmt.Sprintf("%x", h.Sum(nil))
}

func (client *Client) newRequest(method string, url string, body *bytes.Buffer) (*http.Request, error) {

    var req *http.Request
    var err error
    if body == nil {
        req, err = http.NewRequest(method, url, nil)
    } else {
        req, err = http.NewRequest(method, url, body)
    }
    if err != nil {
        return nil, err
    }

    //Override request
    if client.httpHandler != nil {
        if body == nil {
            req = httptest.NewRequest(method, url, nil)
        } else {
            req = httptest.NewRequest(method, url, body)
        }
    }

    if method == "POST" || method == "PUT" {
        req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    }

    req.Header.Add("app-id", client.AppID)
    req.Header.Add("random", client.Random)
    req.Header.Add("hmac", client.Hmac)

    //Setup ResponseWriter recorder
    if client.httpHandler != nil {
        client.wRes = httptest.NewRecorder()
        client.httpHandler(client.wRes, req)
    }

    return req, nil
}

func (client *Client) sendRequest(request *http.Request) (response *http.Response, data []byte, err error) {

    if client.httpHandler != nil {
        response = client.wRes.Result()
    } else {
        response, err = http.DefaultClient.Do(request)
    }

    if err != nil {
        err = TErr("ovo_unavailable_service", client.LocaleID)
        return
    }

    if response.StatusCode >= http.StatusInternalServerError {
        err = TErr("ovo_unavailable_service", client.LocaleID)
        return
    }

    buf := &bytes.Buffer{}
    _, err = io.Copy(buf, response.Body)
    response.Body.Close()
    if err != nil {
        err = TErr("ovo_invalid_response", client.LocaleID)
    }

    data = buf.Bytes()
    return
}

func (client *Client) execRequest(method string, url string, body *bytes.Buffer) (data []byte, err error) {

    req, errReq := client.newRequest(method, url, body)

    if errReq != nil {
        return nil, errReq
    }

    _, data, errResp := client.sendRequest(req)

    if errResp != nil {
        return nil, errResp
    }
    return data, nil
}

func (client *Client) createParams(params Params) *bytes.Buffer {
    form := url.Values{}

    for k, v := range params {
        form.Add(k, v)
    }

    buf := &bytes.Buffer{}
    buf.WriteString(form.Encode())

    return buf
}

func (client *Client) getURL(name string, params ...string) (string, error) {

    re := regexp.MustCompile(":[a-zA-Z0-9_]+")
    vars := re.FindAllString(domainMap[name], -1)
    if len(params) != len(vars) {
        return "", TErr("ovo_unidentified_request", client.LocaleID)
    }

    path, ok := domainMap[name]
    if !ok {
        return "", TErr("ovo_unidentified_request", client.LocaleID)
    }
    counter := 0
    for _, v := range vars {
        path = strings.Replace(path, v, params[counter], 2)
        counter = counter + 1
    }

    url := client.BaseURL + path
    return url, nil
}

func (client *Client) getResponse(data []byte) (Response, error) {
    var r Response
    err := json.Unmarshal(data, &r)

    if err != nil {
        //Enforce data to OvoResponseData, because of inconsistent data type
        if _, ok := err.(*json.UnmarshalTypeError); ok {
            r.Data = ResponseData{}
            return r, nil
        }
        return r, err
    }

    return r, nil

}
