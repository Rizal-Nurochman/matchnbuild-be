package providers

import (
	"github.com/Rizal-Nurochman/matchnbuild/config"
	authController "github.com/Rizal-Nurochman/matchnbuild/modules/auth/controller"
	authRepo "github.com/Rizal-Nurochman/matchnbuild/modules/auth/repository"
	authService "github.com/Rizal-Nurochman/matchnbuild/modules/auth/service"
	designItemHandler "github.com/Rizal-Nurochman/matchnbuild/modules/design_item/controller"
	desigItemRepo "github.com/Rizal-Nurochman/matchnbuild/modules/design_item/repository"
	desigItemService "github.com/Rizal-Nurochman/matchnbuild/modules/design_item/service"
	designerController "github.com/Rizal-Nurochman/matchnbuild/modules/designer/controller"
	designerRepo "github.com/Rizal-Nurochman/matchnbuild/modules/designer/repository"
	designerService "github.com/Rizal-Nurochman/matchnbuild/modules/designer/service"
	prController "github.com/Rizal-Nurochman/matchnbuild/modules/project_request/controller"
	prRepo "github.com/Rizal-Nurochman/matchnbuild/modules/project_request/repository"
	prService "github.com/Rizal-Nurochman/matchnbuild/modules/project_request/service"
	qController "github.com/Rizal-Nurochman/matchnbuild/modules/quotation/controller"
	qRepo "github.com/Rizal-Nurochman/matchnbuild/modules/quotation/repository"
	qService "github.com/Rizal-Nurochman/matchnbuild/modules/quotation/service"
	userController "github.com/Rizal-Nurochman/matchnbuild/modules/user/controller"
	"github.com/Rizal-Nurochman/matchnbuild/modules/user/repository"
	userService "github.com/Rizal-Nurochman/matchnbuild/modules/user/service"
	userPreferenceHandler "github.com/Rizal-Nurochman/matchnbuild/modules/user_preferences/controller"
	userPreferenceRepo "github.com/Rizal-Nurochman/matchnbuild/modules/user_preferences/repository"
	userPreferenceService "github.com/Rizal-Nurochman/matchnbuild/modules/user_preferences/service"
	recController "github.com/Rizal-Nurochman/matchnbuild/modules/recommendation/controller"
	recService "github.com/Rizal-Nurochman/matchnbuild/modules/recommendation/service"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/constants"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func InitDatabase(injector *do.Injector) {
	do.ProvideNamed(injector, constants.DB, func(i *do.Injector) (*gorm.DB, error) {
		return config.SetUpDatabaseConnection(), nil
	})
}

func RegisterDependencies(injector *do.Injector) {
	InitDatabase(injector)

	do.ProvideNamed(injector, constants.JWTService, func(i *do.Injector) (authService.JWTService, error) {
		return authService.NewJWTService(), nil
	})

	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	jwtService := do.MustInvokeNamed[authService.JWTService](injector, constants.JWTService)

	// Repositories
	userRepository := repository.NewUserRepository(db)
	refreshTokenRepository := authRepo.NewRefreshTokenRepository(db)
	projectRequestRepository := prRepo.NewProjectRequestRepository(db)
	conversationRepository := prRepo.NewConversationRepository(db)
	designerProfileRepository := prRepo.NewDesignerProfileRepository(db)
	quotationRepository := qRepo.NewQuotationRepository(db)
	orderRepository := qRepo.NewOrderRepository(db)
	designerRepository := designerRepo.NewDesignerProfileRepository(db)
	designItemRepository := desigItemRepo.NewDesignItemRepository(db)
	userPreferencesRepository := userPreferenceRepo.NewUserPreferenceRepository(db)

	// Services
	userSvc := userService.NewUserService(userRepository, db)
	authSvc := authService.NewAuthService(userRepository, refreshTokenRepository, jwtService, db)
	projectRequestSvc := prService.NewProjectRequestService(projectRequestRepository, conversationRepository, designerProfileRepository, db)
	quotationSvc := qService.NewQuotationService(quotationRepository, orderRepository, projectRequestRepository, designerProfileRepository, db)
	designerSvc := designerService.NewDesignerProfileService(designerRepository, db)
	desigItemSvc := desigItemService.NewDesignItemService(designItemRepository, designerRepository, userPreferencesRepository, db)
	userPrencesSvc := userPreferenceService.NewUserPreferenceService(userPreferencesRepository)
	recommendationSvc := recService.NewRecommendationService(designItemRepository)

	// Controllers
	do.Provide(
		injector, func(i *do.Injector) (userController.UserController, error) {
			return userController.NewUserController(i, userSvc), nil
		},
	)

	do.Provide(
		injector, func(i *do.Injector) (authController.AuthController, error) {
			return authController.NewAuthController(i, authSvc), nil
		},
	)

	do.Provide(
		injector, func(i *do.Injector) (prController.ProjectRequestController, error) {
			return prController.NewProjectRequestController(projectRequestSvc), nil
		},
	)

	do.Provide(
		injector, func(i *do.Injector) (qController.QuotationController, error) {
			return qController.NewQuotationController(quotationSvc), nil
		},
	)

	do.Provide(
		injector, func(i *do.Injector) (designerController.DesignerProfileHandler, error) {
			return designerController.NewDesignerProfileHandler(designerSvc), nil
		},
	)

	do.Provide(
		injector, func(i *do.Injector) (designItemHandler.DesignItemHandler, error) {
			return designItemHandler.NewDesignItemHandler(i, desigItemSvc), nil
		},
	)

	do.Provide(
		injector, func(i *do.Injector) (userPreferenceHandler.UserPreferenceHandler, error) {
			return userPreferenceHandler.NewUserPreferenceHandler(i, userPrencesSvc), nil
		},
	)

	do.Provide(
		injector, func(i *do.Injector) (recController.RecommendationHandler, error) {
			return recController.NewRecommendationHandler(recommendationSvc), nil
		},
	)	
}
