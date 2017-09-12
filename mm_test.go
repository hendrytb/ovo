package ovo

import (
    "database/sql"
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
    _, err = mmsdk.IsLinkageVerified(123456)
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

func TestSaveToDatabase(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil {
        t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
    }
    defer db.Close()

    rows := sqlmock.NewRows([]string{"customer_id", "fg_verified"}).AddRow(1, 0)
    mock.ExpectQuery(`SELECT customer_id, fg_verified`).WillReturnRows(rows)
    mock.ExpectExec(`UPDATE customer_ovo`).WillReturnResult(sqlmock.NewResult(0, 0))

    client := new(Client)
    client.LocaleID = "en"
    mmsdk := client.GetMMsdk(db)
    mmsdk.OvoInfo = &CustomerOvo{
        CustomerID: 1234,
        OvoPhone:   "08282828",
        OvoAuthID:  "234",
        FgVerified: 0,
    }

    err = mmsdk.saveToDatabase()
    if err == nil {
        t.Errorf("Should return error if no row affected")
    }
    if err != nil && err.Error() != TErr("ovo_id_used", client.LocaleID).Error() {
        t.Errorf("Should error ovo id used, if no row affected")
    }
}
