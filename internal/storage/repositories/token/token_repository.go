package token

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	database "github.com/vasiliyantufev/gophermart/internal/db"
	"github.com/vasiliyantufev/gophermart/internal/model"
)

const lengthToken = 32
const lifetimeToken = 100 * time.Hour

type TokenRepository interface {
	Create(user *model.User) (string, error)
	Generate(length int) string
	Validate(token string) bool
}

type Token struct {
	db *database.DB
}

func New(db *database.DB) *Token {
	return &Token{
		db: db,
	}
}

func (t Token) Create(userID int) (string, error) {

	token := t.Generate(lengthToken)
	currentTime := time.Now()

	var id int
	return token, t.db.Pool.QueryRow(
		"INSERT INTO token (user_id, token, created_at, deleted_at) VALUES ($1, $2, $3, $4) RETURNING id",
		userID,
		token,
		currentTime,
		currentTime.Add(time.Hour+lifetimeToken),
	).Scan(&id)
}

func (t Token) Generate(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func (t *Token) Validate(token string) (bool, *model.Token, error) {

	currentTime := time.Now()

	tokenUser := &model.Token{}
	if err := t.db.Pool.QueryRow("SELECT users.id, users.login, token.token, token.deleted_at FROM token "+
		"INNER JOIN users ON users.id = token.user_id  where token.token = $1", token).Scan(
		&tokenUser.UserID,
		&tokenUser.Login,
		&tokenUser.Token,
		&tokenUser.DeletedAt,
	); err != nil {
		return false, nil, err
	}

	if currentTime.After(tokenUser.DeletedAt) {
		return false, tokenUser, nil
	}

	return true, tokenUser, nil
}
