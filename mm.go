package ovo

import (
    "database/sql"
    "encoding/json"
    "errors"
    "fmt"
    "log"
    "net/http"
    "regexp"
    "strings"
    "time"
)

//GetMMsdk : Get Matahari Mall sdk
func (client *Client) GetMMsdk(db *sql.DB) *MatahariMall {
    return &MatahariMall{
        DB:  db,
        API: client,
    }
}

func (c *MatahariMall) parsePhoneNumber(ovoReq *Request) error {
    re := regexp.MustCompile(PhoneValidRegex)
    x := re.MatchString(ovoReq.Phone)
    if x {
        ovoReq.Phone = strings.Replace(ovoReq.Phone, "+", "", 2)
    } else {
        return TErr("ovo_id_invalid", c.API.LocaleID)
    }

    if ovoReq.Phone[0:2] == "62" {
        ovoReq.Phone = "0" + ovoReq.Phone[2:]
    }

    return nil
}

func (c *MatahariMall) getOvoInfoFromStorage(ovoReq *Request) error {
    cOvo := CustomerOvo{}

    var ovoID sql.NullString

    q := `SELECT customer_id,
                   ovo_id,
                   ovo_phone,
                   ovo_auth_id,
                   fg_verified
            FROM customer_ovo
            WHERE customer_id=?`

    err := c.DB.QueryRow(q, ovoReq.CustomerID).Scan(
        &cOvo.CustomerID,
        &ovoID,
        &cOvo.OvoPhone,
        &cOvo.OvoAuthID,
        &cOvo.FgVerified,
    )

    if ovoID.Valid {
        cOvo.OvoID = ovoID.String
    }

    c.OvoInfo = &cOvo
    if err != nil {
        if err == sql.ErrNoRows {
            return nil
        }
        return err
    }

    return nil

}

func (c *MatahariMall) isPhoneNumberAlreadyLinkage(ovoReq *Request) (bool, error) {
    var s sql.NullString
    q := `SELECT ovo_phone
            FROM customer_ovo
            WHERE ovo_phone = ?
              AND ovo_id != '' LIMIT 1`

    err := c.DB.QueryRow(q, ovoReq.Phone).Scan(&s)
    if err != nil {
        if err == sql.ErrNoRows {
            return false, nil
        }
        return false, err
    }

    if !s.Valid {
        return false, nil
    }

    return true, nil
}

//IsLinkageVerified : Check if customer linkage is already verified
func (c *MatahariMall) IsLinkageVerified(customerID int64) (bool, string, error) {
    var s sql.NullString
    var ovoID string
    q := `SELECT ovo_id
            FROM customer_ovo
           WHERE customer_id = ?
             AND fg_verified = 1`

    err := c.DB.QueryRow(q, customerID).Scan(&s)
    if err != nil {
        if err == sql.ErrNoRows {
            return false, ovoID, nil
        }
        return false, ovoID, err
    }

    if !s.Valid {
        return false, ovoID, nil
    }

    if s.Valid {
        ovoID = s.String
    }

    return true, ovoID, nil
}

//IsLinkageVerifiedByPhone : Check if customer linkage is already verified by phone
func (c *MatahariMall) IsLinkageVerifiedByPhone(phone string) (bool, string, error) {
    var s sql.NullString
    var ovoID string
    q := `SELECT ovo_id
            FROM customer_ovo
           WHERE ovo_phone = ?
             AND fg_verified = 1`

    err := c.DB.QueryRow(q, phone).Scan(&s)
    if err != nil {
        if err == sql.ErrNoRows {
            return false, ovoID, nil
        }
        return false, ovoID, err
    }

    if !s.Valid {
        return false, ovoID, nil
    }

    if s.Valid {
        ovoID = s.String
    }

    return true, ovoID, nil
}

func (c *MatahariMall) getCustomerOvoByPhone(phone string) (int64, int, error) {
    var customerID sql.NullInt64
    var fgVerified sql.NullInt64

    q := `SELECT customer_id, fg_verified FROM customer_ovo WHERE ovo_phone = ? LIMIT 1`

    err := c.DB.QueryRow(q, phone).Scan(&customerID, &fgVerified)
    if err != nil {
        return 0, 0, err
    }
    if !customerID.Valid || !fgVerified.Valid {
        return 0, 0, TErr("ovo_customer_unidentified", c.API.LocaleID)
    }

    fgV := int(fgVerified.Int64)
    cID := customerID.Int64

    return cID, fgV, nil
}

