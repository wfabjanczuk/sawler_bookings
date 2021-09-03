package dbrepo

import (
	"context"
	"github.com/wfabjanczuk/sawler_bookings/internal/models"
	"time"
)

const maxQueryTime = 3 * time.Second

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

func (m *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), maxQueryTime)
	defer cancel()

	var newID int

	stmt := `insert into reservation 
(first_name, last_name, email, phone, start_date, end_date, room_id, created_at, updated_at) 
values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`

	err := m.DB.QueryRowContext(ctx, stmt,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

func (m *postgresDBRepo) InsertRoomRestriction(roomRes models.RoomRestriction) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), maxQueryTime)
	defer cancel()

	var newID int

	stmt := `insert into room_restriction 
(room_id, restriction_id, reservation_id, start_date, end_date, created_at, updated_at) 
values ($1, $2, $3, $4, $5, $6, $7) returning id`

	err := m.DB.QueryRowContext(ctx, stmt,
		roomRes.RoomID,
		roomRes.RestrictionID,
		roomRes.ReservationID,
		roomRes.StartDate,
		roomRes.EndDate,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

func (m *postgresDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), maxQueryTime)
	defer cancel()

	var numRows int

	query := `select count(id) from room_restrictions where $1 <= end_date and start_date <= $2 and room_id = $3`
	err := m.DB.QueryRowContext(ctx, query, start, end, roomID).Scan(&numRows)

	if err != nil {
		return false, err
	}

	return numRows == 0, nil
}

func (m *postgresDBRepo) SearchAvailabilityByDates(start, end time.Time) ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), maxQueryTime)
	defer cancel()

	var rooms []models.Room

	query := `select r.id, r.room_name from room r where r.id not in 
	(select room_id from room_restriction rr where $1 <= rr.end_date and rr.start_date <= $2)`

	rows, err := m.DB.QueryContext(ctx, query, start, end)

	if err != nil {
		return rooms, err
	}

	for rows.Next() {
		var room models.Room

		err = rows.Scan(&room.ID, &room.RoomName)

		if err != nil {
			return rooms, err
		}

		rooms = append(rooms, room)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}

func (m *postgresDBRepo) GetRoomById(roomID int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), maxQueryTime)
	defer cancel()

	var room models.Room

	query := `select r.id, r.room_name, r.created_at, r.updated_at from room r where r.id = $1`

	err := m.DB.QueryRowContext(ctx, query, roomID).Scan(&room.ID, &room.RoomName, &room.CreatedAt, &room.CreatedAt)

	if err != nil {
		return room, err
	}

	return room, nil
}
