package mapper

import (
	"github.com/Drathveloper/go-web-skeleton/common/domain"
	"github.com/Drathveloper/go-web-skeleton/common/http/dto"
)

func MapDTOPaginationToDomainPagination(pagination *dto.Pagination) *domain.Pagination {
	return &domain.Pagination{
		Page: pagination.Page,
		Size: pagination.Size,
	}
}
