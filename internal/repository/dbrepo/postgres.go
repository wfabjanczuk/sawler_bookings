package dbrepo

import (
	"context"
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

	stmt := `insert into public.reservation 
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

	stmt := `insert into public.room_restriction 
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

	query := `select count(id) from public.room_restriction where $1 <= end_date and start_date <= $2 and room_id = $3`
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

	query := `select r.id, r.room_name from public.room r where r.id not in 
	(select room_id from public.room_restriction rr where $1 <= rr.end_date and rr.start_date <= $2)`

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

	query := `select r.id, r.room_name, r.created_at, r.updated_at from public.room r where r.id = $1`

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
       from public.user u where u.id = $1`

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

	statement := `update public.user set first_name = $1, last_name = $2, email = $3, access_level = $4, updated_at = $5
	where u.id = $6`

	_, err := m.DB.ExecContext(ctx, statement,
		user.FirstName, user.LastName, user.Email, user.AccessLevel, time.Now(), user.ID,
	)

	return err
}

func (m *postgresDBRepo) Authenticate(email, password string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), maxQueryTime)
	defer cancel()

	var id int
	var passwordHash string

	query := `select u.id, u.password from public.user u where u.email = $1`

	err := m.DB.QueryRowContext(ctx, query, email).Scan(&id, &passwordHash)
	if err != nil {
		return 0, "", errors.New("incorrect email")
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", errors.New("incorrect password")
	} else if err != nil {
		return 0, "", err
	}

	return id, passwordHash, nil
}

func (m *postgresDBRepo) AllReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), maxQueryTime)
	defer cancel()

	var reservations []models.Reservation

	query := `select 
	r.id, r.first_name, r.last_name, r.email, r.phone, r.start_date, r.end_date, r.room_id, r.created_at, r.updated_at, r.processed,
	rm.id, rm.room_name
	from reservation r left join room rm on r.room_id = rm.id
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
	from reservation r left join room rm on r.room_id = rm.id where r.processed = 0
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
