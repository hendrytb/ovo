package ovo

import "fmt"

func (e *OvoError) Error() string {
    return fmt.Sprintf("%s", e.msg)
}

//GetErrCode : Get OVO error code
func GetErrCode(e error) int {
    if ae, ok := e.(*OvoError); ok {
        return ae.code
    }
    return 0
}

//GetErrMsg : Get OVO error message
func GetErrMsg(e error) string {
    return e.Error()
}
