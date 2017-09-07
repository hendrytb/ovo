package ovo

import (
    "errors"
    "fmt"
)

func (e *CustomError) Error() string {
    return fmt.Sprintf("%s", e.msg)
}

//GetErrCode : Get OVO error code
func GetErrCode(e error) int {
    if ae, ok := e.(*CustomError); ok {
        return ae.code
    }
    return 0
}

//GetErrMsg : Get OVO error message
func GetErrMsg(e error) string {
    return e.Error()
}

//TErr : Translate OVO related error message
func TErr(keyword, locale string) error {

    v, ok := ErrMessage[keyword][locale]
    if ok {
        return errors.New(v)
    }

    return errors.New("Unknown OVO Service error")
}

//TCustomErr : Translate OVO related custom error message
func TCustomErr(keyword string, errCode int, locale string) error {

    v, ok := ErrMessage[keyword][locale]
    if ok {
        return &CustomError{errCode, v}
    }

    return errors.New("Unknown OVO Service error")
}
