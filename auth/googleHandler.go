package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

func HandleGoogleUser(db *sql.DB,accessToken string) (*User, error) {
	infoEndpoint := "https://www.googleapis.com/oauth2/v2/userinfo"
	res, err := http.Get(fmt.Sprintf("%s?access_token=%s", infoEndpoint, accessToken))
	if err != nil {
		return nil,err
	}
	defer res.Body.Close()
	
	var userInfo User
	if err := json.NewDecoder(res.Body).Decode(&userInfo); err != nil {
		return nil, err
	}
	
	registerUser(db,userInfo,"google")

	err = SignJWT(&userInfo);
	if(err != nil) {
		return nil,err;
	}
	return &userInfo,nil;
}