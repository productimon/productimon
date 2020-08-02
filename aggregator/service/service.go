package service

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"sync"

	"git.yiad.am/productimon/aggregator/authenticator"
	schema "git.yiad.am/productimon/aggregator/db"
	"git.yiad.am/productimon/aggregator/deviceState"
	"git.yiad.am/productimon/aggregator/notifications"
	"git.yiad.am/productimon/internal"
	spb "git.yiad.am/productimon/proto/svc"
	"github.com/google/uuid"
	"github.com/sethvargo/go-password/password"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	domain    string
	auther    *authenticator.Authenticator
	dbWLock   sync.Mutex
	db        *sql.DB
	log       *zap.Logger
	notifiers map[string]notifications.Notifier

	ds *deviceState.DsMap
}

var (
	flagFirstUser string
)

func init() {
	flag.StringVar(&flagFirstUser, "first_user_email", "admin@productimon.com", "The email address of the auto-created first admin user (only used when running for first time)")
}

func NewService(domain string, auther *authenticator.Authenticator, db *sql.DB, logger *zap.Logger) (*Service, error) {
	// TODO: version control schema and use migrations to update
	if _, err := db.Exec("SELECT 1 FROM users LIMIT 1"); err != nil {
		logger.Info("Initiating database")
		if _, err = db.Exec(string(schema.Data["schema.sql"])); err != nil {
			logger.Error("error init db", zap.Error(err))
			return nil, err
		}
		uid := uuid.New().String()
		rawPwd, err := password.Generate(16, 4, 4, false, false)
		if err != nil {
			logger.Error("error generating password", zap.Error(err))
			return nil, err
		}
		pwd, err := bcrypt.GenerateFromPassword([]byte(rawPwd), bcryptStrength)
		if err != nil {
			logger.Error("error encrypting password", zap.Error(err))
			return nil, err
		}
		if _, err = db.Exec("INSERT INTO users VALUES(?,?,?, TRUE, TRUE)", uid, flagFirstUser, pwd); err != nil {
			logger.Error("error create first user", zap.Error(err))
			return nil, err
		}
		fmt.Println("====================")
		internal.PrintVersion()
		fmt.Printf("Initial Admin User: %s\n", flagFirstUser)
		fmt.Printf("Password: %s\n", rawPwd)
		fmt.Println("====================")
	}
	s := &Service{
		domain:    domain,
		auther:    auther,
		db:        db,
		log:       logger,
		notifiers: make(map[string]notifications.Notifier),
	}
	s.ds = deviceState.NewDsMap(s.lazyInitEidHandler, logger)
	return s, nil
}

func (s *Service) RegisterNotifier(n notifications.Notifier) {
	s.notifiers[n.Name()] = n
}

func (s *Service) Notify(kind, recipient, message string) error {
	n := s.notifiers[kind]
	if n == nil {
		return notifications.ErrNotRegistered
	}
	return n.Notify(recipient, message)
}

func (s *Service) Ping(ctx context.Context, req *spb.DataAggregatorPingRequest) (*spb.DataAggregatorPingResponse, error) {
	rsp := &spb.DataAggregatorPingResponse{
		Payload: req.Payload,
	}
	return rsp, nil
}

func (s Service) DB() *sql.DB {
	return s.db
}

func (s *Service) DBLock() {
	s.dbWLock.Lock()
}

func (s *Service) DBUnlock() {
	s.dbWLock.Unlock()
}
