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
	GetRoomById(roomID int) (models.Room, error)
	GetUserByID(userID int) (models.User, error)
	UpdateUser(user models.User) error
	Authenticate(email, password string) (int, int, error)
	AllReservations() ([]models.Reservation, error)
	NewReservations() ([]models.Reservation, error)
	GetReservationById(id int) (models.Reservation, error)
	UpdateReservation(reservation models.Reservation) error
	UpdateReservationProcessed(reservationID, processed int) error
	DeleteReservation(reservationID int) error
	AllRooms() ([]models.Room, error)
	GetRoomRestrictionsByDate(roomID int, startDate, endDate time.Time) ([]models.RoomRestriction, error)
	DeleteRoomRestriction(roomRestrictionID int) error
}
