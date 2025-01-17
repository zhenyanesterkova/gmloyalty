package user

type User struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	HashKey  string `json:"-"`
}
