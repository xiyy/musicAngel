package bean

const (
	FAVORITE_SONG_ADD    = 1
	FAVORITE_SONG_DELETE = 2
)

type FavoriteSong struct {
	MusicId string `json:"musicid"`
	UserId  int    `json:"userid"`
	Addtime string `json:"addtime"`
}

type FavoriteSongArray struct {
	OperateType int              `json:"operatetype"`
	SongArray   [] *FavoriteSong `json:songarray`
}
