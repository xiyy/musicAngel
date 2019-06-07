package bean

//首字符全部大写
type SongInfo struct {
	Id            int    `json:"id"`
	Music_id      string `json:"music_id"`
	Mv_rid        string `json:"mv_rid"`
	Name          string `json:"name"`
	Song_url      string `json:"song_url"`
	Artist        string `json:"artist"`
	Artid         string `json:"artid"`
	Singer        string `json:"singer"`
	Special       string `json:"special"`
	Ridmd591      string `json:"ridmd_591"`
	Mp3size       string `json:"mp_3_size"`
	Artist_url    string `json:"artist_url"`
	Auther_url    string `json:"auther_url"`
	Playid        string `json:"playid"`
	Artist_pic    string `json:"artist_pic"`
	Artist_pic240 string `json:"artist_pic_240"`
	Path          string `json:"path"`
	Mp3path       string `json:"mp_3_path"`
	Aacpath       string `json:"aacpath"`
	Wmadl         string `json:"wmadl"`
	Mp3dl         string `json:"mp_3_dl"`
	Aacdl         string `json:"aacdl"`
	Lyric         string `json:"lyric"`
	Lyric_zz      string `json:"lyric_zz"`
	Song_mp3_url  string `json:"song_mp_3_url"`
}
