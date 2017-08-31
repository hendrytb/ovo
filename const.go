package ovo

const (
    //NoErrCode : No Error code is set
    NoErrCode = 0

    sendingAuthentication = 1

    //MerchantIDMustNotEmpty : OVO merchant_id must not empty
    MerchantIDMustNotEmpty = 2

    //PhoneMustNotEmpty : OVO phone must not empty
    PhoneMustNotEmpty = 3

    //CustomerNotFound : OVO customer not found
    CustomerNotFound = 4

    //Authenticated : OVO check auth status is authenticated
    Authenticated = 1

    //Unauthenticated : OVO check auth status is not authenticated
    Unauthenticated = 2

    //AuthIDNotFound : OVO check auth status authentication_id is not found
    AuthIDNotFound = 3
)

const (
    //PhoneValidRegex : Regex to check if phone is valid
    PhoneValidRegex = "(0|\\+)([0-9]{5,16})"
)

var (
    domainMap = map[string]string{
        "customer_profile":               "/customers/:customer_id",                                                   //GET
        "calculate_points":               "/customers/:customer_id/points",                                            //PUT
        "pushtopay_transaction":          "/customers/:customer_id/transactions",                                      //POST
        "pushtopay_transaction_status":   "/customers/:customer_id/transactions/:transaction_id",                      //GET
        "pushtopay_void_transaction":     "/customers/:customer_id/transactions/:transaction_id",                      //PUT
        "customer_profile_qr":            "/merchants/:merchant_id/stores/:store_id/terminals/:terminal_id/customers", //GET
        "customer_linkage":               "/customers/:customer_id",                                                   //POST
        "customer_authentication":        "/authentications",                                                          // POST
        "customer_authentication_status": "/authentications/:authentication_id",                                       //GET
    }
)
