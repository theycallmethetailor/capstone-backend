package main

import (
	"github.com/rs/cors"

	databaseConfig "capstone-backend/config"

	controllers "capstone-backend/controllers"

	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"

	types "capstone-backend/models"

	_ "github.com/lib/pq"
)

func main() {
	app := iris.New()

	// Add CORS to application
	app.WrapRouter(cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedHeaders:   []string{"*"},
		AllowedMethods:   []string{"PATCH", "POST", "GET", "DELETE"},
		AllowCredentials: true,
	}).ServeHTTP)

	app.Logger().SetLevel("debug")
	app.Use(recover.New())
	app.Use(logger.New())

	app.OnErrorCode(iris.StatusInternalServerError, func(ctx iris.Context) {
		type Message struct {
			Message string
		}

		var errorMessage Message

		errorMessage.Message = ctx.Values().GetString("message")
		ctx.JSON(errorMessage)
	})

	app.OnErrorCode(iris.StatusNotFound, controllers.NotFound)

	db, _ := databaseConfig.DbStart()

	db.AutoMigrate(&types.NPO{}, &types.Volunteer{}, &types.Event{}, &types.Shift{}, &types.Tag{})

	// NPO Routes:
	app.Get("/api/npos", controllers.GetAllNPOs)
	app.Get("/api/npo/{id:int}", controllers.ShowNPO)
	app.Post("api/npos", controllers.CreateNPO)
	app.Patch("/api/npos/{id:int}", controllers.UpdateNPO)
	app.Delete("api/npos/{id:int}", controllers.DeleteNPO)

	//Volunteers Routes:
	app.Get("/api/volunteers", controllers.GetAllVolunteers)
	app.Get("/api/volunteer/{id:int}", controllers.ShowVolunteer)
	app.Post("/api/volunteers", controllers.CreateVolunteer)
	app.Patch("/api/volunteers/{id:int}", controllers.UpdateVolunteer)
	app.Delete("/api/volunteers/{id:int}", controllers.DeleteVolunteer)

	//Events Routes:
	app.Get("/api/events", controllers.GetAllEvents)
	app.Get("/api/event/{id:int}", controllers.ShowEvent)
	app.Get("/api/open/events", controllers.GetOpenEvents)
	app.Post("/api/events", controllers.CreateEvent)
	app.Patch("/api/events/{id:int}", controllers.UpdateEvent)
	app.Delete("/api/events/delete/{id:int}", controllers.DeleteEvent)

	//Shifts Routes
	app.Patch("/api/shifts/{shiftid:int}", controllers.VolunteerSignup)
	app.Patch("/api/cancel/shifts/{shiftid:int}", controllers.VolunteerCancel)
	app.Get("/api/shifts/volunteers/{id:int}", controllers.GetVolunteerShifts)
	app.Patch("/api/confirm/shifts/{shiftid:int}", controllers.ConfirmShiftWorked)

	//Tags Routes
	app.Get("/api/tags", controllers.GetAllTags)
	app.Get("/api/tags/list", controllers.GetTagList)
	app.Post("/api/tags", controllers.CreateTags)
	app.Delete("/api/tags", controllers.DeleteTags)

	//Reporting Routes
	app.Post("/api/vol/report", controllers.GetVolunteerHours)

	//Login route
	app.Post("/api/login", controllers.Login)

	app.Run(iris.Addr(":8081"), iris.WithoutServerError(iris.ErrServerClosed))
}
