package bean

type Token struct {
	TokenValue string `json:"token"`
	Expire     string `json:"expire"`
}

type AppConfigParam struct {
	AppId     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}
