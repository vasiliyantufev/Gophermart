package server

import (
	"context"
	"database/sql"
	"encoding/json"
	_errors "errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

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

type Ctx string

type server struct {
	log               *logrus.Logger
	cfg               *config.Config
	db                *database.DB
	userRepository    *user.User
	orderRepository   *order.Order
	balanceRepository *balance.Balance
	tokenRepository   *token.Token
	handlers          ServerHandlers
}

func NewServer(logger *logrus.Logger, cfg *config.Config, db *database.DB, userRepository *user.User, orderRepository *order.Order, balanceRepository *balance.Balance, tokenRepository *token.Token) *server {
	return &server{log: logger, cfg: cfg, db: db, userRepository: userRepository, orderRepository: orderRepository,
		balanceRepository: balanceRepository, tokenRepository: tokenRepository}
}

func (s *server) StartServer(r *chi.Mux, cfg *config.Config, log *logrus.Logger) {
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
	req := &model.UserRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		s.log.Error(err)
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	userLogin, err := s.userRepository.FindByLogin(req.Login)
	if err != nil {
		if _errors.Is(err, sql.ErrNoRows) {
			s.log.Error("Invalid username")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !service.CheckPasswordHash(req.Password, userLogin.Password) {
		s.log.Error("Invalid username/password")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := s.tokenRepository.Create(userLogin.ID)
	if err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", token)
	s.log.Info("Successful login")
	w.WriteHeader(http.StatusOK)
}

func (s *server) registerHandler(w http.ResponseWriter, r *http.Request) {
	req := &model.UserRequest{}
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

	userRegistered, err := s.userRepository.FindByLogin(user.Login)
	if err != nil && !_errors.Is(err, sql.ErrNoRows) {
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if userRegistered != nil {
		s.log.Error("Login is already taken")
		w.WriteHeader(http.StatusConflict)
		return
	}

	userID, err := s.userRepository.Create(user)
	if err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := s.tokenRepository.Create(userID)
	if err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", token)
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

	Order, err := strconv.Atoi(strings.TrimSpace(string(resp)))
	if err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if validOrder := service.ValidLuhn(Order); !validOrder {
		s.log.Error("Invalid order number format")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	timeNow := time.Now()
	userID := r.Context().Value(Ctx("UserIDCtx")).(int)

	OrderID := strconv.Itoa(Order)

	order := &model.OrderDB{
		UserID:        userID,
		OrderID:       OrderID,
		CurrentStatus: statuses.New,
		CreatedAt:     timeNow,
		UpdatedAt:     timeNow,
	}

	orderUploaded, err := s.orderRepository.FindByOrderIDAndUserID(order.OrderID, userID)
	if err != nil && !_errors.Is(err, sql.ErrNoRows) {
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if orderUploaded != nil {
		s.log.Error("Order number has already been uploaded by this user")
		w.WriteHeader(http.StatusOK)
		return
	}

	orderUploaded, err = s.orderRepository.FindByOrderID(order.OrderID)
	if err != nil && !_errors.Is(err, sql.ErrNoRows) {
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if orderUploaded != nil {
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
	userID := r.Context().Value(Ctx("UserIDCtx")).(int)
	orderList, err := s.orderRepository.GetOrders(userID)
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

	s.log.Info("Successful request - get orders")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (s *server) getBalanceHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(Ctx("UserIDCtx")).(int)
	balance, err := s.balanceRepository.GetBalance(userID)
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

	s.log.Info("Successful request - get balance")
	w.Header().Set("Content-Type", "application/json")
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

	Order, err := strconv.Atoi(strings.TrimSpace(string(withdraw.Order)))
	if err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if validOrder := service.ValidLuhn(Order); !validOrder {
		s.log.Error("Invalid order number format")
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	userID := r.Context().Value(Ctx("UserIDCtx")).(int)

	err = s.balanceRepository.CheckBalance(userID, withdraw)
	if err != nil {
		if _errors.Is(err, errors.ErrNotFunds) {
			s.log.Info(err)
			w.WriteHeader(http.StatusOK)
			return
		}
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = s.balanceRepository.WithDraw(userID, withdraw)
	if err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.log.Info("Successful request - create withdraw")
	w.WriteHeader(http.StatusOK)
}

func (s *server) getWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(Ctx("UserIDCtx")).(int)
	withDrawals, err := s.balanceRepository.WithDrawals(userID)
	if withDrawals == nil {
		s.log.Error("No write-offs")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(withDrawals)
	if err != nil {
		s.log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	s.log.Info("Successful request - get withdrawals")
	w.Header().Set("Content-Type", "application/json")
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

		ctx := context.WithValue(r.Context(), Ctx("UserIDCtx"), user.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
