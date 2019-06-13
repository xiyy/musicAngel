package httpserver

import (
	"context"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"musicAngel/bean"
	"musicAngel/config"
	"musicAngel/database"
	"musicAngel/encrypt"
	"musicAngel/validate"
	"net"
	"net/http"
	"strconv"
	"time"
)

type HttpServer struct {
	server    *http.Server
	dbManager *database.DbManager
}

const (
	SONG_LIST             = "/song/list"        //返回乐库所有歌曲
	SONG_SINGER           = "/song/singer"      //返回某位歌手的所有歌曲
	SONG_SONGNAME         = "/song/songname"    //根据歌曲名返回歌曲（可能有多项歌曲名相同的歌曲）
	USER_REGISTER         = "/user/register"    //用户注册
	USER_UPDATE_USER_INFO = "/user/update"      //更新用户信息
	FAVORITE_OPRATION     = "/favorite/operate" //添加或者取消收藏列表中某项歌曲
	FAVORITE_LIST         = "/favorite/list"    //返回某个用户的收藏列表
	TOKE_CREATE           = "/token/create"     //创建token
)

func NewHttpServer(dbManager *database.DbManager) *HttpServer {
	return &HttpServer{server: &http.Server{}, dbManager: dbManager}
}
func (httpServer *HttpServer) Serve(listener net.Listener) error {
	log.Println("start serving")
	//注册路由。每收到一个请求，http包内部会启一个协成执行Func，在Func中做具体的业务逻辑。
	// Func中如果只有数据库操作，不需要再启协成，操作完成直接返回。如果既有数据库操作，又有其他业务逻辑，数据库操作要放在单独的协成中执行，
	// 其他耗时的任务也要放在单独的协成中执行，保证所有的耗时任务都是异步执行，节省时间
	http.HandleFunc(SONG_LIST, httpServer.songList)
	http.HandleFunc(SONG_SINGER, httpServer.songsBySinger)
	http.HandleFunc(SONG_SONGNAME, httpServer.songBySongName)
	http.HandleFunc(USER_REGISTER, httpServer.register)
	http.HandleFunc(USER_UPDATE_USER_INFO, httpServer.updateUserInfo)
	http.HandleFunc(FAVORITE_OPRATION, httpServer.favoriteOperate)
	http.HandleFunc(FAVORITE_LIST, httpServer.favoriteList)
	http.HandleFunc(TOKE_CREATE, httpServer.createToken)
	return httpServer.server.Serve(listener)
}

func (httpServer *HttpServer) Stop() {
	log.Println("serve will stop")
	httpServer.server.Shutdown(context.TODO())
}

/**
支持post
http://localhost/token/create
{"app_id":"112233","app_secret":"QaD12&4l*WUPajk,anRM8Yz"}

*/
func (httpServer *HttpServer) createToken(w http.ResponseWriter, r *http.Request) {
	resp := new(Response)
	var postBodyBytes []byte
	if r.Method == "POST" {
		var appConfigParam bean.AppConfigParam
		postBodyBytes, _ = ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		err := json.Unmarshal(postBodyBytes, &appConfigParam)
		if err != nil {
			resp.Code = STATUS_DATA_PARAM_ILLEGAL
			resp.Msg = StatusText(STATUS_DATA_PARAM_ILLEGAL)
			resp.Data = ""
		}

		if appConfigParam.AppId == config.APP_ID && appConfigParam.AppSecret == config.APP_SECRETE {
			//appId+当前时间戳+appSecrete+token有效时间（一个小时）
			currentTimeStamp := time.Now().Unix()
			token := encrypt.Md5Encode(config.APP_ID + strconv.FormatInt(currentTimeStamp, 10) + config.APP_SECRETE + config.TOKEN_VALID_TIME)
			expire := strconv.FormatInt(currentTimeStamp+3600, 10)
			err, storeSuccess := httpServer.dbManager.StoreToken(token, expire)
			if err != nil {
				resp.Code = STATUS_DATABASE_ERROR
				resp.Msg = StatusText(STATUS_DATABASE_ERROR)
				resp.Data = ""
			} else {
				if storeSuccess {
					resp.Code = STATUS_OK
					resp.Msg = StatusText(STATUS_OK)
					resp.Data = &bean.Token{TokenValue: token, Expire: expire}
				}
			}
		} else { //应用非法
			resp.Code = STATUS_APP_ID_IS_ILLEGAL
			resp.Msg = StatusText(STATUS_APP_ID_IS_ILLEGAL)
			resp.Data = ""
		}
	} else {
		resp.Code = STATUS_METHOD_NOT_SUPPORT
		resp.Msg = StatusText(STATUS_METHOD_NOT_SUPPORT)
		resp.Data = false
	}
	//记录日志
	httpServer.dbManager.SaveRequestLog(bean.NewRequestLog(r, postBodyBytes, resp.Code, resp.Msg))
	err, result := resp.ObjToBytes()
	if err == nil {
		w.Write(result)
	}

}

