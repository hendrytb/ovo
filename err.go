package ovo

import "fmt"

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
