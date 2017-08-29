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

type MatahariMall struct {
    DB      *sql.DB
    Api     *Client
    OvoInfo *CustomerOvo
}

type Params map[string]string

//OvoRequest : Request OVO params
type OvoRequest struct {
    CustomerID    int64
    OvoID         string
    Phone         string
    AuthID        string
    AuthStatus    int
    TransactionID string
    TerminalId    string
    FgVerified    int
    Source        string
}

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

type OvoResponse struct {
    Status  int
    Data    OvoResponseData
    Message string
    Code    int
}

type OvoResponseData struct {
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

type OvoError struct {
    code int
    msg  string
}
