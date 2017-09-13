package ovo

import (
    "database/sql"
    "net/http"
    "regexp"
    "strings"
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
func (c *MatahariMall) IsLinkageVerified(customerID int64) (bool, error) {
    var s sql.NullString
    q := `SELECT customer_id
            FROM customer_ovo
           WHERE customer_id = ?
             AND fg_verified = 1`

    err := c.DB.QueryRow(q, customerID).Scan(&s)
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
    err = c.parsePhoneNumber(ovoReq)
    if err != nil {
        return err
    }

    err = c.getOvoInfoFromStorage(ovoReq)
    if err != nil {
        return err
    }
    if c.OvoInfo.FgVerified > 0 {
        if c.OvoInfo.OvoPhone == ovoReq.Phone {
            return TErr("ovo_already_verified", c.API.LocaleID)
        }
        return TErr("ovo_change_verified", c.API.LocaleID)

    } else if c.OvoInfo.OvoPhone != ovoReq.Phone {
        c.OvoInfo.OvoPhone = ovoReq.Phone
    }

    return nil
}

//ValidateOvoIDAndAuthenticateToOvo : Validate Customer by phone number and customer id, will push notification to customer device and open “Input Security Code” screen.
func (c *MatahariMall) ValidateOvoIDAndAuthenticateToOvo(ovoReq *Request) error {
    var err error

    err = c.validateOvoID(ovoReq)
    if err != nil {
        return err
    }

    var isLinked bool
    isLinked, err = c.isPhoneNumberAlreadyLinkage(ovoReq)
    if err != nil {
        return err
    }

    if isLinked {
        return TErr("ovo_id_used", c.API.LocaleID)
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
            c.OvoInfo.CustomerID = ovoReq.CustomerID
            c.OvoInfo.FgVerified = 0
            c.OvoInfo.OvoPhone = ovoReq.Phone
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

    //Check if exist by phone
    _, _, err := c.getCustomerOvoByPhone(ovoInfo.OvoPhone)

    if err != nil {
        if err == sql.ErrNoRows {
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
            return err
        }
    } else {
        var ovoID sql.NullString
        if ovoInfo.OvoID != "" {
            ovoID.String = ovoInfo.OvoID
            ovoID.Valid = true
        }
        sqlUpdate := `UPDATE customer_ovo SET
                                  ovo_id = ?,
                                  ovo_phone = ?,
                                  ovo_auth_id = ?,
                                  fg_verified = ?,
                                  updated_at = NOW()
                                WHERE customer_id = ?`
        resUpdate, errDBUpdate := c.DB.Exec(sqlUpdate, ovoID, ovoInfo.OvoPhone, ovoInfo.OvoAuthID, ovoInfo.FgVerified, ovoInfo.CustomerID)
        if errDBUpdate != nil {
            if strings.Contains(errDBUpdate.Error(), "1062") {
                return TErr("ovo_id_used", c.API.LocaleID)
            }
            return errDBUpdate
        }
        if rowAffected, _ := resUpdate.RowsAffected(); rowAffected == 0 {
            return TErr("ovo_id_used", c.API.LocaleID)
        }
    }
    return nil
}

//CheckOvoStatus : Checking ovo status by customer id
func (c *MatahariMall) CheckOvoStatus(customerID int64) (*CustomerOvo, error) {
    ovoReq := &Request{
        CustomerID: customerID,
    }
    err := c.getOvoInfoFromStorage(ovoReq)
    if err != nil {
        return nil, TErr("ovo_unknown_info", c.API.LocaleID)
    }
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