func (c *MatahariMall) validateOvoID(ovoReq *Request) error {
    var err error
    /* Not to validate phone number
       err = c.parsePhoneNumber(ovoReq)
       if err != nil {
           return err
       }*/

    err = c.getOvoInfoFromStorage(ovoReq)
    if err != nil {
        return err
    }
    if c.OvoInfo.FgVerified > 0 {
        if c.OvoInfo.OvoPhone == ovoReq.Phone {
            return TErr("ovo_already_verified", c.API.LocaleID)
        }
        return TErr("ovo_change_verified", c.API.LocaleID)

    }

    return nil
}

//ValidateOvoIDAndAuthenticateToOvo : Validate Customer by phone number and customer id, will push notification to customer device and open “Input Security Code” screen.
func (c *MatahariMall) ValidateOvoIDAndAuthenticateToOvo(ovoReq *Request) error {
    var err error

    c.OvoReq = ovoReq

    err = c.validateOvoID(ovoReq)
    if err != nil {
        return err
    }

    cuid, fgVerified, _ := c.getCustomerOvoByPhone(c.OvoReq.Phone)
    //If existed customer by phone and customer doesn't match then stop process
    if cuid > 0 && cuid != c.OvoReq.CustomerID {
        return TErr("ovo_id_used", c.API.LocaleID)
    }

    //if existed customer by phone and customer is verified then stop process
    if cuid > 0 && fgVerified > 0 {
        return TErr("ovo_already_verified", c.API.LocaleID)
    }

    err = c.doCustomerAuthenticationAtOvo(ovoReq)
    if err != nil {
        return err
    }

    err = c.saveToDatabase()
    if err != nil {
        return err
    }

    return nil
}

func (c *MatahariMall) doCustomerAuthenticationAtOvo(ovoReq *Request) error {

    params := Params{
        "merchant_id": c.API.MerchantID,
        "phone":       ovoReq.Phone,
    }

    data, err := c.API.CustomerAuthentication(params)
    if err != nil {
        return err
    }

    var r Response
    r, err = c.API.getResponse(data)
    if err != nil {
        return err
    }

    if r.Status == http.StatusCreated {
        if r.Code == sendingAuthentication {
            ovoReq.AuthID = r.Data.AuthenticationID
            ovoReq.AuthStatus = r.Code
            c.OvoInfo.OvoAuthID = r.Data.AuthenticationID
            c.OvoInfo.FgVerified = 0
        } else {
            return &CustomError{r.Code, r.Message}
        }
    } else {
        return &CustomError{r.Code, r.Message}
    }

    return nil
}

func (c *MatahariMall) saveToDatabase() error {

    if c.OvoInfo == nil {
        return TErr("ovo_unknown_info", c.API.LocaleID)
    }

    ovoInfo := c.OvoInfo
    var newLinkage bool

    //If there is no record found and customer is a new one
    if ovoInfo.CustomerID == 0 {
        newLinkage = true
        ovoInfo.CustomerID = c.OvoReq.CustomerID
        ovoInfo.OvoPhone = c.OvoReq.Phone
    }

    if newLinkage {
        sqlInsert := `INSERT INTO
                        customer_ovo(
                            customer_id,
                            ovo_phone,
                            ovo_auth_id,
                            fg_verified,
                            created_at,
                            updated_at,
                            source
                        )
                      VALUES (?, ?, ?, 0, NOW(), NOW(), ?)`
        _, errDBInsert := c.DB.Exec(sqlInsert, ovoInfo.CustomerID, ovoInfo.OvoPhone, ovoInfo.OvoAuthID, c.API.AppID)
        if errDBInsert != nil {
            return errDBInsert
        }
    } else {
        var ovoID sql.NullString

        if ovoInfo.OvoID != "" {
            ovoID.String = ovoInfo.OvoID
            ovoID.Valid = true
        }
        toUpdate := map[string]interface{}{
            "ovo_id":      ovoID,
            "ovo_auth_id": ovoInfo.OvoAuthID,
            "fg_verified": ovoInfo.FgVerified,
        }
        fmt.Println(c.OvoReq)
        if ovoInfo.OvoPhone != c.OvoReq.Phone {
            fmt.Println("ga sama woi")
            toUpdate["ovo_phone"] = c.OvoReq.Phone
        }

        return c.updateCustomerOVO(ovoInfo.CustomerID, toUpdate)
    }
    return nil
}