/**
支持get、post
http://localhost/song/list

*/
func (httpServer *HttpServer) songList(w http.ResponseWriter, r *http.Request) {
	resp := new(Response)
	err := validate.CheckRequest(httpServer.dbManager, r)
	if err == nil {
		err, songInfoList := httpServer.dbManager.QuerySongList()
		if err != nil {
			resp.Code = STATUS_DATABASE_ERROR
			resp.Msg = StatusText(STATUS_DATABASE_ERROR)
			resp.Data = ""
		} else {
			resp.Code = STATUS_OK
			resp.Msg = StatusText(STATUS_OK)
			resp.Data = songInfoList
		}
	} else {
		v, ok := err.(*bean.Err)
		if ok {
			if v.Code == bean.Err_Token_Appid_Is_Illegal {
				resp.Code = STATUS_APP_ID_IS_ILLEGAL
				resp.Msg = StatusText(STATUS_APP_ID_IS_ILLEGAL)
				resp.Data = ""
			} else if v.Code == bean.Err_Token_Is_Illegal {
				resp.Code = STATUS_TOKEN_ILLEGAL
				resp.Msg = StatusText(STATUS_TOKEN_ILLEGAL)
				resp.Data = ""
			} else if v.Code == bean.Err_Token_Is_Expired {
				resp.Code = STATUS_TOKEN_EXPIRES
				resp.Msg = StatusText(STATUS_TOKEN_EXPIRES)
				resp.Data = ""
			}
		} else {
			resp.Code = STATUS_DATABASE_ERROR
			resp.Msg = StatusText(STATUS_DATABASE_ERROR)
			resp.Data = ""

		}
	}
	//记录日志
	httpServer.dbManager.SaveRequestLog(bean.NewRequestLog(r, nil, resp.Code, resp.Msg))
	err, result := resp.ObjToBytes()
	if err == nil {
		w.Write(result)
	}
}

