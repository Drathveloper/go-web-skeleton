package rdbms_test

import (
	"context"
	"database/sql/driver"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Drathveloper/go-web-skeleton/common/database/rdbms"
)

// item is a throwaway model: Paginate is a scope, and a scope can only be
// observed through the statement it ends up producing.
type item struct {
	ID uint
}

func (item) TableName() string {
	return "items"
}

// sqlmock.Sqlmock is an interface by design: it is the handle the library hands
// out to set expectations on.
//
//nolint:ireturn
func setupMockDB(t *testing.T) (sqlmock.Sqlmock, *gorm.DB) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	require.NoError(t, err)

	return mock, gormDB
}

// mockPostgresClient adapts *gorm.DB to rdbms.PostgresClient, the one-method
// interface every repository depends on so that GORM can be swapped for
// go-sqlmock here.
type mockPostgresClient struct {
	db *gorm.DB
}

func (m *mockPostgresClient) WithContext(ctx context.Context) *gorm.DB {
	return m.db.WithContext(ctx)
}

// TestPaginate asserts the SQL that reaches the driver, not just that Paginate
// returns something. The clamping is only observable there: page and pageSize
// come straight from the query string, so a page of -3 or a size of 10000 is
// what an ordinary crawler sends.
//
// GORM emits LIMIT and OFFSET as bound arguments, and omits OFFSET entirely
// when it is zero — hence the two query shapes below.
func TestPaginate(t *testing.T) {
	t.Parallel()

	const (
		queryWithOffset = `^SELECT \* FROM "items" LIMIT \$1 OFFSET \$2$`
		queryNoOffset   = `^SELECT \* FROM "items" LIMIT \$1$`
	)

	tests := []struct {
		name           string
		page           int
		pageSize       int
		expectedLimit  int
		expectedOffset int
	}{
		{
			name:           "test paginate should offset by one full page on page two",
			page:           2,
			pageSize:       10,
			expectedLimit:  10,
			expectedOffset: 10,
		},
		{
			name:           "test paginate should not offset on the first page",
			page:           1,
			pageSize:       10,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name:           "test paginate should clamp page zero to the first page",
			page:           0,
			pageSize:       10,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name:           "test paginate should clamp a negative page to the first page",
			page:           -3,
			pageSize:       20,
			expectedLimit:  20,
			expectedOffset: 0,
		},
		{
			name:           "test paginate should clamp a page size above the maximum",
			page:           1,
			pageSize:       10000,
			expectedLimit:  100,
			expectedOffset: 0,
		},
		{
			name:           "test paginate should offset by the clamped page size, not the requested one",
			page:           3,
			pageSize:       10000,
			expectedLimit:  100,
			expectedOffset: 200,
		},
		{
			name:           "test paginate should keep a page size exactly at the maximum",
			page:           1,
			pageSize:       100,
			expectedLimit:  100,
			expectedOffset: 0,
		},
		{
			name:           "test paginate should fall back to the default page size when it is zero",
			page:           1,
			pageSize:       0,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name:           "test paginate should fall back to the default page size when it is negative",
			page:           4,
			pageSize:       -5,
			expectedLimit:  10,
			expectedOffset: 30,
		},
		{
			name:           "test paginate should keep a page size exactly at the minimum",
			page:           5,
			pageSize:       1,
			expectedLimit:  1,
			expectedOffset: 4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock, gormDB := setupMockDB(t)
			client := &mockPostgresClient{db: gormDB}

			query := queryNoOffset
			args := []driver.Value{tt.expectedLimit}
			if tt.expectedOffset != 0 {
				query = queryWithOffset
				args = append(args, tt.expectedOffset)
			}
			mock.ExpectQuery(query).
				WithArgs(args...).
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

			var items []item
			err := client.WithContext(t.Context()).
				Scopes(rdbms.Paginate(tt.page, tt.pageSize)).
				Find(&items).Error

			require.NoError(t, err)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

// The scope has to compose: it is always applied on top of the filters the
// caller already put on the query, never instead of them.
func TestPaginate_ComposesWithTheRestOfTheQuery(t *testing.T) {
	t.Parallel()

	mock, gormDB := setupMockDB(t)
	client := &mockPostgresClient{db: gormDB}

	mock.ExpectQuery(`^SELECT \* FROM "items" WHERE id > \$1 ORDER BY id LIMIT \$2 OFFSET \$3$`).
		WithArgs(5, 10, 20).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(6))

	var items []item
	err := client.WithContext(t.Context()).
		Where("id > ?", 5).
		Order("id").
		Scopes(rdbms.Paginate(3, 10)).
		Find(&items).Error

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
