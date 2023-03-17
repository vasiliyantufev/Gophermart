package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"github.com/vasiliyantufev/gophermart/internal/config"
	database "github.com/vasiliyantufev/gophermart/internal/db"
	"github.com/vasiliyantufev/gophermart/internal/model"
	"github.com/vasiliyantufev/gophermart/internal/service"
	"github.com/vasiliyantufev/gophermart/internal/storage/errors"
	"github.com/vasiliyantufev/gophermart/internal/storage/repositories/balance"
	"github.com/vasiliyantufev/gophermart/internal/storage/repositories/order"
	"github.com/vasiliyantufev/gophermart/internal/storage/repositories/token"
	"github.com/vasiliyantufev/gophermart/internal/storage/repositories/user"
	"github.com/vasiliyantufev/gophermart/internal/storage/statuses"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ServerHandlers interface {
	loginHandler(w http.ResponseWriter, r *http.Request)
	registerHandler(w http.ResponseWriter, r *http.Request)
	createOrderHandler(w http.ResponseWriter, r *http.Request)
	getOrdersHandler(w http.ResponseWriter, r *http.Request)
	getBalanceHandler(w http.ResponseWriter, r *http.Request)
	createWithdrawHandler(w http.ResponseWriter, r *http.Request)
	getWithdrawalsHandler(w http.ResponseWriter, r *http.Request)
	authMiddleware(w http.ResponseWriter, r *http.Request)
}

type server struct {
	log               logrus.Logger
	cfg               *config.Config
	db                *database.DB
	userRepository    *user.User
	orderRepository   *order.Order
	balanceRepository *balance.Balance
	tokenRepository   *token.Token
	handlers          ServerHandlers
}

func NewServer(logger *logrus.Logger, cfg *config.Config, db *database.DB /*storeSession *sessions.CookieStore,, jwt service.JWT*/) *server {
	return &server{log: *logger, cfg: cfg, db: db}
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
			r.Post("/orders", s.createOrderHandler)
			r.Get("/orders", s.getOrdersHandler)
			r.Get("/balance", s.getBalanceHandler)
			r.Post("/balance/withdraw", s.createWithdrawHandler)
			r.Get("/withdrawals", s.getWithdrawalsHandler)
		})
	})
	return r
}

func (s *server) loginHandler(w http.ResponseWriter, r *http.Request) {

	req := &model.Request{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		s.log.Error(err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	u, err := s.userRepository.FindByLogin(req.Login)
	if err != nil {
		if err == sql.ErrNoRows {
			s.log.Error("Invalid username")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !service.CheckPasswordHash(req.Password, u.Password) {
		s.log.Error("Invalid username/password")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := s.tokenRepository.Create(u)
	if err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", token)
	w.WriteHeader(http.StatusOK)
}

func (s *server) registerHandler(w http.ResponseWriter, r *http.Request) {

	req := &model.Request{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hashedPassword, err := service.HashPassword(req.Password)
	if err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user := &model.User{
		Login:    req.Login,
		Password: hashedPassword,
	}

	u, err := s.userRepository.FindByLogin(user.Login)
	if u != nil {
		s.log.Error("Login is already taken")
		w.WriteHeader(http.StatusConflict)
		return
	}

	err = s.userRepository.Create(user)
	if err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.log.Info("Successful registration")
	w.WriteHeader(http.StatusOK)
}

func (s *server) createOrderHandler(w http.ResponseWriter, r *http.Request) {

	resp, err := io.ReadAll(r.Body)
	if err != nil {
		s.log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	OrderID, err := strconv.Atoi(strings.TrimSpace(string(resp)))
	if err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if validOrder := service.ValidLuhn(OrderID); validOrder == false {
		s.log.Error("Invalid order number format")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	timeNow := time.Now()
	user := r.Context().Value("UserCtx").(*model.TokenUser)

	order := &model.Order{
		UserID:        user.UserID,
		OrderID:       OrderID,
		CurrentStatus: string(statuses.New),
		CreatedAt:     timeNow,
		UpdatedAt:     timeNow,
	}

	o, err := s.orderRepository.FindByOrderIDAndUserID(order.OrderID, user.UserID)
	if o != nil {
		s.log.Error("Order number has already been uploaded by this user")
		w.WriteHeader(http.StatusOK)
		return
	}
	o, _ = s.orderRepository.FindByOrderID(order.OrderID)
	if o != nil {
		s.log.Error("Order number has already been uploaded by another user")
		w.WriteHeader(http.StatusConflict)
		return
	}
	err = s.orderRepository.Create(order)
	if err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.log.Info("New order number accepted for processing")
	w.WriteHeader(http.StatusAccepted)
}

func (s *server) getOrdersHandler(w http.ResponseWriter, r *http.Request) {

	user := r.Context().Value("UserCtx").(*model.TokenUser)

	orderList, err := s.orderRepository.GetOrders(user.UserID)
	if err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if orderList == nil {
		s.log.Error("No data to answer")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	resp, err := json.Marshal(orderList)
	if err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.log.Info("Successful request processing")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (s *server) getBalanceHandler(w http.ResponseWriter, r *http.Request) {

	user := r.Context().Value("UserCtx").(*model.TokenUser)

	balance, err := s.balanceRepository.GetBalance(user.UserID)
	if err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(balance)
	if err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.log.Info("Successful request processing")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (s *server) createWithdrawHandler(w http.ResponseWriter, r *http.Request) {

	withdraw := &model.BalanceWithdraw{}
	if err := json.NewDecoder(r.Body).Decode(withdraw); err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	o, _ := s.orderRepository.FindByOrderID(withdraw.Order)
	if o == nil {
		s.log.Error("Invalid order number")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if withdraw.Sum > 0 {
		withdraw.Sum = -withdraw.Sum
	}

	user := r.Context().Value("UserCtx").(*model.TokenUser)

	err := s.balanceRepository.CheckBalance(user.UserID, withdraw)
	if err != nil {
		if err == errors.ErrNotFunds {
			s.log.Info("There are not enough funds on the account")
			w.WriteHeader(http.StatusOK)
			return
		}
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = s.balanceRepository.WithDraw(user.UserID, withdraw)
	if err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.log.Info("Successful request processing")
	w.WriteHeader(http.StatusOK)
}

func (s *server) getWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {

	user := r.Context().Value("UserCtx").(*model.TokenUser)

	withDrawals, err := s.balanceRepository.WithDrawals(user.UserID)
	if withDrawals == nil {
		s.log.Error(err)
		http.Error(w, "No write-offs", http.StatusNoContent)
	}
	if err != nil {
		s.log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	resp, err := json.Marshal(withDrawals)
	if err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.log.Info("Successful request processing")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (s *server) authMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		token := r.Header.Get("Authorization")
		if token == "" {
			s.log.Error("Token missing")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		valid, user, err := s.tokenRepository.Validate(token)
		if err != nil {
			s.log.Error(err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if !valid {
			s.log.Error("Not validate token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		s.log.Info("User authenticated")

		ctx := context.WithValue(r.Context(), "UserCtx", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
