package repository

import "github.com/wfabjanczuk/sawler_bookings/internal/models"

type DatabaseRepo interface {
	AllUsers() bool
	InsertReservation(res models.Reservation) (int, error)
	InsertRoomRestriction(roomRes models.RoomRestriction) (int, error)
}
