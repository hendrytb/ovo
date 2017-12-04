package ovo

import (
    "database/sql"
    "fmt"
    "net/http"
    "testing"

    sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestIsLinkageVerified(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    mock.ExpectQuery(`SELECT`).WillReturnError(sql.ErrNoRows)

    client := new(Client)
    mmsdk := client.GetMMsdk(db)
    _, _, err = mmsdk.IsLinkageVerified(123456)
    if err != nil {
        t.Errorf("Err No Rows should be nil")
    }
}

func TestGetCustomerOvoByPhone(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    rows := sqlmock.NewRows([]string{"customer_id", "fg_verified"}).AddRow(1, 0)
    mock.ExpectQuery(`SELECT customer_id, fg_verified`).WillReturnRows(rows)

    client := new(Client)

    mmsdk := client.GetMMsdk(db)

    _, _, err = mmsdk.getCustomerOvoByPhone("0818181818")
    if err != nil {
        t.Errorf("All is valid, should not error")
    }
}

func TestSaveToDatabaseUpdate(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    mock.ExpectExec(`UPDATE customer_ovo`).WillReturnResult(sqlmock.NewResult(0, 0))

    client := new(Client)
    client.LocaleID = "en"
    mmsdk := client.GetMMsdk(db)
    mmsdk.OvoInfo = &CustomerOvo{
        CustomerID: 1234,
        OvoID:      "",
        OvoPhone:   "08282828",
        OvoAuthID:  "234",
        FgVerified: 0,
    }
    mmsdk.OvoReq = &Request{
        CustomerID: 1234,
        Phone:      "08282828",
    }

    err = mmsdk.saveToDatabase()
    if err != nil && err.Error() != TErr("ovo_id_used", client.LocaleID).Error() {
        t.Errorf("Should error ovo id used, if no row affected")
    }
}

func TestSaveToDatabaseInsert(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    mock.ExpectExec(`INSERT INTO customer_ovo`).WillReturnResult(sqlmock.NewResult(1, 1))

    client := new(Client)
    client.LocaleID = "en"
    mmsdk := client.GetMMsdk(db)
    mmsdk.OvoInfo = &CustomerOvo{
        CustomerID: 0,
        OvoAuthID:  "234",
        FgVerified: 0,
    }
    mmsdk.OvoReq = &Request{
        Phone: "08282828",
    }

    err = mmsdk.saveToDatabase()
    if err != nil {
        t.Errorf("Should not be error, when all condition met")
    }
}

func TestParsePhoneNumberValid(t *testing.T) {
    client := new(Client)
    client.LocaleID = "en"
    mmsdk := client.GetMMsdk(nil)

    ovoReq := &Request{
        Phone: "0818223456",
    }
    err := mmsdk.parsePhoneNumber(ovoReq)
    if err != nil {
        t.Errorf("Should not error for valid phone number")
    }
}

func TestParsePhoneNumberInvalid(t *testing.T) {
    client := new(Client)
    client.LocaleID = "en"
    mmsdk := client.GetMMsdk(nil)

    ovoReq := &Request{
        Phone: "818223456",
    }
    err := mmsdk.parsePhoneNumber(ovoReq)
    if err == nil {
        t.Errorf("Should error for invalid phone number")
    }
}

func TestGetOvoInfoFromStorage(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    mock.ExpectQuery(`SELECT customer_id`).WillReturnError(sql.ErrNoRows)

    client := new(Client)
    client.LocaleID = "en"
    mmsdk := client.GetMMsdk(db)
    ovoReq := &Request{
        CustomerID: 12345,
    }

    err = mmsdk.getOvoInfoFromStorage(ovoReq)
    if err != nil {
        t.Errorf("Should return nil when no rows found")
    }

}

func TestIsLinkageVerifiedNotValid(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    rows := sqlmock.NewRows([]string{"customer_id"}).AddRow(nil)
    mock.ExpectQuery(`SELECT ovo_id FROM customer_ovo`).WillReturnRows(rows)

    client := new(Client)
    client.LocaleID = "en"
    mmsdk := client.GetMMsdk(db)

    ok, _, erro := mmsdk.IsLinkageVerified(12345)
    if ok {
        t.Errorf("This should not ok if customer_id not valid")
    }
    if erro != nil {
        fmt.Println(erro)
        t.Errorf("This should not error")
    }
}

func TestValidateOvoID(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    rows := sqlmock.NewRows([]string{"customer_id", "ovo_id", "ovo_phone", "ovo_auth_id", "fg_verified"}).AddRow(12345, "6789", "08080808", "123", 1)
    mock.ExpectQuery(`SELECT customer_id`).WillReturnRows(rows)

    ovoReq := &Request{
        CustomerID: 12345,
        Phone:      "08080808",
    }

    client := new(Client)
    client.LocaleID = "en"
    mmsdk := client.GetMMsdk(db)

    erro := mmsdk.validateOvoID(ovoReq)
    if erro == nil {
        t.Errorf("This should return error when already verified")
    }

    if erro != nil && erro.Error() != TErr("ovo_already_verified", client.LocaleID).Error() {
        t.Errorf("This should return Err: " + TErr("ovo_already_verified", client.LocaleID).Error())
    }
}

func TestValidateOvoIDChange(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    rows := sqlmock.NewRows([]string{"customer_id", "ovo_id", "ovo_phone", "ovo_auth_id", "fg_verified"}).AddRow(12345, "6789", "08080808", "123", 1)
    mock.ExpectQuery(`SELECT customer_id`).WillReturnRows(rows)

    ovoReq := &Request{
        CustomerID: 12345,
        Phone:      "0909090909",
    }

    client := new(Client)
    client.LocaleID = "en"
    mmsdk := client.GetMMsdk(db)

    erro := mmsdk.validateOvoID(ovoReq)
    if erro == nil {
        t.Errorf("This should return error when already verified")
    }

    if erro != nil && erro.Error() != TErr("ovo_change_verified", client.LocaleID).Error() {
        t.Errorf("This should return Err: " + TErr("ovo_change_verified", client.LocaleID).Error())
    }
}

func TestValidateOvoIDAndAuthenticateToOvoLinkage(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    rows := sqlmock.NewRows([]string{"customer_id", "ovo_id", "ovo_phone", "ovo_auth_id", "fg_verified"}).AddRow(12345, "6789", "08080808", "123", 1)
    mock.ExpectQuery(`SELECT customer_id`).WillReturnRows(rows)

    rowsOvophone := sqlmock.NewRows([]string{"ovo_phone"}).AddRow("08080808")
    mock.ExpectQuery(`SELECT ovo_phone FROM customer_ovo`).WillReturnRows(rowsOvophone)

    ovoReq := &Request{
        CustomerID: 12345,
        Phone:      "08080808",
    }

    client := new(Client)
    client.LocaleID = "en"
    mmsdk := client.GetMMsdk(db)

    err = mmsdk.ValidateOvoIDAndAuthenticateToOvo(ovoReq)

    if err != nil && err.Error() != TErr("ovo_already_verified", client.LocaleID).Error() {
        t.Errorf("This should return Err: " + TErr("ovo_already_verified", client.LocaleID).Error())
    }
}

func TestValidateOvoIDAndAuthenticateToOvoSuccessUpdate(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    rows := sqlmock.NewRows([]string{"customer_id", "ovo_id", "ovo_phone", "ovo_auth_id", "fg_verified"}).AddRow(12345, "6789", "08080808", "123", 0)
    mock.ExpectQuery(`SELECT customer_id, ovo_id, ovo_phone, ovo_auth_id, fg_verified`).WillReturnRows(rows)

    mock.ExpectQuery(`SELECT customer_id, fg_verified FROM customer_ovo`).WillReturnRows(sqlmock.NewRows([]string{"customer_id", "fg_verified"}).AddRow(12345, 0))
    mock.ExpectExec(`UPDATE customer_ovo`).WillReturnResult(sqlmock.NewResult(0, 1))

    ovoReq := &Request{
        CustomerID: 12345,
        Phone:      "08080808",
    }

    client := new(Client)
    client.LocaleID = "en"
    client.httpHandler = func(w http.ResponseWriter, r *http.Request) {
        data := `{
                    "status": 201,
                    "data": {
                    "authentication_id": "666"
                    },
                    "message": "Success",
                    "code": 1
                }`
        w.Write([]byte(data))
    }
    mmsdk := client.GetMMsdk(db)

    err = mmsdk.ValidateOvoIDAndAuthenticateToOvo(ovoReq)
    if err != nil {
        t.Errorf("This should not error expect success")
    }

}

func TestValidateOvoIDAndAuthenticateToOvoSuccessInsert(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    mock.ExpectQuery(`SELECT customer_id, ovo_id, ovo_phone, ovo_auth_id, fg_verified`).WillReturnError(sql.ErrNoRows)

    mock.ExpectQuery(`SELECT customer_id, fg_verified FROM customer_ovo`).WillReturnError(sql.ErrNoRows)
    mock.ExpectExec(`INSERT`).WillReturnResult(sqlmock.NewResult(1, 1))

    ovoReq := &Request{
        CustomerID: 12345,
        Phone:      "08080808",
    }

    client := new(Client)
    client.LocaleID = "en"
    client.httpHandler = func(w http.ResponseWriter, r *http.Request) {
        data := `{
                    "status": 201,
                    "data": {
                    "authentication_id": "666"
                    },
                    "message": "Success",
                    "code": 1
                }`
        w.Write([]byte(data))
    }
    mmsdk := client.GetMMsdk(db)

    err = mmsdk.ValidateOvoIDAndAuthenticateToOvo(ovoReq)
    if err != nil {
        t.Errorf("This should not error expect success")
    }

}

func TestCheckOvoStatusUnauthenticated(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    mock.ExpectQuery(`SELECT customer_id`).WillReturnError(sql.ErrNoRows)

    client := new(Client)
    client.LocaleID = "en"
    mmsdk := client.GetMMsdk(db)

    _, err = mmsdk.CheckOvoStatus(12345)
    if err == nil {
        t.Errorf("This should error upon not authenticated")
    }
    if err != nil && err.Error() != TErr("ovo_not_authenticated", client.LocaleID).Error() {
        t.Errorf("This should return Err: " + TErr("ovo_not_authenticated", client.LocaleID).Error())
    }
}

func TestCheckOvoStatusSuccess(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    rows := sqlmock.NewRows([]string{"customer_id", "ovo_id", "ovo_phone", "ovo_auth_id", "fg_verified"}).AddRow(12345, "6789", "08080808", "123", 0)
    mock.ExpectQuery(`SELECT customer_id, ovo_id, ovo_phone, ovo_auth_id, fg_verified`).WillReturnRows(rows)

    mock.ExpectExec(`UPDATE customer_ovo`).WillReturnResult(sqlmock.NewResult(0, 1))

    client := new(Client)
    client.LocaleID = "en"
    client.httpHandler = func(w http.ResponseWriter, r *http.Request) {
        data := `{
                    "status": 200,
                    "data": {
                    "loyalty_id": "8000428048133600"
                    },
                    "message": "Authenticated",
                    "code": 1
                }`
        w.Write([]byte(data))
    }
    mmsdk := client.GetMMsdk(db)
    _, errs := mmsdk.CheckOvoStatus(12345)
    if errs != nil {
        t.Errorf("Should not return error upon success")
    }

}
