package commands

type Login struct {
	Login    string
	Password string
}

type Register struct {
	Login    string
	Email    string
	Password string
}

type Logout struct {
	RefreshToken string
	AccessToken  string
}

type Refresh struct {
	RefreshToken string
}

type Validate struct {
	AccessToken string
}
