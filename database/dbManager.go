package database

import (
	"database/sql"
	"log"
	"musicAngel/bean"
	"musicAngel/config"
)

type DbManager struct {
	Db *sql.DB
}

func (dbManager *DbManager) Close() {
	dbManager.Db.Close()
}
func (dbManager *DbManager) QuerySongList() (error, []*bean.SongInfo) {
	rows, err := dbManager.Db.Query("select * from songinfo")
	if err != nil {
		return err, nil
	}
	defer rows.Close()
	var songInfoList []*bean.SongInfo

	for rows.Next() {
		song := new(bean.SongInfo)
		err = rows.Scan(&song.Id, &song.Music_id, &song.Mv_rid, &song.Name, &song.Song_url, &song.Artist, &song.Artid, &song.Singer, &song.Special, &song.Ridmd591, &song.Mp3size, &song.Artist_url, &song.Auther_url, &song.Playid, &song.Artist_pic, &song.Artist_pic240, &song.Path, &song.Mp3path, &song.Aacpath, &song.Wmadl, &song.Mp3dl, &song.Aacdl, &song.Lyric, &song.Lyric_zz, &song.Song_mp3_url)
		if err != nil {
			return err, nil
		}
		log.Println(*song)
		if song != nil {
			songInfoList = append(songInfoList, song)
		}
	}
	if songInfoList != nil {
		log.Println("songInfoList size:", len(songInfoList))
		return nil, songInfoList
	}
	return nil, nil
}

func (dbManager *DbManager) QuerySongsBySinger(singer string) (error, []*bean.SongInfo) {
	rows, err := dbManager.Db.Query("select * from songinfo where singer=?", singer)
	if err != nil {
		return err, nil
	}
	defer rows.Close()
	var songInfoList []*bean.SongInfo

	for rows.Next() {
		song := new(bean.SongInfo)
		err = rows.Scan(&song.Id, &song.Music_id, &song.Mv_rid, &song.Name, &song.Song_url, &song.Artist, &song.Artid, &song.Singer, &song.Special, &song.Ridmd591, &song.Mp3size, &song.Artist_url, &song.Auther_url, &song.Playid, &song.Artist_pic, &song.Artist_pic240, &song.Path, &song.Mp3path, &song.Aacpath, &song.Wmadl, &song.Mp3dl, &song.Aacdl, &song.Lyric, &song.Lyric_zz, &song.Song_mp3_url)
		if err != nil {
			return err, nil
		}
		log.Println(*song)
		if song != nil {
			songInfoList = append(songInfoList, song)
		}
	}
	log.Println("songInfoList size:", len(songInfoList))
	return nil, songInfoList
}

func (dbManager *DbManager) QuerySongBySongName(songName string) (error, []*bean.SongInfo) {
	rows, err := dbManager.Db.Query("select * from songinfo where name=?", songName)
	if err != nil {
		return err, nil
	}
	defer rows.Close()
	var songInfoList []*bean.SongInfo
	for rows.Next() {
		song := new(bean.SongInfo)
		err = rows.Scan(&song.Id, &song.Music_id, &song.Mv_rid, &song.Name, &song.Song_url, &song.Artist, &song.Artid, &song.Singer, &song.Special, &song.Ridmd591, &song.Mp3size, &song.Artist_url, &song.Auther_url, &song.Playid, &song.Artist_pic, &song.Artist_pic240, &song.Path, &song.Mp3path, &song.Aacpath, &song.Wmadl, &song.Mp3dl, &song.Aacdl, &song.Lyric, &song.Lyric_zz, &song.Song_mp3_url)
		if err != nil {
			return err, nil
		}
		log.Println(*song)
		if song != nil {
			songInfoList = append(songInfoList, song)
		}
	}
	log.Println("songInfoList size:", len(songInfoList))
	return nil, songInfoList
}

func (dbManager *DbManager) IsAccountExits(accountName string) bool {
	account := bean.Account{}
	var accountId int
	row := dbManager.Db.QueryRow("select * from account where account=?", accountName)
	row.Scan(&accountId, &account.Account, &account.Password, &account.RegisterDate, &account.LastLoginDate)
	if account.Account != "" {
		return true
	} else {
		return false
	}
}

func (dbManager *DbManager) IsUserExits(userId int) bool {
	user := bean.User{}
	var tempId int
	row := dbManager.Db.QueryRow("select * from user where userid=?", userId)
	row.Scan(&tempId, &user.AccountName, &user.PhoneNum, &user.NickName, &user.Gender, &user.Region, &user.Birthday)
	if user.AccountName != "" {
		return true
	}
	return false

}

//客户端发送的密码是密文，数据库中存储的也是密文
func (dbManager *DbManager) Register(account bean.Account) (error, bool) {
	if dbManager.IsAccountExits(account.Account) {
		return &bean.Err{Code: bean.Err_Account_Has_Exited, Msg: bean.ErrMsg(bean.Err_Account_Has_Exited)}, false
	}
	stmt, err := dbManager.Db.Prepare("insert into account(account,password,registerdate,lastlogindate)values(?,?,?,?)")
	if err != nil {
		return err, false
	}
	rs, err := stmt.Exec(account.Account, account.Password, account.RegisterDate, account.LastLoginDate)
	if err != nil {
		return err, false
	}
	//获得插入的id
	id, err := rs.LastInsertId()
	if err != nil {
		return err, false
	}
	log.Println("LastInsertId:", id)
	//可以获得影响行数
	affect, err := rs.RowsAffected()
	if err != nil {
		return err, false
	}
	log.Println("RowsAffected:", affect)
	//将用户名插入到user表中,user表中account字段和account表中account字段关联
	stmt, err = dbManager.Db.Prepare("insert into user(account,nickname) values(?,?)")
	if err != nil {
		return err, false
	}
	_, err = stmt.Exec(account.Account, "")
	if err != nil {
		return err, false
	}
	return nil, affect > 0
}

