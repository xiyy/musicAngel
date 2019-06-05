package httpserver

import (
	"context"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"musicAngel/bean"
	"musicAngel/database"
	"net"
	"net/http"
	"strconv"
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
)

func NewHttpServer(dbManager *database.DbManager) *HttpServer {
	return &HttpServer{server: &http.Server{}, dbManager: dbManager}
}
func (httpServer *HttpServer) Serve(listener net.Listener) error {
	log.Println("start serving")
	http.HandleFunc(SONG_LIST, httpServer.songList)
	http.HandleFunc(SONG_SINGER, httpServer.songsBySinger)
	http.HandleFunc(SONG_SONGNAME, httpServer.songBySongName)
	http.HandleFunc(USER_REGISTER, httpServer.register)
	http.HandleFunc(USER_UPDATE_USER_INFO, httpServer.updateUserInfo)
	http.HandleFunc(FAVORITE_OPRATION, httpServer.favoriteOperate)
	http.HandleFunc(FAVORITE_LIST, httpServer.favoriteList)
	return httpServer.server.Serve(listener)
}

func (httpServer *HttpServer) Stop() {
	log.Println("serve will stop")
	httpServer.server.Shutdown(context.TODO())
}

/**
支持get、post
http://localhost/song/list

*/
func (httpServer *HttpServer) songList(w http.ResponseWriter, r *http.Request) {
	err, contentBytes := httpServer.dbManager.QuerySongList()
	resp := new(Response)
	if err != nil {
		resp.Code = STATUS_DATABASE_ERROR
		resp.Msg = StatusText(STATUS_DATABASE_ERROR)
		resp.Data = ""
	} else {
		resp.Code = STATUS_OK
		resp.Msg = StatusText(STATUS_OK)
		resp.Data = string(contentBytes)
	}
	err, result := resp.jsonString()
	if err == nil {
		w.Write(result)
	} else {
		w.Write(resp.jsonError())
	}

}

/**
支持get、post
http://localhost/song/singer?singer=周杰伦
http://localhost/song/singer?singer=王小田
*/
func (httpServer *HttpServer) songsBySinger(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	singer := params["singer"][0]
	err, contentBytes := httpServer.dbManager.QuerySongsBySinger(singer)
	resp := new(Response)
	if err != nil {
		resp.Code = STATUS_DATABASE_ERROR
		resp.Msg = StatusText(STATUS_DATABASE_ERROR)
		resp.Data = ""
	} else {
		resp.Code = STATUS_OK
		resp.Msg = StatusText(STATUS_OK)
		resp.Data = string(contentBytes)
	}
	err, result := resp.jsonString()
	if err == nil {
		w.Write(result)
	} else {
		w.Write(resp.jsonError())
	}
}

/**
支持get、post
http://localhost/song/songname?songname=彩虹
http://localhost/song/songname?songname=库中没有这首歌曲
*/
func (httpServer *HttpServer) songBySongName(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	songname := params["songname"][0]
	err, contentBytes := httpServer.dbManager.QuerySongBySongName(songname)
	resp := new(Response)
	if err != nil {
		resp.Code = STATUS_DATABASE_ERROR
		resp.Msg = StatusText(STATUS_DATABASE_ERROR)
		resp.Data = ""
	} else {
		resp.Code = STATUS_OK
		resp.Msg = StatusText(STATUS_OK)
		resp.Data = string(contentBytes)
	}
	err, result := resp.jsonString()
	if err == nil {
		w.Write(result)
	} else {
		w.Write(resp.jsonError())
	}

}

/**
支持post，json数据放在post请求body中
http://localhost/user/register
{"account":"zhangxiaobei","password":"49ba59abbe56e057","registerdate":"2019-06-05 15:22:50","lastlogindate":"2019-06-05 15:22:50"} password使用md5加密

*/
func (httpServer *HttpServer) register(w http.ResponseWriter, r *http.Request) {
	resp := new(Response)
	if r.Method == "POST" {
		var account bean.Account
		err := json.NewDecoder(r.Body).Decode(&account)
		defer r.Body.Close()

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
	err, result := resp.jsonString()
	if err == nil {
		w.Write(result)
	} else {
		w.Write(resp.jsonError())
	}

}

/**
支持post，json数据放在post请求body中
http://localhost/user/update
{"account":"zhangxiaobei","phonenum":"15611067756","nickname":"喜洋洋","gender":1,"region":"北京","birthday":"1991-08-15"}
*/
func (httpServer *HttpServer) updateUserInfo(w http.ResponseWriter, r *http.Request) {
	resp := new(Response)
	if r.Method == "POST" {
		var user bean.User
		err := json.NewDecoder(r.Body).Decode(&user)
		defer r.Body.Close()
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

	err, result := resp.jsonString()
	if err == nil {
		w.Write(result)
	} else {
		w.Write(resp.jsonError())
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
	if r.Method == "POST" {
		var favoriteSongArray bean.FavoriteSongArray
		err := json.NewDecoder(r.Body).Decode(&favoriteSongArray)
		defer r.Body.Close()
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
	err, result := resp.jsonString()
	if err == nil {
		w.Write(result)
	} else {
		w.Write(resp.jsonError())
	}

}

/**
支持post、get
http://localhost/favorite/list?userid=9  http://localhost/favorite/list?userid=112233
*/
func (httpServer *HttpServer) favoriteList(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	userid := params["userid"][0]
	userIdInt,err:=strconv.Atoi(userid)
	if err!=nil {
		log.Fatal("strconv.Atoi(userid) error")
	}
	resp := new(Response)
	isUserExits:=httpServer.dbManager.IsUserExits(userIdInt)
	if isUserExits {
		data, err := httpServer.dbManager.QueryFavoriteSongsByUserId(userid)
		if err != nil {
			resp.Code = STATUS_DATABASE_ERROR
			resp.Msg = StatusText(STATUS_DATABASE_ERROR)
			resp.Data = false
		} else {
			resp.Code = STATUS_OK
			resp.Msg = StatusText(STATUS_OK)
			resp.Data = string(data)
		}
	}else {
		resp.Code = STATUS_ACCOUT_NOT_EXIT
		resp.Msg = StatusText(STATUS_ACCOUT_NOT_EXIT)
		resp.Data = false
	}

	err, result := resp.jsonString()
	if err == nil {
		w.Write(result)
	} else {
		w.Write(resp.jsonError())
	}
}
