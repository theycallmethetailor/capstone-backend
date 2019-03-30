package controllers

import (
	"fmt"
	"time"

	"github.com/kataras/iris"
	databaseConfig "github.com/theycallmethetailor/capstone-backend/config"
	types "github.com/theycallmethetailor/capstone-backend/models"
)

func GetAllEvents(ctx iris.Context) {

	db, _ := databaseConfig.DbStart()

	defer db.Close()

	var events []types.Event

	db.Preload("tags").Find(&events)

	ctx.JSON(events)
}

func ShowEvent(ctx iris.Context) {

	db, _ := databaseConfig.DbStart()

	defer db.Close()

	var event types.Event

	urlParam, _ := ctx.Params().GetInt("id")

	db.First(&event, urlParam)

	db.Model(&event).Related(&event.Shifts)

	var tags []types.Tag

	db.Table("tags").Joins("inner join event_tags on event_tags.tag_id = tags.id").Joins("inner join events on event_tags.event_id = events.id").Where("events.id = ?", event.ID).Find(&tags)

	type ReturnEvent struct {
		ID              uint
		CreatedAt       time.Time
		UpdatedAt       time.Time
		NPOID           uint
		NPOName         string
		Name            string
		StartTime       int64
		EndTime         int64
		Tags            []types.Tag
		Description     string
		Location        string
		NumOfVolunteers int
		Shifts          []types.Shift
	}

	var npoInfo types.NPO
	db.Select("npo_name").First(&npoInfo, event.NPOID)

	returnEvent := ReturnEvent{
		ID:              event.ID,
		CreatedAt:       event.CreatedAt,
		UpdatedAt:       event.UpdatedAt,
		NPOID:           event.NPOID,
		NPOName:         npoInfo.NPOName,
		Name:            event.Name,
		StartTime:       event.StartTime,
		EndTime:         event.EndTime,
		Description:     event.Description,
		Location:        event.Location,
		NumOfVolunteers: event.NumOfVolunteers,
		Shifts:          event.Shifts,
	}

	returnEvent.Tags = tags

	ctx.JSON(returnEvent)
}

func GetOpenEvents(ctx iris.Context) {
	db, _ := databaseConfig.DbStart()

	defer db.Close()

	var events []types.Event

	//Find events that haven't already started.
	now := time.Now().Unix() * 1000

	db.Table("events").Where("start_time > ?", now).Select("id, created_at, updated_at, deleted_at, npo_id, name, start_time, end_time, description, location, num_of_volunteers").Order("events.start_time asc").Find(&events)

	type ReturnEvent struct {
		ID              uint
		CreatedAt       time.Time
		UpdatedAt       time.Time
		NPOID           uint
		NPOName         string
		Name            string
		StartTime       int64
		EndTime         int64
		Tags            []types.Tag
		Description     string
		Location        string
		NumOfVolunteers int
		Duration        int64
	}

	var openEvents []ReturnEvent

	//look for only events that that still have open shifts to fill
	for _, event := range events {
		fmt.Println("start time %v now %v", event.StartTime, now)
		var filledShifts []types.Shift
		db.Table("shifts").Where("event_id = ?", event.ID).Not("volunteer_id", 0).Find(&filledShifts)

		if len(filledShifts) != event.NumOfVolunteers {
			var npoInfo types.NPO
			db.Select("npo_name").First(&npoInfo, event.NPOID)
			duration := (event.EndTime - event.StartTime) / 6000
			returnEvent := ReturnEvent{
				ID:              event.ID,
				CreatedAt:       event.CreatedAt,
				UpdatedAt:       event.UpdatedAt,
				NPOID:           event.NPOID,
				NPOName:         npoInfo.NPOName,
				Name:            event.Name,
				StartTime:       event.StartTime,
				EndTime:         event.EndTime,
				Description:     event.Description,
				Location:        event.Location,
				NumOfVolunteers: event.NumOfVolunteers,
				Duration:        duration,
			}
			var tags []types.Tag

			db.Table("tags").Joins("inner join event_tags on event_tags.tag_id = tags.id").Joins("inner join events on event_tags.event_id = events.id").Where("events.id = ?", event.ID).Find(&tags)

			returnEvent.Tags = tags

			openEvents = append(openEvents, returnEvent)
		}
	}

	ctx.JSON(openEvents)
}

