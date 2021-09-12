package dbrepo

import (
	"context"
	"database/sql"
	"errors"
	"github.com/wfabjanczuk/sawler_bookings/internal/models"
	"golang.org/x/crypto/bcrypt"
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

	stmt := `insert into "reservation" 
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
	var nullableReservationId sql.NullInt32

	if roomRes.ReservationID > 0 {
		nullableReservationId = sql.NullInt32{
			Int32: int32(roomRes.ReservationID),
			Valid: true,
		}
	}

	stmt := `insert into "room_restriction" 
(room_id, restriction_id, reservation_id, start_date, end_date, created_at, updated_at) 
values ($1, $2, $3, $4, $5, $6, $7) returning id`

	err := m.DB.QueryRowContext(ctx, stmt,
		roomRes.RoomID,
		roomRes.RestrictionID,
		nullableReservationId,
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

func (m *postgresDBRepo) DeleteRoomRestriction(roomRestrictionID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), maxQueryTime)
	defer cancel()

	stmt := `delete from "room_restriction" where id = $1`

	_, err := m.DB.ExecContext(ctx, stmt, roomRestrictionID)

	return err
}

func (m *postgresDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), maxQueryTime)
	defer cancel()

	var numRows int

	query := `select count(id) from "room_restriction" where $1 <= end_date and start_date <= $2 and room_id = $3`
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

	query := `select r.id, r.room_name from "room" r where r.id not in 
	(select room_id from "room_restriction" rr where $1 <= rr.end_date and rr.start_date <= $2)`

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

	query := `select r.id, r.room_name, r.created_at, r.updated_at from "room" r where r.id = $1`

	err := m.DB.QueryRowContext(ctx, query, roomID).Scan(&room.ID, &room.RoomName, &room.CreatedAt, &room.CreatedAt)

	if err != nil {
		return room, err
	}

	return room, nil
}

func (m *postgresDBRepo) GetUserByID(userID int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), maxQueryTime)
	defer cancel()

	var user models.User

	query := `select u.id, u.first_name, u.last_name, u.email, u.password, u.access_level, u.created_at, u.updated_at
       from "user" u where u.id = $1`

	err := m.DB.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Password, &user.AccessLevel, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return user, err
	}

	return user, nil
}

func (m *postgresDBRepo) UpdateUser(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), maxQueryTime)
	defer cancel()

	statement := `update "user" set first_name = $1, last_name = $2, email = $3, access_level = $4, updated_at = $5
	where id = $6`

	_, err := m.DB.ExecContext(ctx, statement,
		user.FirstName, user.LastName, user.Email, user.AccessLevel, time.Now(), user.ID,
	)

	return err
}

func (m *postgresDBRepo) Authenticate(email, password string) (int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), maxQueryTime)
	defer cancel()

	var id, accessLevel int
	var passwordHash string

	query := `select u.id, u.access_level, u.password from "user" u where u.email = $1`

	err := m.DB.QueryRowContext(ctx, query, email).Scan(&id, &accessLevel, &passwordHash)
	if err != nil {
		return 0, 0, errors.New("incorrect email")
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, 0, errors.New("incorrect password")
	} else if err != nil {
		return 0, 0, err
	}

	return id, accessLevel, nil
}

func (m *postgresDBRepo) AllReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), maxQueryTime)
	defer cancel()

	var reservations []models.Reservation

	query := `select 
	r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, r.end_date, r.room_id, r.created_at, r.updated_at, r.processed,
	rm.id, rm.room_name
	from "reservation" r left join "room" rm on r.room_id = rm.id
	order by r.start_date asc`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}

	for rows.Next() {
		var r models.Reservation
		var rm models.Room

		err := rows.Scan(
			&r.ID,
			&r.FirstName,
			&r.LastName,
			&r.Email,
			&r.Phone,
			&r.StartDate,
			&r.EndDate,
			&r.RoomID,
			&r.CreatedAt,
			&r.UpdatedAt,
			&r.Processed,
			&rm.ID,
			&rm.RoomName,
		)

		if err != nil {
			return reservations, err
		}

		r.Room = rm
		reservations = append(reservations, r)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

func (m *postgresDBRepo) NewReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), maxQueryTime)
	defer cancel()

	var reservations []models.Reservation

	query := `select 
	r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, r.end_date, r.room_id, r.created_at, r.updated_at, r.processed,
	rm.id, rm.room_name
	from "reservation" r left join "room" rm on r.room_id = rm.id where r.processed = 0
	order by r.start_date asc`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}

	for rows.Next() {
		var r models.Reservation
		var rm models.Room

		err := rows.Scan(
			&r.ID,
			&r.FirstName,
			&r.LastName,
			&r.Email,
			&r.Phone,
			&r.StartDate,
			&r.EndDate,
			&r.RoomID,
			&r.CreatedAt,
			&r.UpdatedAt,
			&r.Processed,
			&rm.ID,
			&rm.RoomName,
		)

		if err != nil {
			return reservations, err
		}

		r.Room = rm
		reservations = append(reservations, r)
	}

	if err = rows.Err(); err != nil {
		return reservations, err
	}

	return reservations, nil
}

func (m *postgresDBRepo) GetReservationById(id int) (models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), maxQueryTime)
	defer cancel()

	query := `select 
	r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, r.end_date, r.room_id, r.created_at, r.updated_at, r.processed,
	rm.id, rm.room_name
	from "reservation" r left join "room" rm on r.room_id = rm.id where r.id = $1
	order by r.start_date asc`

	var r models.Reservation
	var rm models.Room

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&r.ID,
		&r.FirstName,
		&r.LastName,
		&r.Email,
		&r.Phone,
		&r.StartDate,
		&r.EndDate,
		&r.RoomID,
		&r.CreatedAt,
		&r.UpdatedAt,
		&r.Processed,
		&rm.ID,
		&rm.RoomName,
	)
	r.Room = rm

	return r, err
}

func (m *postgresDBRepo) UpdateReservation(reservation models.Reservation) error {
	ctx, cancel := context.WithTimeout(context.Background(), maxQueryTime)
	defer cancel()

	statement := `update "reservation"
	set first_name = $1, last_name = $2, email = $3, phone = $4, updated_at = $5 where id = $6`

	_, err := m.DB.ExecContext(ctx, statement,
		reservation.FirstName, reservation.LastName, reservation.Email, reservation.Phone, time.Now(), reservation.ID,
	)

	return err
}

func (m *postgresDBRepo) UpdateReservationProcessed(reservationID, processed int) error {
	ctx, cancel := context.WithTimeout(context.Background(), maxQueryTime)
	defer cancel()

	statement := `update "reservation" set processed = $1 where id = $2`

	_, err := m.DB.ExecContext(ctx, statement, processed, reservationID)

	return err
}

func (m *postgresDBRepo) DeleteReservation(reservationID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), maxQueryTime)
	defer cancel()

	statement := `delete from "reservation" where id = $1`

	_, err := m.DB.ExecContext(ctx, statement, reservationID)

	return err
}

func (m *postgresDBRepo) AllRooms() ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), maxQueryTime)
	defer cancel()

	var rooms []models.Room

	query := `select rm.id, rm.room_name, rm.created_at, rm.updated_at from "room" rm order by rm.room_name asc`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return rooms, err
	}

	for rows.Next() {
		var rm models.Room

		err := rows.Scan(
			&rm.ID,
			&rm.RoomName,
			&rm.CreatedAt,
			&rm.UpdatedAt,
		)

		if err != nil {
			return rooms, err
		}

		rooms = append(rooms, rm)
	}

	if err = rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}

func (m *postgresDBRepo) GetRoomRestrictionsByDate(roomID int, startDate, endDate time.Time) ([]models.RoomRestriction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), maxQueryTime)
	defer cancel()

	var roomRestrictions []models.RoomRestriction

	query := `select id, coalesce(reservation_id, 0), restriction_id, room_id, start_date, end_date, created_at, updated_at
	from "room_restriction" where $1 <= start_date and end_date <= $2 and room_id = $3 order by start_date asc`

	rows, err := m.DB.QueryContext(ctx, query, startDate, endDate, roomID)
	if err != nil {
		return roomRestrictions, err
	}

	for rows.Next() {
		var rr models.RoomRestriction

		err := rows.Scan(
			&rr.ID,
			&rr.ReservationID,
			&rr.RestrictionID,
			&rr.RoomID,
			&rr.StartDate,
			&rr.EndDate,
			&rr.CreatedAt,
			&rr.UpdatedAt,
		)

		if err != nil {
			return roomRestrictions, err
		}

		roomRestrictions = append(roomRestrictions, rr)
	}

	if err = rows.Err(); err != nil {
		return roomRestrictions, err
	}

	return roomRestrictions, nil
}
