package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Rizal-Nurochman/matchnbuild/middlewares"
	"github.com/Rizal-Nurochman/matchnbuild/modules/auth"
	"github.com/Rizal-Nurochman/matchnbuild/modules/chat"
	chatWs "github.com/Rizal-Nurochman/matchnbuild/modules/chat/websocket"
	"github.com/Rizal-Nurochman/matchnbuild/modules/design_item"
	"github.com/Rizal-Nurochman/matchnbuild/modules/designer"
	"github.com/Rizal-Nurochman/matchnbuild/modules/payment"
	"github.com/Rizal-Nurochman/matchnbuild/modules/project_request"
	"github.com/Rizal-Nurochman/matchnbuild/modules/quotation"
	"github.com/Rizal-Nurochman/matchnbuild/modules/recommendation"
	"github.com/Rizal-Nurochman/matchnbuild/modules/upload"
	"github.com/Rizal-Nurochman/matchnbuild/modules/user"
	"github.com/Rizal-Nurochman/matchnbuild/modules/user_preferences"
	"github.com/Rizal-Nurochman/matchnbuild/pkg/constants"
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

func run(server *gin.Engine, injector *do.Injector) {
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

	httpServer := &http.Server{
		Addr:    serve,
		Handler: server,
	}

	// Run the HTTP server in the background so we can listen for OS signals and
	// shut everything down gracefully.
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("error running server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server...")

	// Stop the websocket hub first so all client connections are closed and
	// presence/last_seen is flushed before the process exits.
	if hub, err := do.InvokeNamed[*chatWs.Hub](injector, constants.ChatHub); err == nil {
		hub.Shutdown()
		log.Println("[WS] Hub stopped")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("forced shutdown: %v", err)
	}

	log.Println("server exited")
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

	run(server, injector)
}
