package query

import (
	"github.com/Caknoooo/go-pagination"
	"gorm.io/gorm"
)

type Designer struct {
	ID              string  `json:"id"`
	UserID          string  `json:"user_id"`
	Name            string  `json:"name"`
	ProfilePicture  *string `json:"profile_picture"`
	Bio             string  `json:"bio"`
	ExperienceYears int     `json:"experience_years"`
	IsVerified      bool    `json:"is_verified"`
	IsAvailable     bool    `json:"is_available"`
	Location        string  `json:"location"`
}

type DesignerFilter struct {
	pagination.BaseFilter
}

func (f *DesignerFilter) ApplyFilters(query *gorm.DB) *gorm.DB {
	query = query.Joins("JOIN users ON users.id = designer_profiles.user_id").Select("designer_profiles.*, users.name, users.profile_picture")
	return query
}

func (f *DesignerFilter) GetTableName() string {
	return "designer_profiles"
}

func (f *DesignerFilter) GetSearchFields() []string {
	return []string{"name"}
}

func (f *DesignerFilter) GetDefaultSort() string {
	return "id asc"
}

func (f *DesignerFilter) GetIncludes() []string {
	return f.Includes
}

func (f *DesignerFilter) GetPagination() pagination.PaginationRequest {
	return f.Pagination
}

func (f *DesignerFilter) Validate() {
	var validIncludes []string
	allowedIncludes := f.GetAllowedIncludes()
	for _, include := range f.Includes {
		if allowedIncludes[include] {
			validIncludes = append(validIncludes, include)
		}
	}
	f.Includes = validIncludes
}

func (f *DesignerFilter) GetAllowedIncludes() map[string]bool {
	return map[string]bool{}
}
