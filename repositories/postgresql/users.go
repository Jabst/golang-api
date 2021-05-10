package postgresql

import (
	"code/tech-test/domain/users/models"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	pgerr "github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
)

var (
	ErrWrongVersion  = errors.New("wrong version")
	ErrDuplicateUser = errors.New("duplicate nickname")
	ErrUserNotFound  = errors.New("user not found")
)

type UserStore struct {
	pool *sql.DB
}

func NewUserStore(pool *sql.DB) *UserStore {
	return &UserStore{pool}
}

func (s UserStore) Get(ctx context.Context, id int) (models.User, error) {

	row := s.pool.QueryRowContext(ctx, `
		SELECT id, first_name, last_name, nickname, password, email, country, disabled, version, created_at, updated_at
		FROM users
		WHERE id = $1 AND disabled = 'f' 
	`, id)

	return s.scan(row)
}

func queryComposer(terms map[string]string) (string, []interface{}) {
	if len(terms) == 0 {
		return "", nil
	}

	var expressions []string = make([]string, 0)
	var index int = 1
	filterParams := make([]interface{}, 0)

	for key, value := range terms {
		expressions = append(expressions, fmt.Sprintf(" %s = $%d ", key, index))
		filterParams = append(filterParams, value)
		index++
	}

	return strings.Join(expressions, "AND") + " AND", filterParams
}

func (s UserStore) List(ctx context.Context, queryTerm map[string]string) ([]models.User, error) {

	var users []models.User = make([]models.User, 0)

	filterArguments, filterParams := queryComposer(queryTerm)

	rows, err := s.pool.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, first_name, last_name, nickname, password, email, country, disabled, version, created_at, updated_at
		FROM users
		WHERE %s disabled = 'f' 
	`, filterArguments), filterParams...)
	if err != nil {
		return nil, fmt.Errorf("%w failed to query context", err)
	}

	defer rows.Close()

	users, err = s.scanMultipleRows(rows)
	if err != nil {
		return nil, fmt.Errorf("%w error scan multiple rows", err)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w rows returned error", err)
	}

	return users, nil

}

func (s UserStore) Store(ctx context.Context, user models.User, version uint32) (models.User, error) {
	var result models.User

	tx, err := s.pool.Begin()
	if err != nil {
		return models.User{}, fmt.Errorf("%w failed to begin transaction", err)
	}

	current, err := s.lockForUpdate(ctx, tx, user.ID)
	if err != nil {
		tx.Rollback()
		return models.User{}, err
	}

	if current != version {
		tx.Rollback()
		return models.User{}, ErrWrongVersion
	}

	if current == 0 {
		result, err = s.create(ctx, tx, user)
	} else {
		result, err = s.update(ctx, tx, user, version)
	}
	if err != nil {
		tx.Rollback()
		return models.User{}, err
	}

	tx.Commit()

	if err != nil {
		return models.User{}, err
	}

	return result, nil
}

func (s UserStore) lockForUpdate(ctx context.Context, tx *sql.Tx, id int) (uint32, error) {
	var version uint32

	row := tx.QueryRowContext(ctx, `
		SELECT version
		FROM users
		WHERE id = $1 FOR UPDATE NOWAIT
	`, id)

	err := row.Scan(&version)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	return version, nil
}

func (s UserStore) Delete(ctx context.Context, id int) (models.User, error) {
	row := s.pool.QueryRowContext(ctx, `
		UPDATE users
		SET disabled = 't', updated_at = NOW()
		WHERE id = $1
		RETURNING id, first_name, last_name, nickname, password, email, country, disabled, version, created_at, updated_at
	`, id)

	return s.scan(row)
}

func (s UserStore) create(ctx context.Context, tx *sql.Tx, user models.User) (models.User, error) {

	row := tx.QueryRowContext(ctx, `
		INSERT INTO users(first_name, last_name, nickname, password, email, country)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, first_name, last_name, nickname, password, email, country, disabled, version, created_at, updated_at
	`,
		user.FirstName,
		user.LastName,
		user.Nickname,
		user.Password,
		user.Email,
		user.Country,
	)
	return s.scan(row)
}

func (s UserStore) update(ctx context.Context, tx *sql.Tx, user models.User, version uint32) (models.User, error) {

	row := tx.QueryRowContext(ctx, `
		UPDATE users
		SET first_name = $1, last_name = $2, nickname = $3, password = $4,
		email = $5, country = $6, version = $7, updated_at = NOW()
		WHERE id = $8 AND version = $9
		RETURNING id, first_name, last_name, nickname, password, email, country, disabled, version, created_at, updated_at
	`,
		user.FirstName,
		user.LastName,
		user.Nickname,
		user.Password,
		user.Email,
		user.Country,
		user.Meta.GetVersion()+1,
		user.ID,
		version,
	)
	return s.scan(row)
}

func (s UserStore) scan(row *sql.Row) (models.User, error) {
	var (
		id        int
		firstname string
		lastname  string
		nickname  string
		password  string
		email     string
		country   string
		disabled  bool
		version   uint32
		createdAt time.Time
		updatedAt time.Time
	)

	if err := row.Scan(
		&id,
		&firstname,
		&lastname,
		&nickname,
		&password,
		&email,
		&country, &disabled, &version, &createdAt, &updatedAt); err != nil {
		if pgErr, ok := err.(pgx.PgError); ok {
			if pgErr.Code == pgerr.UniqueViolation {
				return models.User{}, ErrDuplicateUser
			}
		}

		if err == sql.ErrNoRows {
			return models.User{}, ErrUserNotFound
		}

		return models.User{}, err
	}

	return s.hydrateUser(id, firstname, lastname, nickname, password, email, country, disabled, version, createdAt, updatedAt), nil
}

func (s UserStore) scanMultipleRows(rows *sql.Rows) ([]models.User, error) {
	var (
		users []models.User = make([]models.User, 0)
	)

	type User struct {
		id        int
		firstname string
		lastname  string
		nickname  string
		password  string
		email     string
		country   string
		disabled  bool
		version   uint32
		createdAt time.Time
		updatedAt time.Time
	}

	for rows.Next() {
		var scannedUser User
		if err := rows.Scan(
			&scannedUser.id,
			&scannedUser.firstname,
			&scannedUser.lastname,
			&scannedUser.nickname,
			&scannedUser.password,
			&scannedUser.email,
			&scannedUser.country, &scannedUser.disabled, &scannedUser.version, &scannedUser.createdAt, &scannedUser.updatedAt); err != nil {
			if pgErr, ok := err.(pgx.PgError); ok {
				if pgErr.Code == pgerr.UniqueViolation {
					return nil, ErrDuplicateUser
				}
			}

			if err == sql.ErrNoRows {
				return nil, ErrUserNotFound
			}

			return nil, err
		}

		user := s.hydrateUser(scannedUser.id, scannedUser.firstname, scannedUser.lastname,
			scannedUser.nickname, scannedUser.password, scannedUser.email, scannedUser.country,
			scannedUser.disabled, scannedUser.version, scannedUser.createdAt, scannedUser.updatedAt)

		users = append(users, user)
	}

	return users, nil
}

func (s UserStore) hydrateUser(id int, fn, ln, nickname, password, email, country string, disabled bool, version uint32, createdAt, updatedAt time.Time) models.User {
	user := models.NewUser(id, fn, ln, nickname, password, email, country)

	user.Meta.HydrateMeta(version, createdAt, updatedAt, disabled)

	return user
}
