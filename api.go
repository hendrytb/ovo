package ovo

//GetCustomerProfile : Get Customer Profile
func (client *Client) GetCustomerProfile(customerID string) ([]byte, error) {
    url, err := client.getURL("customer_profile", customerID)

    if err != nil {
        return nil, err
    }

    data, errReq := client.execRequest("GET", url, nil)
    if errReq != nil {
        return nil, errReq
    }

    return data, nil
}

//GetCustomerProfileQR : Get Customer Profile (QR)
func (client *Client) GetCustomerProfileQR(merchantID, storeID, terminalID string) ([]byte, error) {
    url, err := client.getURL("customer_profile_qr", merchantID, storeID, terminalID)

    if err != nil {
        return nil, err
    }

    data, errReq := client.execRequest("GET", url, nil)
    if errReq != nil {
        return nil, errReq
    }

    return data, nil
}

//CalculatePoints : Calculate Points
func (client *Client) CalculatePoints(customerID string, params Params) ([]byte, error) {

    url, err := client.getURL("calculate_points", customerID)

    if err != nil {
        return nil, err
    }

    if params == nil {
        params = Params{}
    }

    buf := client.createParams(params)

    data, errReq := client.execRequest("PUT", url, buf)

    if errReq != nil {
        return nil, errReq
    }

    return data, nil
}

//CreateTransaction : Create Push to Pay / Scan to Pay Transaction
func (client *Client) CreateTransaction(customerID string, params Params) ([]byte, error) {

    url, err := client.getURL("pushtopay_transaction", customerID)

    if err != nil {
        return nil, err
    }

    if params == nil {
        params = Params{}
    }

    buf := client.createParams(params)

    data, errReq := client.execRequest("POST", url, buf)

    if errReq != nil {
        return nil, errReq
    }

    return data, nil
}

//CheckTransactionStatus : Check Push to Pay / Scan To Pay Transaction Status
func (client *Client) CheckTransactionStatus(customerID string, transactionID interface{}) ([]byte, error) {
    url, err := client.getURL("pushtopay_transaction_status", customerID)

    if err != nil {
        return nil, err
    }

    data, errReq := client.execRequest("GET", url, nil)
    if errReq != nil {
        return nil, errReq
    }

    return data, nil
}

//VoidTransaction : Void Push To Pay / Scan To Pay Transaction
func (client *Client) VoidTransaction(customerID, transactionID string, params Params) ([]byte, error) {

    url, err := client.getURL("pushtopay_void_transaction", customerID, transactionID)

    if err != nil {
        return nil, err
    }

    if params == nil {
        params = Params{}
    }

    buf := client.createParams(params)

    data, errReq := client.execRequest("PUT", url, buf)

    if errReq != nil {
        return nil, errReq
    }

    return data, nil
}

//CreateCustomerLinkage : Customer Creation / Linkage
func (client *Client) CreateCustomerLinkage(customerID string, params Params) ([]byte, error) {

    url, err := client.getURL("customer_linkage", customerID)

    if err != nil {
        return nil, err
    }

    if params == nil {
        params = Params{}
    }

    buf := client.createParams(params)

    data, errReq := client.execRequest("POST", url, buf)

    if errReq != nil {
        return nil, errReq
    }

    return data, nil
}

//CustomerAuthentication : Customer authentication, this API will push notification to customer device and open “Input Security Code” screen.
func (client *Client) CustomerAuthentication(params Params) ([]byte, error) {

    url, err := client.getURL("customer_authentication")

    if err != nil {
        return nil, err
    }

    if params == nil {
        params = Params{}
    }

    buf := client.createParams(params)

    data, errReq := client.execRequest("POST", url, buf)

    if errReq != nil {
        return nil, errReq
    }

    return data, nil

}

//CheckCustomerAuthenticationStatus : Check Customer Authentication Status
func (client *Client) CheckCustomerAuthenticationStatus(authenticationID string) ([]byte, error) {
    url, err := client.getURL("customer_authentication_status", authenticationID)
    if err != nil {
        return nil, err
    }

    data, errReq := client.execRequest("GET", url, nil)
    if errReq != nil {
        return nil, errReq
    }

    return data, nil
}