/**
支持get、post
http://localhost/song/singer?singer=周杰伦
http://localhost/song/singer?singer=王小田
*/
func (httpServer *HttpServer) songsBySinger(w http.ResponseWriter, r *http.Request) {
	resp := new(Response)
	err := validate.CheckRequest(httpServer.dbManager, r)
	if err == nil {
		params := r.URL.Query()
		singer := params["singer"][0]
		err, songInfoList := httpServer.dbManager.QuerySongsBySinger(singer)

		if err != nil {
			resp.Code = STATUS_DATABASE_ERROR
			resp.Msg = StatusText(STATUS_DATABASE_ERROR)
			resp.Data = ""
		} else {
			resp.Code = STATUS_OK
			resp.Msg = StatusText(STATUS_OK)
			if songInfoList == nil {
				resp.Data = ""
			} else {
				resp.Data = songInfoList
			}

		}
	} else {
		v, ok := err.(*bean.Err)
		if ok {
			if v.Code == bean.Err_Token_Appid_Is_Illegal {
				resp.Code = STATUS_APP_ID_IS_ILLEGAL
				resp.Msg = StatusText(STATUS_APP_ID_IS_ILLEGAL)
				resp.Data = ""
			} else if v.Code == bean.Err_Token_Is_Illegal {
				resp.Code = STATUS_TOKEN_ILLEGAL
				resp.Msg = StatusText(STATUS_TOKEN_ILLEGAL)
				resp.Data = ""
			} else if v.Code == bean.Err_Token_Is_Expired {
				resp.Code = STATUS_TOKEN_EXPIRES
				resp.Msg = StatusText(STATUS_TOKEN_EXPIRES)
				resp.Data = ""
			}
		} else {
			resp.Code = STATUS_DATABASE_ERROR
			resp.Msg = StatusText(STATUS_DATABASE_ERROR)
			resp.Data = ""

		}
	}
	//记录日志
	httpServer.dbManager.SaveRequestLog(bean.NewRequestLog(r, nil, resp.Code, resp.Msg))
	err, result := resp.ObjToBytes()
	if err == nil {
		w.Write(result)
	}
}

/**
支持get、post
http://localhost/song/songname?songname=彩虹
http://localhost/song/songname?songname=库中没有这首歌曲
*/
func (httpServer *HttpServer) songBySongName(w http.ResponseWriter, r *http.Request) {
	resp := new(Response)
	err := validate.CheckRequest(httpServer.dbManager, r)
	if err == nil {
		params := r.URL.Query()
		songname := params["songname"][0]
		err, songInfoList := httpServer.dbManager.QuerySongBySongName(songname)

		if err != nil {
			resp.Code = STATUS_DATABASE_ERROR
			resp.Msg = StatusText(STATUS_DATABASE_ERROR)
			resp.Data = ""
		} else {
			resp.Code = STATUS_OK
			resp.Msg = StatusText(STATUS_OK)
			if songInfoList == nil {
				resp.Data = ""
			} else {
				resp.Data = songInfoList
			}
		}
	} else {
		v, ok := err.(*bean.Err)
		if ok {
			if v.Code == bean.Err_Token_Appid_Is_Illegal {
				resp.Code = STATUS_APP_ID_IS_ILLEGAL
				resp.Msg = StatusText(STATUS_APP_ID_IS_ILLEGAL)
				resp.Data = ""
			} else if v.Code == bean.Err_Token_Is_Illegal {
				resp.Code = STATUS_TOKEN_ILLEGAL
				resp.Msg = StatusText(STATUS_TOKEN_ILLEGAL)
				resp.Data = ""
			} else if v.Code == bean.Err_Token_Is_Expired {
				resp.Code = STATUS_TOKEN_EXPIRES
				resp.Msg = StatusText(STATUS_TOKEN_EXPIRES)
				resp.Data = ""
			}
		} else {
			resp.Code = STATUS_DATABASE_ERROR
			resp.Msg = StatusText(STATUS_DATABASE_ERROR)
			resp.Data = ""

		}
	}
	//记录日志
	httpServer.dbManager.SaveRequestLog(bean.NewRequestLog(r, nil, resp.Code, resp.Msg))
	err, result := resp.ObjToBytes()
	if err == nil {
		w.Write(result)
	}

}

