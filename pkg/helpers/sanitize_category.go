package helpers

import "github.com/Rizal-Nurochman/matchnbuild/modules/design_item/dto"

func SanitizeByCategory(req *dto.DesignItemCreateRequest) {
	if req.Category == dto.CategoryArchitecture {
		req.RoomArea = nil
		req.RoomType = nil
	}
	if req.Category == dto.CategoryInterior {
		req.LandAreaMin = nil
		req.LandAreaMax = nil
		req.BuildingArea = nil
		req.NumFloors = nil
		req.NumBedrooms = nil
	}
}