package validate

import (
	"database/sql"
	"log"
	"musicAngel/bean"
	"musicAngel/config"
	"musicAngel/database"
	"net/http"
	"strconv"
	"time"
)

func CheckRequest(dbManager *database.DbManager, r *http.Request) error {
	appid := r.Header.Get("appid")
	token := r.Header.Get("token")
	if appid != config.APP_ID {
		return &bean.Err{Code: bean.Err_Token_Appid_Is_Illegal, Msg: bean.ErrMsg(bean.Err_Token_Appid_Is_Illegal)}
	}
	err, t := dbManager.QueryToken(appid, token)
	if err != nil {
		if err == sql.ErrNoRows {
			return &bean.Err{Code: bean.Err_Token_Is_Illegal, Msg: bean.ErrMsg(bean.Err_Token_Is_Illegal)}
		}
		return err
	}
	if t != nil {
		if t.TokenValue == "" {
			return &bean.Err{Code: bean.Err_Token_Is_Illegal, Msg: bean.ErrMsg(bean.Err_Token_Is_Illegal)}
		} else {
			tokenInt, _ := strconv.Atoi(t.Expire)
			currentTimeStamp := int(time.Now().Unix())
			log.Println(tokenInt, currentTimeStamp)
			if tokenInt <= currentTimeStamp {
				return &bean.Err{Code: bean.Err_Token_Is_Expired, Msg: bean.ErrMsg(bean.Err_Token_Is_Expired)}
			} else {
				return nil
			}
		}
	}
	return &bean.Err{Code: bean.Err_Token_Is_Illegal, Msg: bean.ErrMsg(bean.Err_Token_Is_Illegal)}
}