/**
支持post，json数据放在post请求body中
http://localhost/user/register
{"account":"zhangxiaobei","password":"49ba59abbe56e057","registerdate":"2019-06-05 15:22:50","lastlogindate":"2019-06-05 15:22:50"} password使用md5加密

*/
func (httpServer *HttpServer) register(w http.ResponseWriter, r *http.Request) {
	resp := new(Response)
	var postBodyBytes []byte
	err := validate.CheckRequest(httpServer.dbManager, r)
	if err == nil {
		if r.Method == "POST" {
			var account bean.Account
			postBodyBytes, _ = ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			err := json.Unmarshal(postBodyBytes, &account)
			if err != nil {
				resp.Code = STATUS_JSON_DATA_ILLEGAL
				resp.Msg = StatusText(STATUS_JSON_DATA_ILLEGAL)
				resp.Data = false
			} else {
				err, _ := httpServer.dbManager.Register(account)
				if err != nil {
					value, ok := err.(*bean.Err) //断言
					if ok {
						if value.Code == bean.Err_Account_Has_Exited {
							resp.Code = STATUS_ACCOUNT_HAS_EXITED
							resp.Msg = StatusText(STATUS_ACCOUNT_HAS_EXITED)
							resp.Data = false
						}
					} else {
						resp.Code = STATUS_DATABASE_ERROR
						resp.Msg = StatusText(STATUS_DATABASE_ERROR)
						resp.Data = false
					}

				} else {
					resp.Code = STATUS_OK
					resp.Msg = StatusText(STATUS_OK)
					resp.Data = true
				}
			}

		} else {
			resp.Code = STATUS_METHOD_NOT_SUPPORT
			resp.Msg = StatusText(STATUS_METHOD_NOT_SUPPORT)
			resp.Data = false
		}
	} else {
		v, ok := err.(*bean.Err)
		if ok {
			if v.Code == bean.Err_Token_Appid_Is_Illegal {
				resp.Code = STATUS_APP_ID_IS_ILLEGAL
				resp.Msg = StatusText(STATUS_APP_ID_IS_ILLEGAL)
				resp.Data = ""
			} else if v.Code == bean.Err_Token_Is_Illegal {
				resp.Code = STATUS_TOKEN_ILLEGAL
				resp.Msg = StatusText(STATUS_TOKEN_ILLEGAL)
				resp.Data = ""
			} else if v.Code == bean.Err_Token_Is_Expired {
				resp.Code = STATUS_TOKEN_EXPIRES
				resp.Msg = StatusText(STATUS_TOKEN_EXPIRES)
				resp.Data = ""
			}
		} else {
			resp.Code = STATUS_DATABASE_ERROR
			resp.Msg = StatusText(STATUS_DATABASE_ERROR)
			resp.Data = ""

		}
	}
	//记录日志
	httpServer.dbManager.SaveRequestLog(bean.NewRequestLog(r, postBodyBytes, resp.Code, resp.Msg))
	err, result := resp.ObjToBytes()
	if err == nil {
		w.Write(result)
	}

}

