package httpserver

import "encoding/json"

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

const (
	STATUS_OK                 = 0
	STATUS_DATABASE_ERROR     = -5001
	STATUS_JSON_ERROR         = -5002
	STATUS_JSON_DATA_ILLEGAL  = -5003
	STATUS_ACCOUT_NOT_EXIT    = -5004
	STATUS_METHOD_NOT_SUPPORT = -5005
	STATUS_ACCOUNT_HAS_EXITED = -5006
	STATUS_DATA_PARAM_ILLEGAL = -5007
	STATUS_DATE_IS_NULL       = -5008
	STATUS_APP_ID_IS_ILLEGAL  = -5009
	STATUS_TOKEN_ILLEGAL      = -5010
	STATUS_TOKEN_EXPIRES      = -5011
)

var statusText = map[int]string{
	STATUS_OK:                 "success",
	STATUS_DATABASE_ERROR:     "DataBase Error",
	STATUS_JSON_ERROR:         "Json Error",
	STATUS_JSON_DATA_ILLEGAL:  "Json Data Illegal",
	STATUS_ACCOUT_NOT_EXIT:    "Accout Not Exit",
	STATUS_METHOD_NOT_SUPPORT: "Method Not Support",
	STATUS_ACCOUNT_HAS_EXITED: "Account Has Exited",
	STATUS_DATA_PARAM_ILLEGAL: "Param Is Illegal",
	STATUS_DATE_IS_NULL:       "Data Is Null",
	STATUS_APP_ID_IS_ILLEGAL:  "AppId Is Illegal",
	STATUS_TOKEN_ILLEGAL:      "Token Illegal",
	STATUS_TOKEN_EXPIRES:      "Token Expires",
}

func StatusText(code int) string {
	return statusText[code]
}

/**
将Response转成json字符串，返回给客户端
1、Response所有属性的首字母要全部大写，否则json包在解析时，由于json包与Response不在一个包中，导致Response中的属性对json不可见，从而解析失败
2、Response所有属性后面要有json的key名称，否者也会解析失败
*/

func (resp *Response) ObjToBytes() (error, []byte) {
	bytes, err := json.Marshal(resp)
	return err, bytes
}