func (c *MatahariMall) updateCustomerOVO(customerID interface{}, toUpdate map[string]interface{}) error {
    sqlUpdate := `UPDATE customer_ovo SET updated_at = NOW() `
    var setStr []string
    var vals []interface{}
    for k, v := range toUpdate {
        setStr = append(setStr, k+" = ?")
        vals = append(vals, v)
    }
    if len(setStr) > 0 {
        sqlUpdate = sqlUpdate + "," + strings.Join(setStr, ",") + " "
    }

    vals = append(vals, customerID)
    sqlUpdate += " WHERE customer_id = ?"
    res, err := c.DB.Exec(sqlUpdate, vals...)

    if err != nil {
        if strings.Contains(err.Error(), "1062") {
            return TErr("ovo_id_used", c.API.LocaleID)
        }
        return err
    }

    if rowAffected, _ := res.RowsAffected(); rowAffected == 0 {
        return TErr("ovo_id_used", c.API.LocaleID)
    }

    return nil
}

//CheckOvoStatus : Checking ovo status by customer id
func (c *MatahariMall) CheckOvoStatus(customerID int64) (*CustomerOvo, error) {
    ovoReq := &Request{
        CustomerID: customerID,
    }
    c.OvoReq = ovoReq
    err := c.getOvoInfoFromStorage(ovoReq)
    if err != nil {
        return nil, TErr("ovo_unknown_info", c.API.LocaleID)
    }
    c.OvoReq.Phone = c.OvoInfo.OvoPhone
    if c.OvoInfo.CustomerID == 0 {
        return nil, TErr("ovo_not_authenticated", c.API.LocaleID)
    } else if c.OvoInfo.FgVerified <= 0 {
        err = c.getCustomerAuthenticationStatusAtOvo()
        if err != nil {
            return nil, err
        }

        err = c.saveToDatabase()
        if err != nil {
            return nil, err
        }
    }

    return c.OvoInfo, nil

}

func (c *MatahariMall) getCustomerAuthenticationStatusAtOvo() error {
    data, err := c.API.CheckCustomerAuthenticationStatus(c.OvoInfo.OvoAuthID)
    if err != nil {
        return err
    }
    var r Response
    r, err = c.API.getResponse(data)
    if err != nil {
        return err
    }

    if r.Status == http.StatusOK {
        if r.Code == Authenticated {
            c.OvoInfo.OvoID = r.Data.LoyaltyID
            c.OvoInfo.FgVerified = 1
            return nil
        }
    }

    if r.Code == Unauthenticated || r.Code == AuthIDNotFound || r.Code == CustomerNotFound {
        return TErr("ovo_retry_verification", c.API.LocaleID)
    }

    return err

}

//CalculateHyperOvoPoint : Calculate Ovo Point for Hyper only
func (c *MatahariMall) CalculateHyperOvoPoint(ovoID string, param Params) error {

    data, err := c.API.CalculatePoints(ovoID, param)
    if err != nil {
        return err
    }
    var r Response
    r, err = c.API.getResponse(data)
    if err != nil {
        fmt.Printf("Err CalculateHyperOvoPoint %v \n", err)
        return err
    }

    if r.Status == http.StatusOK {
        log.Printf("%#v", r.Data)
        return nil
    }
    fmt.Printf("Response CalculateHyperOvoPoint %v \n", r)
    if r.Code != Success {
        if r.Code == DuplicateMerchantInvoice {
            return nil
        }
        return errors.New(r.Message)
    }

    return err

}

//AddOvoPointHistory : Add Ovo Point History
func (c *MatahariMall) AddOvoPointHistory(customerID, orderID int64, soNumber, pointType string, payload Params, flags ...map[string]interface{}) error {
    jsonPayload, err := json.Marshal(payload)
    if err != nil {
        return err
    }

    var fgFailed interface{}
    fgFailed = 0

    if len(flags) > 0 {
        flag := flags[0]
        if val, ok := flag["fg_failed"]; ok {
            fgFailed = val
        }
    }

    sqlInsert := `INSERT INTO
                        ovo_points(
                            customer_id,
                            order_id,
                            so_number,
                            type,
                            payload,
                            fg_failed
                        )
                      VALUES (?, ?, ?, ?, ?, ?)`
    _, errDBInsert := c.DB.Exec(sqlInsert, customerID, orderID, soNumber, pointType, jsonPayload, fgFailed)
    if errDBInsert != nil {
        return errDBInsert
    }

    return nil
}

//AddBgLinkage : Background linkage
func (c *MatahariMall) AddBgLinkage(ovoReq *Request, verifiedTime time.Duration) {
    go func() {
        time.Sleep(time.Second * 5)
        if err := c.ValidateOvoIDAndAuthenticateToOvo(ovoReq); err != nil {
            return
        }
        //Wait user action on the ovo app before verifying
        time.Sleep(time.Second * verifiedTime)
        //Verified the auth id
        _, err := c.CheckOvoStatus(ovoReq.CustomerID)
        if err != nil {
            return
        }
    }()
}
