package service

import (
	"strings"

	"refurbished-marketplace/services/users/internal/database"
	shareddb "refurbished-marketplace/shared/db"
)

func normalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

func isValidEmailShape(email string) bool {
	return strings.Contains(email, "@") && len(email) >= 3
}

func mapDBUser(u database.User) User {
	return User{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

func mapNotFound(err, notFoundErr error) error {
	return shareddb.MapErrNoRows(err, notFoundErr)
}

func mapInvalidCredentials(err error) error {
	return shareddb.MapErrNoRows(err, ErrInvalidCredentials)
}

func isPostgresUniqueViolation(err error) bool {
	return shareddb.IsUniqueViolation(err)
}
