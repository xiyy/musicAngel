package bean

type Account struct {
	Account       string `json:"account"`
	Password      string `json:"password"`
	RegisterDate  string `json:"registerdate"`
	LastLoginDate string `json:"lastlogindate"`
}
