package bean

type Err struct {
	Code int
	Msg  string
}

const (
	Err_Data_ILLEGAL           = 1000
	Err_Data_Is_Null           = 1001
	Err_Data_Param_Illegal     = 1002
	Err_Account_Not_Exit       = 2000
	Err_Account_Has_Exited     = 2001
	Err_Token_Appid_Is_Illegal = 3001
	Err_Token_Is_Illegal       = 3002
	Err_Token_Is_Expired       = 3003
)

var errMap = map[int]string{
	Err_Data_ILLEGAL:           "Data Illegal",
	Err_Account_Not_Exit:       "Account Not Exit",
	Err_Account_Has_Exited:     "Account Has Exited",
	Err_Data_Is_Null:           "Data Is Null",
	Err_Data_Param_Illegal:     "Param Is Illegal",
	Err_Token_Appid_Is_Illegal: "Appid Is Illegal",
	Err_Token_Is_Illegal:       "Token Is Illegal",
	Err_Token_Is_Expired:       "Token Is Expired",
}

//实现error接口
func (err *Err) Error() string {
	return err.Msg
}

func ErrMsg(errCode int) string {
	return errMap[errCode]
}
