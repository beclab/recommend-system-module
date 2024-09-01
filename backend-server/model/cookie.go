package model

type SettingReqModel struct {
	Domain string `json:"domain"`
}

type SettingRecordRespModel struct {
	Domain   string  `json:"domain" `
	Name     string  `json:"name"`
	Value    string  `json:"value" `
	Expires  float32 `json:"expirationDate"`
	Path     string  `json:"path"`
	Secure   bool    `json:"secure"`
	HttpOnly bool    `json:"httpOnly"`
	SameSite string  `json:"sameSite"`
	Other    string  `json:"other"`
}

type SettingDomainRespModel struct {
	Domain  string                   `json:"domain" `
	Account string                   `json:"account"`
	Records []SettingRecordRespModel `json:"records"`
}

type SettingResponseModel struct {
	Code    int                      `json:"code"`
	Message string                   `json:"message"`
	Data    []SettingDomainRespModel `json:"data"`
}
