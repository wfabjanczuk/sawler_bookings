package repository

import (
	"github.com/wfabjanczuk/sawler_bookings/internal/models"
	"time"
)

type DatabaseRepo interface {
	AllUsers() bool
	InsertReservation(res models.Reservation) (int, error)
	InsertRoomRestriction(roomRes models.RoomRestriction) (int, error)
	SearchAvailabilityByDatesByRoomID(start, end time.Time, roomId int) (bool, error)
	SearchAvailabilityByDates(start, end time.Time) ([]models.Room, error)
}
