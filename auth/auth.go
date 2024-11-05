package auth

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func registerUser(db *sql.DB,userInfo User,provider string) {

	user,_:= GetUser(db,userInfo.Email)
	if(user == nil){
		CreateUser(db,userInfo,provider)
	}else{
		fmt.Println("Already Exists")
	}

}

func SignJWT(user *User) error {
	claims := jwt.MapClaims{
		"name": user.Name,
		"email": user.Email,
		"iss": "oauth-app-golang",
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	}
	var err error
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,claims)
	user.Jwt,err = token.SignedString([]byte(os.Getenv("JWT_KEY")))
	if err != nil {
		return err;
	}
	return nil
}

func ParseJWT(token string)(*User,error){
	JWT,err := jwt.Parse(token,func(tok *jwt.Token)(interface{}, error){
		if _, ok := tok.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", tok.Header["alg"])
		}
		return []byte(os.Getenv("JWT_KEY")), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := JWT.Claims.(jwt.MapClaims); ok && JWT.Valid {

		user := &User{
			Name:  claims["name"].(string),
			Email: claims["email"].(string),
		}
		return user, nil
	} else {
		return nil, fmt.Errorf("invalid token")
	}
}

func GetAuthUser(db *sql.DB,token string) bool {
	claims,err := ParseJWT(token);
	if err != nil {
		return false;
	}
	user,err := GetUser(db,claims.Email);
	if err != nil {
		return false;
	}
	if user.Email != claims.Email {
		return false
	}
	return true;
}

func CreateUser(db *sql.DB, user User,provider string) (sql.Result,error) {
	query := `INSERT INTO Users (name, email, picture, provider) VALUES ($1, $2, $3, $4)`
	res,err := db.Exec(query, user.Name, user.Email, user.Picture, provider)
	if err != nil {
		return nil,err
	}
	return res,nil
}

func GetUser(db *sql.DB, email string) (*User, error) {

	query := `SELECT name, email, picture FROM Users WHERE email = $1`
	row := db.QueryRow(query,email)

	var user User
	err := row.Scan(&user.Name, &user.Email, &user.Picture)

	if err != nil {
		return nil, err
	}

	return &user, nil
}