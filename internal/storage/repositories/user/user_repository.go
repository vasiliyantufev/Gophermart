package user

import (
	"time"

	database "github.com/vasiliyantufev/gophermart/internal/db"
	"github.com/vasiliyantufev/gophermart/internal/model"
)

type Constructor interface {
	Create(user *model.User) error
	FindByID(id int) (*model.User, error)
	FindByLogin(login string) (*model.User, error)
}

type User struct {
	db *database.DB
}

func New(db *database.DB) *User {
	return &User{
		db: db,
	}
}

func (u *User) Create(user *model.User) (int, error) {
	var id int
	if err := u.db.Pool.QueryRow(
		"INSERT INTO users (login, password, created_at) VALUES ($1, $2, $3) RETURNING id",
		user.Login,
		user.Password,
		time.Now(),
	).Scan(&id); err != nil {
		return id, err
	}

	return id, nil
}

func (u *User) FindByID(id int) (*model.User, error) {

	user := &model.User{}

	if err := u.db.Pool.QueryRow("SELECT * FROM users where id=$1", id).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.CreatedAt,
	); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *User) FindByLogin(login string) (*model.User, error) {

	user := &model.User{}

	if err := u.db.Pool.QueryRow("SELECT * FROM users where login=$1", login).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.CreatedAt,
	); err != nil {
		return nil, err
	}

	return user, nil
}