/**
支持post，json数据放在post请求body中
http://localhost/user/update
{"account":"zhangxiaobei","phonenum":"15611067756","nickname":"喜洋洋","gender":1,"region":"北京","birthday":"1991-08-15"}
*/
func (httpServer *HttpServer) updateUserInfo(w http.ResponseWriter, r *http.Request) {
	resp := new(Response)
	var postBodyBytes []byte
	err := validate.CheckRequest(httpServer.dbManager, r)
	if err == nil {
		if r.Method == "POST" {
			var user bean.User
			postBodyBytes, _ = ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			err := json.Unmarshal(postBodyBytes, &user)
			if err != nil {
				resp.Code = STATUS_JSON_DATA_ILLEGAL
				resp.Msg = StatusText(STATUS_JSON_DATA_ILLEGAL)
				resp.Data = false
			} else {
				err, _ := httpServer.dbManager.UpdateUserInfo(user)
				if err != nil {
					value, ok := err.(*bean.Err) //断言
					if ok {
						if value.Code == bean.Err_Data_ILLEGAL {
							resp.Code = STATUS_JSON_DATA_ILLEGAL
							resp.Msg = StatusText(STATUS_JSON_DATA_ILLEGAL)
							resp.Data = false
						}
						if value.Code == bean.Err_Account_Not_Exit {
							resp.Code = STATUS_ACCOUT_NOT_EXIT
							resp.Msg = StatusText(STATUS_ACCOUT_NOT_EXIT)
							resp.Data = false
						}
					} else {
						resp.Code = STATUS_DATABASE_ERROR
						resp.Msg = StatusText(STATUS_DATABASE_ERROR)
						resp.Data = false
					}

				} else {
					resp.Code = STATUS_OK
					resp.Msg = StatusText(STATUS_OK)
					resp.Data = true
				}
			}
		} else {
			resp.Code = STATUS_METHOD_NOT_SUPPORT
			resp.Msg = StatusText(STATUS_METHOD_NOT_SUPPORT)
			resp.Data = false
		}
	} else {
		v, ok := err.(*bean.Err)
		if ok {
			if v.Code == bean.Err_Token_Appid_Is_Illegal {
				resp.Code = STATUS_APP_ID_IS_ILLEGAL
				resp.Msg = StatusText(STATUS_APP_ID_IS_ILLEGAL)
				resp.Data = ""
			} else if v.Code == bean.Err_Token_Is_Illegal {
				resp.Code = STATUS_TOKEN_ILLEGAL
				resp.Msg = StatusText(STATUS_TOKEN_ILLEGAL)
				resp.Data = ""
			} else if v.Code == bean.Err_Token_Is_Expired {
				resp.Code = STATUS_TOKEN_EXPIRES
				resp.Msg = StatusText(STATUS_TOKEN_EXPIRES)
				resp.Data = ""
			}
		} else {
			resp.Code = STATUS_DATABASE_ERROR
			resp.Msg = StatusText(STATUS_DATABASE_ERROR)
			resp.Data = ""

		}
	}
	//记录日志
	httpServer.dbManager.SaveRequestLog(bean.NewRequestLog(r, postBodyBytes, resp.Code, resp.Msg))
	err, result := resp.ObjToBytes()
	if err == nil {
		w.Write(result)
	}

}

/**
支持post请求
http://localhost/favorite/operate
{"operatetype":1,"songarray":[{"musicid":"64854183","userid":9,"addtime":"2019-06-05 19:22:20"},{"musicid":"64949656","userid":9,"addtime":"2019-06-05 19:22:20"}]}  添加收藏
{"operatetype":2,"songarray":[{"musicid":"64854183","userid":9}]}  {"operatetype":2,"songarray":[]} 取消收藏
*/
func (httpServer *HttpServer) favoriteOperate(w http.ResponseWriter, r *http.Request) {
	resp := new(Response)
	var postBodyBytes []byte
	err := validate.CheckRequest(httpServer.dbManager, r)
	if err == nil {
		if r.Method == "POST" {
			var favoriteSongArray bean.FavoriteSongArray
			postBodyBytes, _ = ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			err := json.Unmarshal(postBodyBytes, &favoriteSongArray)
			if err != nil {
				resp.Code = STATUS_JSON_DATA_ILLEGAL
				resp.Msg = StatusText(STATUS_JSON_DATA_ILLEGAL)
				resp.Data = false
			} else {
				operateType := favoriteSongArray.OperateType
				songArray := favoriteSongArray.SongArray
				err, _ := httpServer.dbManager.OperateFavoriteSongs(operateType, songArray)
				if err != nil {
					value, ok := err.(*bean.Err) //断言
					if ok {
						if value.Code == bean.Err_Data_Param_Illegal {
							resp.Code = STATUS_DATA_PARAM_ILLEGAL
							resp.Msg = StatusText(STATUS_DATA_PARAM_ILLEGAL)
							resp.Data = false
						} else if value.Code == bean.Err_Data_Is_Null {
							resp.Code = STATUS_DATE_IS_NULL
							resp.Msg = StatusText(STATUS_DATE_IS_NULL)
							resp.Data = false
						}

					} else {
						resp.Code = STATUS_DATABASE_ERROR
						resp.Msg = StatusText(STATUS_DATABASE_ERROR)
						resp.Data = false
					}
				} else {
					resp.Code = STATUS_OK
					resp.Msg = StatusText(STATUS_OK)
					resp.Data = true
				}
			}
		} else {
			resp.Code = STATUS_METHOD_NOT_SUPPORT
			resp.Msg = StatusText(STATUS_METHOD_NOT_SUPPORT)
			resp.Data = false
		}
	} else {
		v, ok := err.(*bean.Err)
		if ok {
			if v.Code == bean.Err_Token_Appid_Is_Illegal {
				resp.Code = STATUS_APP_ID_IS_ILLEGAL
				resp.Msg = StatusText(STATUS_APP_ID_IS_ILLEGAL)
				resp.Data = ""
			} else if v.Code == bean.Err_Token_Is_Illegal {
				resp.Code = STATUS_TOKEN_ILLEGAL
				resp.Msg = StatusText(STATUS_TOKEN_ILLEGAL)
				resp.Data = ""
			} else if v.Code == bean.Err_Token_Is_Expired {
				resp.Code = STATUS_TOKEN_EXPIRES
				resp.Msg = StatusText(STATUS_TOKEN_EXPIRES)
				resp.Data = ""
			}
		} else {
			resp.Code = STATUS_DATABASE_ERROR
			resp.Msg = StatusText(STATUS_DATABASE_ERROR)
			resp.Data = ""

		}
	}
	//记录日志
	httpServer.dbManager.SaveRequestLog(bean.NewRequestLog(r, postBodyBytes, resp.Code, resp.Msg))
	err, result := resp.ObjToBytes()
	if err == nil {
		w.Write(result)
	}

}

