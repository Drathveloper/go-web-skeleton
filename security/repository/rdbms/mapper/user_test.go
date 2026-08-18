package mapper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/security/domain"
	"github.com/Drathveloper/go-web-skeleton/security/repository/rdbms/entity"
	"github.com/Drathveloper/go-web-skeleton/security/repository/rdbms/mapper"
)

func TestEntityUserToDomainUser(t *testing.T) {
	tests := []struct {
		input    *entity.User
		expected *domain.User
		name     string
	}{
		{
			name: "user with multiple roles",
			input: &entity.User{
				Model:    gorm.Model{ID: 1234},
				Username: "johndoe",
				Password: "hashed_password",
				Roles:    []string{"admin", "editor"},
			},
			expected: &domain.User{
				ID:       1234,
				Username: "johndoe",
				Password: "hashed_password",
				Roles:    []commondomain.Role{"admin", "editor"},
			},
		},
		{
			name: "user with a single role",
			input: &entity.User{
				Model:    gorm.Model{ID: 4567},
				Username: "janedoe",
				Password: "another_hash",
				Roles:    []string{"viewer"},
			},
			expected: &domain.User{
				ID:       4567,
				Username: "janedoe",
				Password: "another_hash",
				Roles:    []commondomain.Role{"viewer"},
			},
		},
		{
			name: "user with no roles",
			input: &entity.User{
				Model:    gorm.Model{ID: 1341},
				Username: "noroles",
				Password: "some_hash",
				Roles:    []string{},
			},
			expected: &domain.User{
				ID:       1341,
				Username: "noroles",
				Password: "some_hash",
				Roles:    []commondomain.Role{},
			},
		},
		{
			name: "user with nil roles",
			input: &entity.User{
				Model:    gorm.Model{ID: 0},
				Username: "nilroles",
				Password: "hash",
				Roles:    nil,
			},
			expected: &domain.User{
				ID:       0,
				Username: "nilroles",
				Password: "hash",
				Roles:    []commondomain.Role{},
			},
		},
		{
			name: "user with empty fields",
			input: &entity.User{
				Model:    gorm.Model{ID: 0},
				Username: "",
				Password: "",
				Roles:    []string{},
			},
			expected: &domain.User{
				ID:       0,
				Username: "",
				Password: "",
				Roles:    []commondomain.Role{},
			},
		},
		{
			name:     "nil entity maps to zero value user instead of panicking",
			input:    nil,
			expected: &domain.User{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.EntityUserToDomainUser(tt.input)

			require.NotNil(t, result)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestEntityUsersToDomainUsers(t *testing.T) {
	tests := []struct {
		name     string
		users    []entity.User
		expected []domain.User
	}{
		{
			name:     "empty list",
			users:    []entity.User{},
			expected: []domain.User{},
		},
		{
			name:     "nil list",
			users:    nil,
			expected: []domain.User{},
		},
		{
			name: "one user in the list",
			users: []entity.User{
				{
					Username: "someUsername",
					Password: "somePassword",
					Roles:    []string{"someRole"},
					Model:    gorm.Model{ID: 1234},
				},
			},
			expected: []domain.User{
				{
					Username: "someUsername",
					Password: "somePassword",
					Roles:    []commondomain.Role{"someRole"},
					ID:       1234,
				},
			},
		},
		{
			name: "multiple users in the list",
			users: []entity.User{
				{
					Username: "someUsername1",
					Password: "somePassword1",
					Roles:    []string{"someRole1"},
					Model:    gorm.Model{ID: 123},
				},
				{
					Username: "someUsername2",
					Password: "somePassword2",
					Roles:    []string{"someRole2"},
					Model:    gorm.Model{ID: 456},
				},
			},
			expected: []domain.User{
				{
					Username: "someUsername1",
					Password: "somePassword1",
					Roles:    []commondomain.Role{"someRole1"},
					ID:       123,
				},
				{
					Username: "someUsername2",
					Password: "somePassword2",
					Roles:    []commondomain.Role{"someRole2"},
					ID:       456,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := mapper.EntityUsersToDomainUsers(tt.users)

			require.Equal(t, tt.expected, actual)
		})
	}
}

func TestDomainUserToEntityUser(t *testing.T) {
	tests := []struct {
		user     *domain.User
		expected *entity.User
		name     string
	}{
		{
			name: "map user should succeed",
			user: &domain.User{
				ID:       1234,
				Username: "johndoe",
				Password: "hashed_password",
				Roles:    []commondomain.Role{"admin", "editor"},
			},
			expected: &entity.User{
				Model:    gorm.Model{ID: 1234},
				Username: "johndoe",
				Password: "hashed_password",
				Roles:    []string{"admin", "editor"},
			},
		},
		{
			name: "map user without roles should produce an empty array, never nil",
			user: &domain.User{
				ID:       7,
				Username: "noroles",
				Password: "hash",
				Roles:    nil,
			},
			expected: &entity.User{
				Model:    gorm.Model{ID: 7},
				Username: "noroles",
				Password: "hash",
				Roles:    []string{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := mapper.DomainUserToEntityUser(tt.user)

			require.Equal(t, tt.expected, actual)
		})
	}
}
