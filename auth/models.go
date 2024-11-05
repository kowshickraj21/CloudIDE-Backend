package auth

type User struct {
	Email    string `json:"email"`
	Picture  string `json:"picture"`
	Name     string `json:"name"`
	Jwt string `json:"jwt"`
}
