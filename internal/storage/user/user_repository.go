package user

import (
	database "github.com/vasiliyantufev/gophermart/internal/db"
	"github.com/vasiliyantufev/gophermart/internal/model"
	"time"
)

type UserRepository interface {
	Create(*model.User, database.DB) error
	FindByID(int, database.DB) (*model.User, error)
	FindByLogin(string, database.DB) (*model.User, error)
}

type User struct {
	userRepository *UserRepository
}

/*, db database.DB*/
func (r *User) Create(user *model.User, db *database.DB) error {

	timeNow := time.Now()

	return db.Pool.QueryRow(
		"INSERT INTO users (login, password, created_at, updated_at) VALUES ($1, $2, $3, $4) RETURNING id",
		user.Login,
		user.Password,
		timeNow,
		timeNow,
	).Scan(&user.ID)
}

func (r *User) FindByID(id int, db database.DB) (*model.User, error) {

	user := &model.User{}

	if err := db.Pool.QueryRow("SELECT * FROM users where id=$1", id).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return user, nil
}

func (r *User) FindByLogin(login string, db *database.DB) (*model.User, error) {

	user := &model.User{}

	if err := db.Pool.QueryRow("SELECT * FROM users where login=$1", login).Scan(
		&user.ID,
		&user.Login,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, err
	}

	return user, nil
}
