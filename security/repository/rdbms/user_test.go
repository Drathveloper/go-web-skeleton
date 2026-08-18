package rdbms_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Drathveloper/go-web-skeleton/common/database"
	commondomain "github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/security/domain"
	"github.com/Drathveloper/go-web-skeleton/security/repository/rdbms"
)

// setupMockDB puts go-sqlmock behind a real *gorm.DB, so the assertions below are
// on the SQL GORM actually emits rather than on a hand written string.
// one-method rdbms.PostgresClient exists for.
//
//nolint:ireturn
func setupMockDB(t *testing.T) (sqlmock.Sqlmock, *gorm.DB) {
	t.Helper()

	// sqlmock.QueryMatcherRegexp is the default matcher; it is stated explicitly
	// so the expectations below read as the regexes they are.
	sqlDB, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	require.NoError(t, err)

	return mock, gormDB
}

// mockPostgresClient adapts *gorm.DB to the rdbms.PostgresClient interface. That
// interface has a single method precisely so GORM can be swapped for go-sqlmock here.
type mockPostgresClient struct {
	db *gorm.DB
}

func (m *mockPostgresClient) WithContext(ctx context.Context) *gorm.DB {
	return m.db.WithContext(ctx)
}

func TestUser_FindUserByUsername(t *testing.T) {
	// GORM emits the First() limit as a bound argument, not as a literal, so the
	// query takes two placeholders: the username and the limit.
	const queryRegex = `^SELECT \* FROM "users" WHERE username = \$1 ` +
		`AND "users"\."deleted_at" IS NULL ORDER BY "users"\."id" LIMIT \$2$`

	tests := []struct {
		wantErrIs  error
		mockSetup  func(mock sqlmock.Sqlmock)
		wantUser   *domain.User
		name       string
		username   string
		wantErrMsg string
		wantErr    bool
	}{
		{
			name:     "find user by username succeed",
			username: "johndoe",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"id", "username", "password", "roles"}).
					AddRow(1, "johndoe", "1234", "{admin,editor}")

				mock.ExpectQuery(queryRegex).
					WithArgs("johndoe", 1).
					WillReturnRows(rows)
			},
			wantUser: &domain.User{
				ID:       1,
				Username: "johndoe",
				Password: "1234",
				Roles:    []commondomain.Role{"admin", "editor"},
			},
		},
		{
			name:     "find user by username failed when record not present",
			username: "ghost",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(queryRegex).
					WithArgs("ghost", 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr:    true,
			wantErrIs:  database.ErrRecordNotFound,
			wantErrMsg: "repository find user by username failed: database record not found",
		},
		{
			name:     "find user by username translates an empty result set into ErrRecordNotFound",
			username: "ghost",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(queryRegex).
					WithArgs("ghost", 1).
					WillReturnRows(mock.NewRows([]string{"id", "username", "password", "roles"}))
			},
			wantErr:    true,
			wantErrIs:  database.ErrRecordNotFound,
			wantErrMsg: "repository find user by username failed: database record not found",
		},
		{
			name:     "find user by username failed when unexpected error",
			username: "johndoe",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(queryRegex).
					WithArgs("johndoe", 1).
					WillReturnError(errors.New("connection refused"))
			},
			wantErr:    true,
			wantErrMsg: "repository find user by username failed: connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, gormDB := setupMockDB(t)
			tt.mockSetup(mock)

			repo := rdbms.NewUser(&mockPostgresClient{db: gormDB})

			got, err := repo.FindUserByUsername(context.Background(), tt.username)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, got)

				if tt.wantErrIs != nil {
					require.ErrorIs(t, err, tt.wantErrIs)
				} else {
					require.NotErrorIs(t, err, database.ErrRecordNotFound)
				}
				require.Equal(t, tt.wantErrMsg, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantUser, got)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUser_FindAllUsers(t *testing.T) {
	tests := []struct {
		mockSetup  func(mock sqlmock.Sqlmock)
		pagination *commondomain.Pagination
		name       string
		wantErrMsg string
		wantUsers  []domain.User
		wantErr    bool
	}{
		{
			name:       "find all users succeed",
			pagination: &commondomain.Pagination{Page: 1, Size: 10},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"id", "username", "password", "roles"}).
					AddRow(1, "johndoe", "1234", "{admin,editor}").
					AddRow(2, "jane", "5678", "{viewer,editor}")

				// Page one has no OFFSET: GORM omits a zero offset.
				mock.ExpectQuery(`^SELECT \* FROM "users" WHERE "users"\."deleted_at" IS NULL LIMIT \$1$`).
					WithArgs(10).
					WillReturnRows(rows)
			},
			wantUsers: []domain.User{
				{
					ID:       1,
					Username: "johndoe",
					Password: "1234",
					Roles:    []commondomain.Role{"admin", "editor"},
				},
				{
					ID:       2,
					Username: "jane",
					Password: "5678",
					Roles:    []commondomain.Role{"viewer", "editor"},
				},
			},
		},
		{
			name:       "find all users applies the offset from page two on",
			pagination: &commondomain.Pagination{Page: 3, Size: 10},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"id", "username", "password", "roles"}).
					AddRow(21, "user21", "pwd", "{admin}")

				mock.ExpectQuery(
					`^SELECT \* FROM "users" WHERE "users"\."deleted_at" IS NULL LIMIT \$1 OFFSET \$2$`).
					WithArgs(10, 20).
					WillReturnRows(rows)
			},
			wantUsers: []domain.User{
				{
					ID:       21,
					Username: "user21",
					Password: "pwd",
					Roles:    []commondomain.Role{"admin"},
				},
			},
		},
		{
			name:       "find all users falls back to the default page size on a zero pagination",
			pagination: &commondomain.Pagination{},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"id", "username", "password", "roles"})

				mock.ExpectQuery(`^SELECT \* FROM "users" WHERE "users"\."deleted_at" IS NULL LIMIT \$1$`).
					WithArgs(10).
					WillReturnRows(rows)
			},
			wantUsers: []domain.User{},
		},
		{
			name:       "find all users caps the page size at the maximum",
			pagination: &commondomain.Pagination{Page: 1, Size: 5000},
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"id", "username", "password", "roles"})

				mock.ExpectQuery(`^SELECT \* FROM "users" WHERE "users"\."deleted_at" IS NULL LIMIT \$1$`).
					WithArgs(100).
					WillReturnRows(rows)
			},
			wantUsers: []domain.User{},
		},
		{
			name:       "find all users should succeed when records not present",
			pagination: &commondomain.Pagination{Page: 1, Size: 10},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`^SELECT \* FROM "users" WHERE "users"\."deleted_at" IS NULL LIMIT \$1$`).
					WithArgs(10).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantErr:   false,
			wantUsers: []domain.User{},
		},
		{
			name:       "find all users failed when unexpected error",
			pagination: &commondomain.Pagination{Page: 1, Size: 10},
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`^SELECT \* FROM "users" WHERE "users"\."deleted_at" IS NULL LIMIT \$1$`).
					WithArgs(10).
					WillReturnError(errors.New("connection refused"))
			},
			wantErr:    true,
			wantErrMsg: "repository find all users failed: connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, gormDB := setupMockDB(t)
			tt.mockSetup(mock)

			repo := rdbms.NewUser(&mockPostgresClient{db: gormDB})

			got, err := repo.FindAllUsers(context.Background(), tt.pagination)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
				require.Equal(t, tt.wantErrMsg, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantUsers, got)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUser_FindAllUserLookups(t *testing.T) {
	// The lookup projects only the two columns a select box needs; the password
	// hash of every user must not travel just to fill a dropdown.
	const queryRegex = `^SELECT "id","username" FROM "users" WHERE "users"\."deleted_at" IS NULL$`

	tests := []struct {
		mockSetup  func(mock sqlmock.Sqlmock)
		name       string
		wantErrMsg string
		wantUsers  []domain.User
		wantErr    bool
	}{
		{
			name: "find all user lookups succeed",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"id", "username"}).
					AddRow(1, "johndoe").
					AddRow(2, "jane")

				mock.ExpectQuery(queryRegex).
					WithoutArgs().
					WillReturnRows(rows)
			},
			wantUsers: []domain.User{
				{ID: 1, Username: "johndoe", Roles: []commondomain.Role{}},
				{ID: 2, Username: "jane", Roles: []commondomain.Role{}},
			},
		},
		{
			name: "find all user lookups returns an empty list when there are no users",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(queryRegex).
					WithoutArgs().
					WillReturnRows(mock.NewRows([]string{"id", "username"}))
			},
			wantUsers: []domain.User{},
		},
		{
			name: "find all user lookups failed when unexpected error",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(queryRegex).
					WithoutArgs().
					WillReturnError(errors.New("connection refused"))
			},
			wantErr:    true,
			wantErrMsg: "repository find all users failed: connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, gormDB := setupMockDB(t)
			tt.mockSetup(mock)

			repo := rdbms.NewUser(&mockPostgresClient{db: gormDB})

			got, err := repo.FindAllUserLookups(context.Background())

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, got)
				require.Equal(t, tt.wantErrMsg, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantUsers, got)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUser_FindUserByID(t *testing.T) {
	const queryRegex = `^SELECT \* FROM "users" WHERE "users"\."id" = \$1 ` +
		`AND "users"\."deleted_at" IS NULL ORDER BY "users"\."id" LIMIT \$2$`

	tests := []struct {
		wantErrIs  error
		mockSetup  func(mock sqlmock.Sqlmock)
		wantUser   *domain.User
		name       string
		wantErrMsg string
		userID     uint
		wantErr    bool
	}{
		{
			name: "find user by id succeed",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"id", "username", "password", "roles"}).
					AddRow(1, "johndoe", "1234", "{admin,editor}")

				mock.ExpectQuery(queryRegex).
					WithArgs(1, 1).
					WillReturnRows(rows)
			},
			userID: 1,
			wantUser: &domain.User{
				ID:       1,
				Username: "johndoe",
				Password: "1234",
				Roles:    []commondomain.Role{"admin", "editor"},
			},
		},
		{
			name: "find user by id failed when record not present",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(queryRegex).
					WithArgs(1, 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			userID:     1,
			wantErr:    true,
			wantErrIs:  database.ErrRecordNotFound,
			wantErrMsg: "repository find user by id failed: database record not found",
		},
		{
			name: "find user by id translates an empty result set into ErrRecordNotFound",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(queryRegex).
					WithArgs(42, 1).
					WillReturnRows(mock.NewRows([]string{"id", "username", "password", "roles"}))
			},
			userID:     42,
			wantErr:    true,
			wantErrIs:  database.ErrRecordNotFound,
			wantErrMsg: "repository find user by id failed: database record not found",
		},
		{
			name: "find user by id failed when unexpected error",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(queryRegex).
					WithArgs(1, 1).
					WillReturnError(errors.New("connection refused"))
			},
			userID:     1,
			wantErr:    true,
			wantErrMsg: "repository find user by id failed: connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, gormDB := setupMockDB(t)
			tt.mockSetup(mock)

			repo := rdbms.NewUser(&mockPostgresClient{db: gormDB})

			got, err := repo.FindUserByID(context.Background(), tt.userID)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, got)

				if tt.wantErrIs != nil {
					require.ErrorIs(t, err, tt.wantErrIs)
				} else {
					require.NotErrorIs(t, err, database.ErrRecordNotFound)
				}
				require.Equal(t, tt.wantErrMsg, err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.wantUser, got)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUser_CreateUser(t *testing.T) {
	const queryRegex = `^INSERT INTO "users" ` +
		`\("created_at","updated_at","deleted_at","username","password","roles"\) ` +
		`VALUES \(\$1,\$2,\$3,\$4,\$5,\$6\) RETURNING "id"$`

	tests := []struct {
		mockSetup  func(mock sqlmock.Sqlmock)
		name       string
		user       *domain.User
		wantErrMsg string
		wantErr    bool
	}{
		{
			name: "create user succeed",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(queryRegex).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, "johndoe", "1234", `{"admin","editor"}`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()
			},
			user: &domain.User{
				Username: "johndoe",
				Password: "1234",
				Roles:    []commondomain.Role{"admin", "editor"},
			},
		},
		{
			name: "create should fail when username already exists",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(queryRegex).
					WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), nil, "johndoe", "1234", `{"admin","editor"}`).
					WillReturnError(gorm.ErrCheckConstraintViolated)
				mock.ExpectRollback()
			},
			user: &domain.User{
				Username: "johndoe",
				Password: "1234",
				Roles:    []commondomain.Role{"admin", "editor"},
			},
			wantErr:    true,
			wantErrMsg: "repository create user failed: violates check constraint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, gormDB := setupMockDB(t)
			tt.mockSetup(mock)

			repo := rdbms.NewUser(&mockPostgresClient{db: gormDB})

			err := repo.CreateUser(context.Background(), tt.user)

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.wantErrMsg, err.Error())
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUser_UpdateUser(t *testing.T) {
	// Updates() adds the primary key of the entity to the explicit Where, so the
	// statement carries the id twice and the soft delete guard on top.
	const queryRegex = `^UPDATE "users" SET "updated_at"=\$1,"username"=\$2,"password"=\$3,"roles"=\$4 ` +
		`WHERE id = \$5 AND "users"\."deleted_at" IS NULL AND "id" = \$6$`

	tests := []struct {
		mockSetup  func(mock sqlmock.Sqlmock)
		name       string
		user       *domain.User
		wantErrMsg string
		wantErr    bool
	}{
		{
			name: "update user succeed",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(queryRegex).
					WithArgs(sqlmock.AnyArg(), "johndoe", "1234", `{"admin","editor"}`, 1, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			user: &domain.User{
				ID:       1,
				Username: "johndoe",
				Password: "1234",
				Roles:    []commondomain.Role{"admin", "editor"},
			},
		},
		{
			name: "update should fail when the database rejects the statement",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectExec(queryRegex).
					WithArgs(sqlmock.AnyArg(), "johndoe", "1234", `{"admin","editor"}`, 1, 1).
					WillReturnError(gorm.ErrCheckConstraintViolated)
				mock.ExpectRollback()
			},
			user: &domain.User{
				Username: "johndoe",
				Password: "1234",
				Roles:    []commondomain.Role{"admin", "editor"},
				ID:       1,
			},
			wantErr:    true,
			wantErrMsg: "repository update user failed: violates check constraint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, gormDB := setupMockDB(t)
			tt.mockSetup(mock)

			repo := rdbms.NewUser(&mockPostgresClient{db: gormDB})

			err := repo.UpdateUser(context.Background(), tt.user)

			if tt.wantErr {
				require.Error(t, err)
				require.Equal(t, tt.wantErrMsg, err.Error())
			} else {
				require.NoError(t, err)
			}

			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
