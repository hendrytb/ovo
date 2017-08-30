package ovo

import (
    "bytes"
    "crypto/hmac"
    "crypto/sha256"
    "database/sql"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"
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
    }

    c.setAuthorizationKey()

    return c

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

    if method == "POST" || method == "PUT" {
        req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    }

    req.Header.Add("app-id", client.AppID)
    req.Header.Add("random", client.Random)
    req.Header.Add("hmac", client.Hmac)

    return req, nil
}

func (client *Client) sendRequest(request *http.Request) (response *http.Response, data []byte, err error) {
    response, err = http.DefaultClient.Do(request)

    if err != nil {
        err = fmt.Errorf("cannot reach server. %v", err)
        return
    }

    buf := &bytes.Buffer{}
    _, err = io.Copy(buf, response.Body)
    response.Body.Close()

    if err != nil {
        err = fmt.Errorf("cannot read response. %v", err)
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
        return "", errors.New("Incomplete parameters request")
    }

    path, ok := domainMap[name]
    if !ok {
        return "", errors.New("Unknown url name")
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
        if v, ok := err.(*json.UnmarshalTypeError); ok {
            if v.Field == "Data" {
                r.Data = ResponseData{}
                return r, nil
            }
        }
        return r, err
    }

    return r, nil

}

//GetMMsdk : Get Matahari Mall sdk
func (client *Client) GetMMsdk(db *sql.DB) *MatahariMall {
    return &MatahariMall{
        DB:  db,
        API: client,
    }
}
