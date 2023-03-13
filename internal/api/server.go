package api

import (
	"context"
	"encoding/json"
	"github.com/vasiliyantufev/gophermart/internal/storage"
	"github.com/vasiliyantufev/gophermart/internal/storage/token"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"

	"github.com/vasiliyantufev/gophermart/internal/config"
	database "github.com/vasiliyantufev/gophermart/internal/db"
	"github.com/vasiliyantufev/gophermart/internal/model"
	"github.com/vasiliyantufev/gophermart/internal/service"
	"github.com/vasiliyantufev/gophermart/internal/storage/balance"
	"github.com/vasiliyantufev/gophermart/internal/storage/order"
	"github.com/vasiliyantufev/gophermart/internal/storage/user"
)

type ServerHandlers interface {
	loginHandler(w http.ResponseWriter, r *http.Request)
	registerHandler(w http.ResponseWriter, r *http.Request)
	postOrderHandler(w http.ResponseWriter, r *http.Request)
	getOrderHandler(w http.ResponseWriter, r *http.Request)
	getOrdersHandler(w http.ResponseWriter, r *http.Request)
	getBalanceHandler(w http.ResponseWriter, r *http.Request)
	postWithdrawHandler(w http.ResponseWriter, r *http.Request)
	getWithdrawalsHandler(w http.ResponseWriter, r *http.Request)
	authMiddleware(w http.ResponseWriter, r *http.Request)
}

const sessionName = "gophermart"

type server struct {
	log               logrus.Logger
	cfg               *config.Config
	db                *database.DB
	userRepository    *user.User
	orderRepository   *order.Order
	balanceRepository *balance.Balance
	tokenRepository   *token.Token
	handlers          ServerHandlers
	storeSession      sessions.Store
	jwt               service.JWT
}

func NewServer(logger *logrus.Logger, cfg *config.Config, db *database.DB, storeSession *sessions.CookieStore, jwt service.JWT) *server {
	return &server{log: *logger, cfg: cfg, db: db, storeSession: storeSession, jwt: jwt}
}

func (s *server) StartServer(r *chi.Mux, cfg *config.Config, log *logrus.Logger) {

	s.userRepository = user.New(s.db)
	s.orderRepository = order.New(s.db)
	s.balanceRepository = balance.New(s.db)
	s.tokenRepository = token.New(s.db)

	log.Infof("Starting application %v\n", cfg.Address)
	if con := http.ListenAndServe(cfg.Address, r); con != nil {
		log.Fatal(con)
	}
}

func (s *server) Route() *chi.Mux {

	r := chi.NewRouter()

	r.Route("/api/user", func(r chi.Router) {
		r.Post("/register", s.registerHandler)
		r.Post("/login", s.loginHandler)
		r.Group(func(r chi.Router) {
			r.Use(s.authMiddleware)
			r.Post("/orders", s.postOrderHandler)
			r.Get("/orders", s.getOrdersHandler)
			r.Get("/orders/{id}", s.getOrderHandler)
			r.Get("/balance", s.getBalanceHandler)
			r.Post("/balance/withdraw", s.getBalanceHandler)
			r.Get("/withdrawals", s.postWithdrawHandler)
		})
	})
	return r
}

func (s *server) loginHandler(w http.ResponseWriter, r *http.Request) {

	type request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	req := &request{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		s.log.Error(err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	user := &model.User{
		Login:    req.Login,
		Password: req.Password,
	}

	u, err := s.userRepository.FindByLogin(user.Login)
	if err != nil {
		s.log.Error("Invalid username")
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if service.CheckPasswordHash(req.Password, u.Password) {
		s.log.Error("Invalid username/password")
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	userPayload := &model.UserPayload{
		ID:    u.ID,
		Login: u.Login,
	}

	token, err := s.jwt.GenerateToken(userPayload)
	if err != nil {
		s.log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	r.Header.Set("Authorization", token)

	s.log.Info("Successful login")
	http.Error(w, "Successful login", http.StatusOK)

}

func (s *server) registerHandler(w http.ResponseWriter, r *http.Request) {

	type request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	req := &request{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		s.log.Error(err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	password, err := service.HashPassword(req.Password)
	if err != nil {
		s.log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	user := &model.User{
		Login:    req.Login,
		Password: password,
	}

	u, err := s.userRepository.FindByLogin(user.Login)
	if u != nil {
		s.log.Error("Login is already taken")
		http.Error(w, "Login is already taken", http.StatusConflict)
		return
	}

	err = s.userRepository.Create(user)
	if err != nil {
		s.log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.log.Info("Successful registration")
	http.Error(w, "Successful registration", http.StatusOK)
}

func (s *server) postOrderHandler(w http.ResponseWriter, r *http.Request) {

	resp, err := io.ReadAll(r.Body)
	if err != nil {
		s.log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	OrderID, err := strconv.Atoi(strings.TrimSpace(string(resp)))
	if err != nil {
		s.log.Error(err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if validOrder := service.ValidLuhn(OrderID); validOrder == false {
		s.log.Error("Invalid order number format")
		http.Error(w, "invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	timeNow := time.Now()
	user := r.Context().Value("ctxUser").(*model.User)

	order := &model.Order{
		UserID:        user.ID,
		OrderID:       OrderID,
		CurrentStatus: storage.Statuses(0),
		CreatedAt:     timeNow,
		UpdatedAt:     timeNow,
	}

	o, _ := s.orderRepository.FindByLoginAndID(order.ID, user)
	if o == nil {
		s.log.Error("Order number has already been uploaded by this user")
		http.Error(w, "Order number has already been uploaded by this user", http.StatusOK)
		return
	}
	o, _ = s.orderRepository.FindByID(order.ID)
	if o == nil {
		s.log.Error("Order number has already been uploaded by another user")
		http.Error(w, "Order number has already been uploaded by another user", http.StatusConflict)
		return
	}
	err = s.orderRepository.Create(order)
	if err != nil {
		s.log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.log.Info("New order number accepted for processing")
	http.Error(w, "New order number accepted for processing", http.StatusAccepted)
}

func (s *server) getOrderHandler(w http.ResponseWriter, r *http.Request) {

}

func (s *server) getOrdersHandler(w http.ResponseWriter, r *http.Request) {

	user := r.Context().Value("ctxUser").(*model.User)

	o, _ := s.orderRepository.GetOrders(user.ID)
	if o == nil {
		s.log.Error("No data to answer")
		http.Error(w, "No data to answer", http.StatusNoContent)
	}

	s.log.Error("Successful request processing")
	http.Error(w, "Successful request processing", http.StatusOK)

}

func (s *server) getBalanceHandler(w http.ResponseWriter, r *http.Request) {

}

func (s *server) postWithdrawHandler(w http.ResponseWriter, r *http.Request) {

}

func (s *server) getWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {

}

func (s *server) authMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "User not authenticated", http.StatusUnauthorized)
			s.log.Error("User not authenticated")
			return
		}
		s.log.Info("User authenticated")

		user, err := s.jwt.ParseToken(token)
		if err != nil {
			s.log.Error(err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), "ctxUser", user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
