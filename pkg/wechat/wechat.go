package wechat

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type SessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// GetWechatSession 获取微信session
func GetWechatSession(code, appID, appSecret string) (*SessionResponse, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appID,
		appSecret,
		code,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var session SessionResponse
	if err := json.Unmarshal(body, &session); err != nil {
		return nil, err
	}

	if session.ErrCode != 0 {
		return nil, fmt.Errorf("wechat error: %d - %s", session.ErrCode, session.ErrMsg)
	}

	return &session, nil
}
