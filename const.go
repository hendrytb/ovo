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

var (
    //ErrMessage : Ovo related error message
    ErrMessage = map[string]map[string]string{
        "ovo_unavailable_service": {
            "id": "Maaf, layanan OVO sedang tidak aktif, mohon mencoba beberapa saat lagi",
            "en": "Sorry, OVO Service currently not available, please try again in a moment",
        },
        "ovo_unregistered_customer": {
            "id": "Maaf, anda tidak terdaftar sebagai pelanggan OVO",
            "en": "Sorry, You're not registered as OVO customer",
        },
        "ovo_phone_empty": {
            "id": "Nomor telepon tidak boleh kosong",
            "en": "Phone number cannot be empty",
        },
        "ovo_id_invalid": {
            "id": "Silahkan masukkan OVO ID/Nomor telepon yang benar.",
            "en": "Invalid OVO ID (Phone)",
        },
        "ovo_retry_verification": {
            "id": "Silahkan ulangi Verifikasi OVO ID anda",
            "en": "Please retry your OVO ID verification",
        },
        "ovo_invalid_response": {
            "id": "Terjadi kesalahan pada layanan OVO",
            "en": "Cannot read OVO response",
        },
        "ovo_unidentified_request": {
            "id": "Permintaan layanan OVO yang tidak teridentifikasi",
            "en": "Unidentified OVO service request",
        },
        "ovo_customer_unidentified": {
            "id": "Terjadi kesalahan, pelanggan tidak ditemukan",
            "en": "Error occured, customer not found",
        },
        "ovo_already_verified": {
            "id": "Akun OVO sudah pernah terverifikasi",
            "en": "OVO account already verified",
        },
        "ovo_change_verified": {
            "id": "OVO ID yang telah terverifikasi tidak bisa diganti",
            "en": "Cannot change OVO id that has been verified",
        },
        "ovo_id_used": {
            "id": "OVO ID telah digunakan oleh pelanggan lain",
            "en": "OVO ID already used by other customer",
        },
        "ovo_unknown_info": {
            "id": "Informasi OVO anda tidak ditemukan",
            "en": "Cannot load OVO Information",
        },
        "ovo_not_authenticated": {
            "id": "Maaf, Anda belum terotentifikasi",
            "en": "Sorry, You are not yet authenticated",
        },
    }
)
