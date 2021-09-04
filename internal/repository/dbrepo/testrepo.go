package dbrepo

import (
	"errors"
	"github.com/wfabjanczuk/sawler_bookings/internal/models"
	"time"
)

func (m *testDBRepo) AllUsers() bool {
	return true
}

func (m *testDBRepo) InsertReservation(res models.Reservation) (int, error) {
	return 1, nil
}

func (m *testDBRepo) InsertRoomRestriction(roomRes models.RoomRestriction) (int, error) {
	return 1, nil
}

func (m *testDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	return true, nil
}

func (m *testDBRepo) SearchAvailabilityByDates(start, end time.Time) ([]models.Room, error) {
	return []models.Room{}, nil
}

func (m *testDBRepo) GetRoomById(roomID int) (models.Room, error) {
	if roomID > 2 {
		return models.Room{}, errors.New("room not found")
	}

	return models.Room{}, nil
}
