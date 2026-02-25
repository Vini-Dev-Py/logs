package usecase

import (
	"errors"
	"logs-bff/internal/domain/model"
	"logs-bff/internal/ports/out"

	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct{ Users out.UserRepository }

func (u AuthUsecase) Login(email, password string) (model.User, error) {
	user, err := u.Users.FindByEmail(nil, email)
	if err != nil {
		return model.User{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return model.User{}, errors.New("invalid credentials")
	}
	return user, nil
}
