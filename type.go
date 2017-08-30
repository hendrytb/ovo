package ovo

import (
    "database/sql"
    "time"
)

//Client : Contstructor for OVO Client/App
type Client struct {
    BaseURL    string
    APIKey     string
    AppID      string
    MerchantID string
    Random     string
    Hmac       string
}

//MatahariMall : Type for MatahariMall sdk
type MatahariMall struct {
    DB      *sql.DB
    API     *Client
    OvoInfo *CustomerOvo
}

//Params : Type for api parameters
type Params map[string]string

//Request : Type request OVO
type Request struct {
    CustomerID    int64
    OvoID         string
    Phone         string
    AuthID        string
    AuthStatus    int
    TransactionID string
    TerminalID    string
    FgVerified    int
    Source        string
}

//CustomerOvo : Type for struct name and field the same as db table
type CustomerOvo struct {
    CustomerID int64
    OvoID      string
    OvoPhone   string
    OvoAuthID  string
    FgVerified int
    CreatedAt  *time.Time
    UpdatedAt  *time.Time
    Source     string
}

//Response : Type for Ovo api response
type Response struct {
    Status  int
    Data    ResponseData
    Message string
    Code    int
}

//ResponseData : Part of Ovo api response data
type ResponseData struct {
    AuthenticationID string `json:"authentication_id"`
    LoyaltyID        string `json:"loyalty_id"`
    Fullname         string `json:"fullname"`
    Birthdate        string `json:"birthdate"`
    Phone            string `json:"phone"`
    Email            string `json:"email"`
    Level            string `json:"level"`
    CustomerFullname string `json:"customer_fullname"`
    CustomerPhone    string `json:"customer_phone"`
    OrderID          string `json:"order_id"`
    VoucherCode      string `json:"voucher_code"`
    ApprovalCode     string `json:"approval_code"`
    MerchantInvoice  string `json:"merchant_invoice"`
}

//CustomError : Ovo Error handler
type CustomError struct {
    code int
    msg  string
}
