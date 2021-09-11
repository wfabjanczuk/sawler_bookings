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
	if res.RoomID == 2 {
		return 0, errors.New("room not found")
	}

	return 1, nil
}

func (m *testDBRepo) InsertRoomRestriction(roomRes models.RoomRestriction) (int, error) {
	if roomRes.RoomID == 1000 {
		return 0, errors.New("room not found")
	}

	return 1, nil
}

func (m *testDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	if roomID > 2 {
		return false, errors.New("room not found")
	}

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

func (m *testDBRepo) GetUserByID(userID int) (models.User, error) {
	return models.User{}, nil
}

func (m *testDBRepo) UpdateUser(user models.User) error {
	return nil
}

func (m *testDBRepo) Authenticate(email, password string) (int, string, error) {
	return 0, "", nil
}

func (m *testDBRepo) AllReservations() ([]models.Reservation, error) {
	return []models.Reservation{}, nil
}

func (m *testDBRepo) NewReservations() ([]models.Reservation, error) {
	return []models.Reservation{}, nil
}

func (m *testDBRepo) GetReservationById(id int) (models.Reservation, error) {
	return models.Reservation{}, nil
}

func (m *testDBRepo) UpdateReservation(reservation models.Reservation) error {
	return nil
}

func (m *testDBRepo) DeleteReservation(reservationID int) error {
	return nil
}

func (m *testDBRepo) UpdateReservationProcessed(reservationID, processed int) error {
	return nil
}

func (m *testDBRepo) AllRooms() ([]models.Room, error) {
	return []models.Room{}, nil
}

func (m *testDBRepo) GetRoomRestrictionsByDate(roomID int, startDate, endDate time.Time) ([]models.RoomRestriction, error) {
	return []models.RoomRestriction{}, nil
}

func (m *testDBRepo) DeleteRoomRestriction(roomRestrictionID int) error {
	return nil
}
