package ovo

import (
    "database/sql"
    "errors"
    "fmt"
    "regexp"
    "strings"
)

func (c *MatahariMall) parsePhoneNumber(ovoReq *OvoRequest) error {
    re := regexp.MustCompile("(0|\\+)([0-9]{5,16})")
    x := re.MatchString(ovoReq.Phone)

    if x {
        ovoReq.Phone = strings.Replace(ovoReq.Phone, "+", "", 2)
    } else {
        return errors.New("Invalid phone number")
    }

    if ovoReq.Phone[0:2] == "62" {
        ovoReq.Phone = "0" + ovoReq.Phone[2:]
    }

    return nil
}

func (c *MatahariMall) getOvoInfoFromStorage(ovoReq *OvoRequest) error {
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

func (c *MatahariMall) isPhoneNumberAlreadyLinkage(ovoReq *OvoRequest) (bool, error) {
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
        return 0, 0, errors.New("Customer not found")
    }

    fgV := int(fgVerified.Int64)
    cID := customerID.Int64

    return cID, fgV, nil
}

func (c *MatahariMall) validateOvoID(ovoReq *OvoRequest) error {
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
            return errors.New("Already verified")
        }

        return errors.New("Cannot change OVO id that has been verified")

    } else if c.OvoInfo.OvoPhone != ovoReq.Phone {
        c.OvoInfo.OvoPhone = ovoReq.Phone
    }

    return nil
}

//ValidateOvoIDAndAuthenticateToOvo : Validate Customer by phone number and customer id, will push notification to customer device and open “Input Security Code” screen.
func (c *MatahariMall) ValidateOvoIDAndAuthenticateToOvo(ovoReq *OvoRequest) error {
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
        return errors.New("Phone Number already used by other customer")
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

func (c *MatahariMall) doCustomerAuthenticationAtOvo(ovoReq *OvoRequest) error {

    params := Params{
        "merchant_id": c.Api.MerchantID,
        "phone":       ovoReq.Phone,
    }

    data, err := c.Api.CustomerAuthentication(params)
    if err != nil {
        return err
    }

    var r OvoResponse
    r, err = c.Api.getResponse(data)
    if err != nil {
        return err
    }

    if r.Status == httpCreated {
        if r.Code == sendingAuthentication {
            ovoReq.AuthID = r.Data.AuthenticationID
            ovoReq.AuthStatus = r.Code
            c.OvoInfo.OvoAuthID = r.Data.AuthenticationID
            c.OvoInfo.CustomerID = ovoReq.CustomerID
            c.OvoInfo.FgVerified = 0
            c.OvoInfo.OvoPhone = ovoReq.Phone
        } else {
            return &OvoError{r.Code, r.Message}
        }
    } else {
        return &OvoError{r.Code, r.Message}
    }

    return nil
}

func (c *MatahariMall) saveToDatabase() error {

    if c.OvoInfo == nil {
        return errors.New("Ovo Information not found")
    }

    ovoInfo := c.OvoInfo

    //Check if exist by phone
    _, _, err := c.getCustomerOvoByPhone(ovoInfo.OvoPhone)

    if err != nil {
        if err == sql.ErrNoRows {
            fmt.Println("CREATE NEW OVO CUSTOMER")
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
            _, errDBInsert := c.DB.Exec(sqlInsert, ovoInfo.CustomerID, ovoInfo.OvoPhone, ovoInfo.OvoAuthID, c.Api.AppID)
            if errDBInsert != nil {
                return errDBInsert
            }
        }
    } else {
        fmt.Println("UPDATE OVO CUSTOMER")
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
        _, errDBUpdate := c.DB.Exec(sqlUpdate, ovoID, ovoInfo.OvoPhone, ovoInfo.OvoAuthID, ovoInfo.FgVerified, ovoInfo.CustomerID)
        if errDBUpdate != nil {
            return errDBUpdate
        }
    }
    return nil
}

//CheckOvoStatus : Checking ovo status by customer id
func (c *MatahariMall) CheckOvoStatus(customerID int64) error {
    ovoReq := &OvoRequest{
        CustomerID: customerID,
    }

    err := c.getOvoInfoFromStorage(ovoReq)
    if err != nil {
        return errors.New("Cannot load Ovo Info")
    }

    if c.OvoInfo.CustomerID > 0 && c.OvoInfo.FgVerified <= 0 {
        err = c.getCustomerAuthenticationStatusAtOvo()
        if err != nil {
            return err
        }

        err = c.saveToDatabase()
        if err != nil {
            return err
        }
    }

    return nil

}

func (c *MatahariMall) getCustomerAuthenticationStatusAtOvo() error {
    data, err := c.Api.CheckCustomerAuthenticationStatus(c.OvoInfo.OvoAuthID)
    if err != nil {
        return err
    }

    var r OvoResponse
    r, err = c.Api.getResponse(data)
    if err != nil {
        return err
    }

    if r.Status == httpOk {
        if r.Code == Authenticated {
            c.OvoInfo.OvoID = r.Data.LoyaltyID
            c.OvoInfo.FgVerified = 1
            return nil
        }
    }

    return &OvoError{r.Code, r.Message}

}
