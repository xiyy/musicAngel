package bean

type User struct {
	AccountName string `json:"account"`
	PhoneNum    string `json:"phonenum"`
	NickName    string `json:nickname`
	Gender      int    `json:"gender"`
	Region      string `json:"region"`
	Birthday    string `json:"birthday"`
}
