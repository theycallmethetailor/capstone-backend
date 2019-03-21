package controllers

import (
	"time"

	"github.com/kataras/iris"
	databaseConfig "github.com/theycallmethetailor/capstone-backend/config"
	types "github.com/theycallmethetailor/capstone-backend/models"
)

func GetAllNPOs(ctx iris.Context) {
	// Create connection to database
	db, _ := databaseConfig.DbStart()
	// Close connection when function is done running
	defer db.Close()

	// Create container for all users
	var npos []types.NPO

	// Query database for all users and populate container with said results
	db.Find(&npos)

	// Respond to request with JSON of all the users
	ctx.JSON(npos)
}

func GetVolunteerHours(ctx iris.Context) {
	db, _ := databaseConfig.DbStart()

	defer db.Close()

	type Shift struct {
		ShiftID         uint
		EventID         uint
		EventName       string
		WasWorked       bool
		ActualStartTime int64
		ActualEndTime   int64
	}

	type VolunteerHours struct {
		VolunteerID uint
		Username    string
		FirstName   string
		LastName    string
		Shifts      []Shift
	}

	//get the startDate in Epoch time from the query params
	startDate, err := ctx.URLParamInt("startDate")

	if err != nil {
		ctx.Values().Set("message", "Please provide a valid start date for your request.")
		ctx.StatusCode(500)
	}
	//get the startDate in Epoch time from the query params
	endDate, err := ctx.URLParamInt("endDate")

	if err != nil {
		ctx.Values().Set("message", "Please provide a valid end date for your request.")
		ctx.StatusCode(500)
	}
	//get NPOID from query params
	npoID, err := ctx.URLParamInt("npoid")
	if err != nil {
		ctx.Values().Set("message", "Please provide a valid NPO ID for your request.")
		ctx.StatusCode(500)
	}

	now := time.Now().Unix()

	if startDate > int(now) {
		ctx.Values().Set("message", "Unable to provide report for future date span.")
		ctx.StatusCode(500)
	}

	var events []types.Event

	db.Table("events").Where("npoid = ? AND start_time >= ? AND start_time <= ?", npoID, startDate, endDate).Find(&events)

	ctx.JSON(events)

}

func ShowNPO(ctx iris.Context) {

	// Create connection to database
	db, _ := databaseConfig.DbStart()
	// Close connection when function is done running
	defer db.Close()

	// Create container for one user
	var npo types.NPO

	// Acquire the id via the url params
	urlParam, _ := ctx.Params().GetInt("id")

	// Query database for user with a certain ID
	db.First(&npo, urlParam)

	db.Model(&npo).Related(&npo.Events)

	ctx.JSON(npo)
}

func CreateNPO(ctx iris.Context) {
	db, _ := databaseConfig.DbStart()

	defer db.Close()

	var requestBody types.NPO

	ctx.ReadJSON(&requestBody)

	npo := types.NPO{
		NPOName:     requestBody.NPOName,
		Description: requestBody.Description,
		Website:     requestBody.Website,
		Email:       requestBody.Email,
		FirstName:   requestBody.FirstName,
		LastName:    requestBody.LastName,
		Password:    requestBody.Password,
	}

	db.NewRecord(npo)
	db.Create(&npo)

	if db.NewRecord(npo) == false {
		ctx.JSON(npo)
	} else {
	}

}

func UpdateNPO(ctx iris.Context) {

	db, _ := databaseConfig.DbStart()

	defer db.Close()

	var npo types.NPO

	urlParam, _ := ctx.Params().GetInt("id")

	db.First(&npo, urlParam)

	var requestBody types.NPO

	ctx.ReadJSON(&requestBody)

	// Update multiple attributes with `struct`, will only update those changed & non blank fields
	db.Model(&npo).Updates(types.NPO{
		NPOName:     requestBody.NPOName,
		Description: requestBody.Description,
		Website:     requestBody.Website,
		Email:       requestBody.Email,
		FirstName:   requestBody.FirstName,
		LastName:    requestBody.LastName,
		Password:    requestBody.Password,
	})

	var updatedNPO types.NPO

	db.First(&updatedNPO, urlParam)

	ctx.JSON(updatedNPO)

}

func DeleteNPO(ctx iris.Context) {

	//connect to database at start, close db at end of func
	db, _ := databaseConfig.DbStart()
	defer db.Close()

	var npo types.NPO
	urlParam, _ := ctx.Params().GetInt("id")

	db.First(&npo, urlParam)

	db.Unscoped().Delete(&npo)
	ctx.JSON(npo)

}
