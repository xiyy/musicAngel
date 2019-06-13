package bean

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type RequestLog struct {
	UserId     int
	Uuid       string
	Ip         string
	Appid      string
	Service    string
	Token      string
	PostBody   string
	Time       string
	Successful int
	RespCode   int
	RespMsg    string
}

func NewRequestLog(request *http.Request, postBody []byte, respCode int, respMsg string) *RequestLog {
	service := request.URL.Host + request.URL.Port() + request.URL.Path
	appid := request.Header.Get("appid")
	token := request.Header.Get("token")
	var postBodyStr string
	if postBody != nil {
		postBodyStr = string(postBody)
	}
	ip := func() string {
		clientIP := request.Header.Get("X-Forwarded-For")
		if index := strings.IndexByte(clientIP, ','); index >= 0 {
			clientIP = clientIP[0:index]
		}
		clientIP = strings.TrimSpace(clientIP)
		if clientIP != "" {
			return clientIP
		}
		clientIP = strings.TrimSpace(request.Header.Get("X-Real-Ip"))
		if clientIP != "" {
			return clientIP
		}
		return ""
	}()
	var userId int
	var uuid string
	var successful int
	if respCode == 0 {
		successful = 0
	} else {
		successful = 1
	}
	currentTimeStamp := strconv.FormatInt(time.Now().Unix(), 10)
	return &RequestLog{UserId: userId, Uuid: uuid, Ip: ip, Appid: appid, Service: service, Token: token, PostBody: postBodyStr, Time: currentTimeStamp, Successful: successful, RespCode: respCode, RespMsg: respMsg}
}
