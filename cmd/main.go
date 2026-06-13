package main

import (
	"log"
	"os"

	"github.com/Rizal-Nurochman/matchnbuild/middlewares"
	"github.com/Rizal-Nurochman/matchnbuild/modules/auth"
	"github.com/Rizal-Nurochman/matchnbuild/modules/chat"
	"github.com/Rizal-Nurochman/matchnbuild/modules/design_item"
	"github.com/Rizal-Nurochman/matchnbuild/modules/designer"
	"github.com/Rizal-Nurochman/matchnbuild/modules/payment"
	"github.com/Rizal-Nurochman/matchnbuild/modules/project_request"
	"github.com/Rizal-Nurochman/matchnbuild/modules/quotation"
	"github.com/Rizal-Nurochman/matchnbuild/modules/recommendation"
	"github.com/Rizal-Nurochman/matchnbuild/modules/upload"
	"github.com/Rizal-Nurochman/matchnbuild/modules/user"
	"github.com/Rizal-Nurochman/matchnbuild/modules/user_preferences"
	"github.com/Rizal-Nurochman/matchnbuild/providers"
	"github.com/Rizal-Nurochman/matchnbuild/script"
	"github.com/samber/do"

	"github.com/common-nighthawk/go-figure"
	"github.com/gin-gonic/gin"
)

func args(injector *do.Injector) bool {
	if len(os.Args) > 1 {
		flag := script.Commands(injector)
		return flag
	}

	return true
}

func run(server *gin.Engine) {
	server.Static("/assets", "./assets")

	port := os.Getenv("GOLANG_PORT")
	if port == "" {
		port = "8888"
	}

	var serve string
	if os.Getenv("APP_ENV") == "localhost" {
		serve = "0.0.0.0:" + port
	} else {
		serve = ":" + port
	}

	myFigure := figure.NewColorFigure("Caknoo", "", "green", true)
	myFigure.Print()

	if err := server.Run(serve); err != nil {
		log.Fatalf("error running server: %v", err)
	}
}

func main() {
	var (
		injector = do.New()
	)

	providers.RegisterDependencies(injector)

	if !args(injector) {
		return
	}

	server := gin.Default()
	server.Use(middlewares.CORSMiddleware())

	v1 := server.Group("/api/v1")
	{
		// Register module routes
		user.RegisterRoutes(v1, injector)
		auth.RegisterRoutes(v1, injector)
		chat.RegisterRoutes(v1, injector)
		upload.RegisterRoutes(v1, injector)
		project_request.RegisterRoutes(v1, injector)
		quotation.RegisterRoutes(v1, injector)
		payment.RegisterRoutes(v1, injector)
		designer.RegisterRoutes(v1, injector)
		design_item.RegisterRoutes(v1, injector)
		user_preferences.RegisterRoutes(v1, injector)
		recommendation.RegisterRoutes(v1, injector)
	}

	run(server)
}