/**
支持post、get
http://localhost/favorite/list?userid=9  http://localhost/favorite/list?userid=112233
*/
func (httpServer *HttpServer) favoriteList(w http.ResponseWriter, r *http.Request) {
	resp := new(Response)
	err := validate.CheckRequest(httpServer.dbManager, r)
	if err == nil {
		params := r.URL.Query()
		userid := params["userid"][0]
		userIdInt, err := strconv.Atoi(userid)
		if err != nil {
			log.Fatal("strconv.Atoi(userid) error")
		}
		isUserExits := httpServer.dbManager.IsUserExits(userIdInt)
		if isUserExits {
			favoriteSongList, err := httpServer.dbManager.QueryFavoriteSongsByUserId(userid)
			if err != nil {
				resp.Code = STATUS_DATABASE_ERROR
				resp.Msg = StatusText(STATUS_DATABASE_ERROR)
				resp.Data = false
			} else {
				resp.Code = STATUS_OK
				resp.Msg = StatusText(STATUS_OK)
				if favoriteSongList == nil {
					resp.Data = ""
				} else {
					resp.Data = favoriteSongList
				}
			}
		} else {
			resp.Code = STATUS_ACCOUT_NOT_EXIT
			resp.Msg = StatusText(STATUS_ACCOUT_NOT_EXIT)
			resp.Data = false
		}

	} else {
		v, ok := err.(*bean.Err)
		if ok {
			if v.Code == bean.Err_Token_Appid_Is_Illegal {
				resp.Code = STATUS_APP_ID_IS_ILLEGAL
				resp.Msg = StatusText(STATUS_APP_ID_IS_ILLEGAL)
				resp.Data = ""
			} else if v.Code == bean.Err_Token_Is_Illegal {
				resp.Code = STATUS_TOKEN_ILLEGAL
				resp.Msg = StatusText(STATUS_TOKEN_ILLEGAL)
				resp.Data = ""
			} else if v.Code == bean.Err_Token_Is_Expired {
				resp.Code = STATUS_TOKEN_EXPIRES
				resp.Msg = StatusText(STATUS_TOKEN_EXPIRES)
				resp.Data = ""
			}
		} else {
			resp.Code = STATUS_DATABASE_ERROR
			resp.Msg = StatusText(STATUS_DATABASE_ERROR)
			resp.Data = ""

		}
	}
	//记录日志
	httpServer.dbManager.SaveRequestLog(bean.NewRequestLog(r, nil, resp.Code, resp.Msg))
	err, result := resp.ObjToBytes()
	if err == nil {
		w.Write(result)
	}
}
