package mapper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/common/http/dto"
	"github.com/Drathveloper/go-web-skeleton/common/http/mapper"
)

func TestMapDTOPaginationToDomainPagination(t *testing.T) {
	t.Parallel()

	requestPagination := &dto.Pagination{
		Page: 1,
		Size: 2,
	}
	expected := &domain.Pagination{
		Page: 1,
		Size: 2,
	}

	actual := mapper.MapDTOPaginationToDomainPagination(requestPagination)

	require.Equal(t, expected, actual)
}
