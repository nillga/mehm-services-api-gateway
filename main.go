package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/nillga/mehm-services-api-gateway/controller"
	_ "github.com/nillga/mehm-services-api-gateway/docs"
	router "github.com/nillga/mehm-services-api-gateway/http"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

var apiController = controller.NewApiGatewayController()
var apiRouter = router.NewApiGatewayRouter()

// @title           Swagger Example API
// @version         1.0
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:420/api
// @BasePath  /

func main() {
	cr := chi.NewRouter()

	cr.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:1323/swagger/doc.json"), //The url pointing to API definition
	))

	apiRouter.GET("/api/mehms", apiController.GetAllMehms)
	apiRouter.GET("/api/mehms/{id}", apiController.GetSpecificMehm)
	apiRouter.POST("/api/mehms/{id}/like", apiController.LikeMehm)
	apiRouter.POST("/api/mehms/{id}/remove", apiController.DeleteMehm)
	apiRouter.GET("/api/user", apiController.ResolveProfile)
	apiRouter.POST("/api/user/delete", apiController.DeleteUser)
	apiRouter.GET("/api/comments/{id}", apiController.GetComment)
	apiRouter.POST("/api/comments/new", apiController.PostComment)
	apiRouter.POST("/api/comments/update", apiController.EditComment)
	apiRouter.POST("/api/mehms/{id}/update", apiController.EditMehm)

	c := cors.New(cors.Options{
		AllowedHeaders: []string{"Authorization", "Credentials", "Cookie"},
	})
	l := log.Logger{}
	l.SetOutput(os.Stdout)
	c.Log = &l

	go func() {
		log.Fatalln(http.ListenAndServe(os.Getenv("SWAG"), c.Handler(cr)))
	}()
	apiRouter.SERVE(os.Getenv("PORT"))
}