func CreateEvent(ctx iris.Context) {

	db, _ := databaseConfig.DbStart()

	defer db.Close()

	type RequestEvent struct {
		NPOID           uint
		Name            string
		StartTime       int64
		EndTime         int64
		Tags            []string
		Description     string
		Location        string
		NumOfVolunteers int
		Shifts          []types.Shift
	}

	var requestBody RequestEvent

	ctx.ReadJSON(&requestBody)

	event := types.Event{
		NPOID:           requestBody.NPOID,
		Name:            requestBody.Name,
		StartTime:       requestBody.StartTime,
		EndTime:         requestBody.EndTime,
		Description:     requestBody.Description,
		Location:        requestBody.Location,
		NumOfVolunteers: requestBody.NumOfVolunteers,
	}

	db.NewRecord(event)
	db.Create(&event)

	for i := 0; i < event.NumOfVolunteers; i++ {
		shift := types.Shift{
			EventID:         event.ID,
			ActualStartTime: event.StartTime,
			ActualEndTime:   event.EndTime,
		}
		db.NewRecord(shift)
		db.Create(&shift)
	}

	var eventTags []types.Tag
	for _, newTag := range requestBody.Tags {

		tag := types.Tag{
			TagName: newTag,
		}
		//Only create a new tag if the tag doesn't already exist
		db.FirstOrCreate(&tag, types.Tag{
			TagName: newTag,
		})

		type EventTag struct {
			EventID uint
			TagID   uint
		}
		newTag := EventTag{
			EventID: event.ID,
			TagID:   tag.ID,
		}
		db.NewRecord(newTag)
		db.Create(&newTag)
		eventTags = append(eventTags, tag)
	}

	var newEvent types.Event

	db.First(&newEvent, event.ID)

	db.Model(&newEvent).Related(&newEvent.Shifts)

	type ReturnEvent struct {
		ID              uint
		CreatedAt       time.Time
		UpdatedAt       time.Time
		NPOID           uint
		NPOName         string
		Name            string
		StartTime       int64
		EndTime         int64
		Tags            []types.Tag
		Description     string
		Location        string
		NumOfVolunteers int
		Shifts          []types.Shift
	}

	var npoInfo types.NPO
	db.First(&npoInfo)

	returnEvent := ReturnEvent{
		ID:              newEvent.ID,
		CreatedAt:       newEvent.CreatedAt,
		UpdatedAt:       newEvent.UpdatedAt,
		NPOID:           newEvent.NPOID,
		NPOName:         npoInfo.NPOName,
		Name:            newEvent.Name,
		StartTime:       newEvent.StartTime,
		EndTime:         newEvent.EndTime,
		Tags:            eventTags,
		Description:     newEvent.Description,
		Location:        newEvent.Location,
		NumOfVolunteers: newEvent.NumOfVolunteers,
		Shifts:          newEvent.Shifts,
	}

	ctx.JSON(returnEvent)
}

func UpdateEvent(ctx iris.Context) {

	db, _ := databaseConfig.DbStart()

	defer db.Close()

	var event types.Event

	urlParam, _ := ctx.Params().GetInt("id")

	db.First(&event, urlParam)

	var requestBody types.Event

	ctx.ReadJSON(&requestBody)

	now := time.Now().Unix()

	//if the event hasn't started:
	if now < event.StartTime {

		// if the NPO changes the start time and the event hasn't started, also change the start time for all shifts
		if requestBody.StartTime != event.StartTime {
			fmt.Print("The request start time is not equal to the event start time")
			db.Table("shifts").Where("event_id = ?", event.ID).Updates(map[string]interface{}{"actual_start_time": requestBody.StartTime})
		}

		//if the NPO changes the end time, and the event hasn't started, also change the end time for all shifts
		if requestBody.EndTime != event.EndTime {
			db.Table("shifts").Where("event_id = ?", event.ID).Updates(map[string]interface{}{"actual_end_time": requestBody.EndTime})
		}

		//change updated fields on Event itself including start/end times
		db.Model(&event).Updates(types.Event{
			Name:        requestBody.Name,
			StartTime:   requestBody.StartTime,
			EndTime:     requestBody.EndTime,
			Tags:        requestBody.Tags,
			Description: requestBody.Description,
			Location:    requestBody.Location,
		})

		if event.NumOfVolunteers < requestBody.NumOfVolunteers {

			additionalVolunteers := requestBody.NumOfVolunteers - event.NumOfVolunteers
			//add new shifts
			for i := 0; i < additionalVolunteers; i++ {
				shift := types.Shift{
					EventID:         event.ID,
					ActualStartTime: event.StartTime,
					ActualEndTime:   event.EndTime,
				}
				db.NewRecord(shift)
				db.Create(&shift)
			}

			//update NumOfVolunteers
			db.Model(&event).Updates(types.Event{
				NumOfVolunteers: requestBody.NumOfVolunteers,
			})
		}

		var newEvent types.Event

		db.First(&newEvent, event.ID)

		db.Model(&event).Related(&newEvent.Shifts)

		ctx.JSON(newEvent)

	} else if event.StartTime == requestBody.StartTime && event.EndTime == requestBody.EndTime {
		//change updated fields on Event itself including start/end times
		db.Model(&event).Updates(types.Event{
			Name:        requestBody.Name,
			StartTime:   requestBody.StartTime,
			EndTime:     requestBody.EndTime,
			Tags:        requestBody.Tags,
			Description: requestBody.Description,
			Location:    requestBody.Location,
		})

		if event.NumOfVolunteers < requestBody.NumOfVolunteers {

			additionalVolunteers := requestBody.NumOfVolunteers - event.NumOfVolunteers
			//add new shifts
			for i := 0; i < additionalVolunteers; i++ {
				shift := types.Shift{
					EventID:         event.ID,
					ActualStartTime: event.StartTime,
					ActualEndTime:   event.EndTime,
				}
				db.NewRecord(shift)
				db.Create(&shift)
			}

			//update NumOfVolunteers
			db.Model(&event).Updates(types.Event{
				NumOfVolunteers: requestBody.NumOfVolunteers,
			})
		}

		var newEvent types.Event

		db.First(&newEvent, event.ID)

		db.Model(&event).Related(&newEvent.Shifts)

		ctx.JSON(newEvent)
	} else {
		ctx.Values().Set("message", "Unable to alter start or end times once an event has already started.")
		ctx.StatusCode(500)
	}

}
