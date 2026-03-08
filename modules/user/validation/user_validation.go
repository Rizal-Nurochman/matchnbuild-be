package validation

import (
	"github.com/Rizal-Nurochman/matchnbuild/modules/user/dto"
	"github.com/go-playground/validator/v10"
)

type UserValidation struct {
	validate *validator.Validate
}

func NewUserValidation() *UserValidation {
	validate := validator.New()

	validate.RegisterValidation("name", validateName)

	return &UserValidation{
		validate: validate,
	}
}

func (v *UserValidation) ValidateUserCreateRequest(req dto.UserCreateRequest) error {
	return v.validate.Struct(req)
}

func (v *UserValidation) ValidateUserUpdateRequest(req dto.UserUpdateRequest) error {
	return v.validate.Struct(req)
}

func validateName(fl validator.FieldLevel) bool {
	name := fl.Field().String()
	// Name should not be empty and not too long
	return len(name) > 0 && len(name) <= 100
}
