package api

import (
	"context"
	"encoding/json"
	"fmt"
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
}

func NewServer(logger *logrus.Logger, cfg *config.Config, db *database.DB, storeSession *sessions.CookieStore) *server {
	return &server{log: *logger, cfg: cfg, db: db, storeSession: storeSession}
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

func (s *server) loginHandler(w http.ResponseWriter, r *http.Request) { //curl -X POST http://localhost:8080/api/user/login -H "Content-Type: application/json" -d '{"login": "login", "password": "password"}'

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

	session, err := s.storeSession.Get(r, sessionName)
	if err != nil {
		s.log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	session.Values["user_id"] = u.ID
	if err := s.storeSession.Save(r, w, session); err != nil {
		s.log.Error(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	OrderID, err := strconv.Atoi(strings.TrimSpace(string(resp)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	s.log.Info(order)

	o, _ := s.orderRepository.FindByID(order.ID, s.db)
	if o == nil {
		s.log.Error("Order already created")
		return
	}
	err = s.orderRepository.Create(order, s.db)
	if err != nil {
		s.log.Error(err)
		return
	}
}

func (s *server) getOrderHandler(w http.ResponseWriter, r *http.Request) {

}

func (s *server) getOrdersHandler(w http.ResponseWriter, r *http.Request) {

	user := r.Context().Value("ctxUser").(*model.User)

	o, _ := s.orderRepository.GetOrders(user.ID, s.db)

	fmt.Print(o)

}

func (s *server) getBalanceHandler(w http.ResponseWriter, r *http.Request) {

}

func (s *server) postWithdrawHandler(w http.ResponseWriter, r *http.Request) {

}

func (s *server) getWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {

}

func (s *server) authMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		session, err := s.storeSession.Get(r, sessionName)
		if err != nil {
			s.log.Error(err)
			return
		}

		id, auth := session.Values["user_id"]
		if !auth {
			s.log.Error("Unauthorized")
			return
		}
		s.log.Info("User authorized")

		u, _ := s.userRepository.FindByID(id)
		if u == nil {
			s.log.Error("User not find")
			return
		}

		ctx := context.WithValue(r.Context(), "ctxUser", u)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
