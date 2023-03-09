package api

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"github.com/vasiliyantufev/gophermart/internal/config"
	database "github.com/vasiliyantufev/gophermart/internal/db"
	"github.com/vasiliyantufev/gophermart/internal/model"
	"github.com/vasiliyantufev/gophermart/internal/storage/order"
	"github.com/vasiliyantufev/gophermart/internal/storage/user"
	"io"
	"net/http"
	"strconv"
	"time"
)

type ServerHandlers interface {
	loginHandler(w http.ResponseWriter, r *http.Request)
	registerHandler(w http.ResponseWriter, r *http.Request)
	postOrdersHandler(w http.ResponseWriter, r *http.Request)
	getOrdersHandler(w http.ResponseWriter, r *http.Request)
	getBalanceHandler(w http.ResponseWriter, r *http.Request)
	postWithdrawHandler(w http.ResponseWriter, r *http.Request)
	getWithdrawalsHandler(w http.ResponseWriter, r *http.Request)
	authMiddleware(w http.ResponseWriter, r *http.Request)
}

const sessionName = "gophermart"

type server struct {
	log             logrus.Logger
	cfg             *config.Config
	db              *database.DB
	userRepository  user.User
	orderRepository order.Order
	handlers        ServerHandlers
	storeSession    sessions.Store
}

func NewServer(logger *logrus.Logger, cfg *config.Config, db *database.DB, storeSession *sessions.CookieStore) *server {
	return &server{log: *logger, cfg: cfg, db: db, storeSession: storeSession}
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
			r.Post("/orders", s.postOrdersHandler)
			r.Get("/orders", s.getOrdersHandler)
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
		return
	}

	user := &model.User{
		Login:    req.Login,
		Password: req.Password,
	}

	u, _ := s.userRepository.FindByLogin(user.Login, s.db)
	if u == nil {
		s.log.Error("User no find")
		return
	}

	session, err := s.storeSession.Get(r, sessionName)
	if err != nil {
		s.log.Error(err)
		return
	}

	session.Values["user_id"] = u.ID
	if err := s.storeSession.Save(r, w, session); err != nil {
		s.log.Error(err)
		return
	}

	p, _ := s.storeSession.Get(r, sessionName)

	s.log.Error(p.Values)
}

func (s *server) registerHandler(w http.ResponseWriter, r *http.Request) {

	type request struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	req := &request{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		s.log.Error(err)
		return
	}

	user := &model.User{
		Login:    req.Login,
		Password: req.Password,
	}

	u, _ := s.userRepository.FindByLogin(user.Login, s.db)
	if u == nil {
		err := s.userRepository.Create(user, s.db)
		if err != nil {
			s.log.Error(err)
			return
		}
	}

}

func (s *server) postOrdersHandler(w http.ResponseWriter, r *http.Request) {

	resp, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	OrderNumber, _ := strconv.Atoi(string(resp))
	timeNow := time.Now()

	order := &model.Order{
		UserID:      1,
		OrderNumber: OrderNumber,
		Status:      "STATUS",
		Accrual:     123,
		UpdatedAt:   timeNow,
		CreatedAt:   timeNow,
		UploadedAt:  timeNow,
	}

	o, _ := s.orderRepository.FindByID(order.ID, s.db)
	if o == nil {
		err := s.orderRepository.Create(order, s.db)
		if err != nil {
			s.log.Error(err)
			return
		}
	}
}

func (s *server) getOrdersHandler(w http.ResponseWriter, r *http.Request) {

	resp, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	UserID, _ := strconv.Atoi(string(resp))

	o, _ := s.orderRepository.GetOrders(UserID, s.db)

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

		s.log.Error(session.Values)

		_, ok := session.Values["user_id"]
		if !ok {
			s.log.Error("Unauthorized")
			return
		}

		fmt.Print("auth")
		next.ServeHTTP(w, r)
	})
}