func (dbManager *DbManager) UpdateUserInfo(user bean.User) (error, bool) {
	isAccountExits := dbManager.IsAccountExits(user.AccountName)
	log.Println(user)
	if isAccountExits {
		gender := user.Gender
		nickName := user.NickName
		if (gender != 0 && gender != 1) || nickName == "" {
			return &bean.Err{Code: bean.Err_Data_ILLEGAL, Msg: bean.ErrMsg(bean.Err_Data_ILLEGAL)}, false
		}
		stmt, err := dbManager.Db.Prepare("update user set phonenum=?,nickname=?,gender=?,region=?,birthday=? where account=?")
		if err != nil {
			return err, false
		}
		rs, err := stmt.Exec(user.PhoneNum, user.NickName, user.Gender, user.Region, user.Birthday, user.AccountName)
		if err != nil {
			return err, false
		}
		//获得插入的id
		id, err := rs.LastInsertId()
		if err != nil {
			return err, false
		}
		log.Println("LastInsertId():", id)
		//可以获得影响行数
		affect, err := rs.RowsAffected()
		if err != nil {
			return err, false
		}
		log.Println("RowsAffected():", affect)
		return nil, affect > 0

	} else {
		return &bean.Err{Code: bean.Err_Account_Not_Exit, Msg: bean.ErrMsg(bean.Err_Account_Not_Exit)}, false
	}
}

func (dbManager *DbManager) OperateFavoriteSongs(operateType int, songs []*bean.FavoriteSong) (error, bool) {
	len := len(songs)
	if len >= 1 {
		if operateType == bean.FAVORITE_SONG_ADD {
			stmt, err := dbManager.Db.Prepare("insert into favorite(music_id,userid,addtime)values(?,?,?)")
			if err != nil {
				return err, false
			}
			for _, eachSong := range songs {
				rs, err := stmt.Exec(eachSong.MusicId, eachSong.UserId, eachSong.Addtime)
				if err != nil {
					return err, false
				}
				//获得插入的id
				id, err := rs.LastInsertId()
				if err != nil {
					return err, false
				}
				log.Println("LastInsertId():", id)
				//影响行数
				affect, err := rs.RowsAffected()
				if err != nil {
					return err, false
				}
				log.Println("RowsAffected():", affect)
			}
			return nil, true
		} else if operateType == bean.FAVORITE_SONG_DELETE {
			for _, eachSong := range songs {
				ret3, err := dbManager.Db.Exec("delete from favorite where music_id = ? and userid = ?", eachSong.MusicId, eachSong.UserId)
				if err != nil {
					return err, false
				}
				//可以获得影响行数
				affect, err := ret3.RowsAffected()
				if err != nil {
					return err, false
				}
				log.Println("RowsAffected():", affect)
			}
			return nil, true
		} else {
			return &bean.Err{Code: bean.Err_Data_Param_Illegal, Msg: bean.ErrMsg(bean.Err_Data_Param_Illegal)}, false
		}
	}
	return &bean.Err{Code: bean.Err_Data_Is_Null, Msg: bean.ErrMsg(bean.Err_Data_Is_Null)}, false

}

func (dbManager *DbManager) QueryFavoriteSongsByUserId(userid string) ([]*bean.FavoriteSong, error) {
	rows, err := dbManager.Db.Query("select * from favorite where userid=?", userid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var favoriteSongList []*bean.FavoriteSong

	for rows.Next() {
		favoriteSong := new(bean.FavoriteSong)
		id := 0
		err = rows.Scan(&id, &favoriteSong.MusicId, &favoriteSong.UserId, &favoriteSong.Addtime)
		if err != nil {
			return nil, err
		}
		if favoriteSong != nil {
			favoriteSongList = append(favoriteSongList, favoriteSong)
		}
	}
	log.Println("favoriteSongList size:", len(favoriteSongList))
	return favoriteSongList, nil
}

func (dbManager *DbManager) StoreToken(token, expireTime string) (error, bool) {
	tx, err := dbManager.Db.Begin()
	if err != nil {
		return err, false
	}
	var result sql.Result
	result, err = tx.Exec("insert into token (appid,token,expiretime) values(?,?,?)", config.APP_ID, token, expireTime)
	if err != nil {
		return err, false
	}
	affectRows, _ := result.RowsAffected()
	tx.Commit()
	return nil, affectRows > 0
}

func (dbManager *DbManager) QueryToken(appid, token string) (error, *bean.Token) {
	tokenValue := bean.Token{}
	row := dbManager.Db.QueryRow("select token,expiretime from token where appid=? and token=?", appid, token)
	err := row.Scan(&tokenValue.TokenValue, &tokenValue.Expire)
	if err != nil {
		return err, nil
	}
	log.Println(tokenValue.TokenValue, tokenValue.Expire)
	return nil, &tokenValue
}

func (dbManager *DbManager) SaveRequestLog(log *bean.RequestLog) error {
	tx, err := dbManager.Db.Begin()
	if err != nil {
		return err
	}
	_, err = tx.Exec("insert into requestlog (userid,uuid,ip,appid,service,token,postbody,requesttime,successful,respcode,respmsg) values (?,?,?,?,?,?,?,?,?,?,?)", log.UserId,
		log.Uuid, log.Ip, log.Appid, log.Service, log.Token, log.PostBody, log.Time, log.Successful, log.RespCode, log.RespMsg)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}
