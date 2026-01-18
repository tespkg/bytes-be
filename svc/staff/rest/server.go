package rest

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-module/carbon/v2"
	socketio "github.com/googollee/go-socket.io"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	cors "github.com/rs/cors/wrapper/gin"
	"github.com/sirupsen/logrus"
	"github.com/tespkg/bytes-be/config"
	"github.com/tespkg/bytes-be/internal/ingredient"
	bytesmatch "github.com/tespkg/bytes-be/proto/bytes_match"
	"github.com/tespkg/bytes-be/svc/utils"
	"github.com/tespkg/clickpay"
	"github.com/tespkg/smartpay"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"tespkg.in/go-genproto/sespb"
	"tespkg.in/go-genproto/smspb"
	"time"
)

type Server struct {
	config     config.Config
	httpServer *http.Server
	engin      *gin.Engine
	corsConfig cors.Options

	redisCli *redis.Client
	db       *gorm.DB

	smsClient smspb.SMSClient
	sesClient sespb.SESClient

	nominatimClient *utils.Client

	bytesMatchClient bytesmatch.BytesMatchClient

	theSp       smartpay.SmartPay
	theClickPay clickpay.ClickPay

	ingredientAnalysis ingredient.Analysis

	socketServer *socketio.Server

	shutdownChan chan struct{}
}

func (s *Server) Name() string {
	return "bytes backend rest"
}

func (s *Server) Load(config config.Config) error {
	s.shutdownChan = make(chan struct{}, 1)
	s.config = config

	//load cors
	s.corsConfig = cors.Options{
		AllowedOrigins:   strings.Split(s.config.CorsHosts, ","),
		AllowedMethods:   []string{"HEAD", "GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}

	//load carbon
	s.loadCarbon()

	//load SES/SMS client
	if s.config.SessmsAddr != "" {
		smsConn, err := newGrpcConn(s.config.SessmsAddr, "", "", "")
		if err != nil {
			return errors.Wrap(err, "couldn't connect to ses grpc")
		}
		s.sesClient = sespb.NewSESClient(smsConn)
		s.smsClient = smspb.NewSMSClient(smsConn)
	}

	//load redis
	if err := s.loadRedis(); err != nil {
		return err
	}

	//load postgres
	if err := s.loadPostgres(); err != nil {
		return err
	}

	//load nominatim client
	if s.config.NominatimAddr != "" {
		s.nominatimClient = utils.NewClient(s.config.NominatimAddr)
	}

	//load bytes match grpc client
	if err := s.loadBytesMatch(); err != nil {
		return err
	}

	//load smart pay
	if err := s.loadSmartPay(); err != nil {
		return err
	}

	//load client pay
	if err := s.loadClickPay(); err != nil {
		return err
	}

	//load ingredient analysis
	if err := s.loadIngredientAnalysis(); err != nil {
		return err
	}

	//init socket
	s.socketServer = initSocketIOServer()

	//set global

	s.engin = gin.New()
	s.ginRouter()

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.config.REST.Host, s.config.REST.Port),
		Handler: s.engin,
	}
	s.httpServer = httpServer

	return nil
}

func (s *Server) Run(readyChan chan bool) {
	logrus.WithField("host", s.config.REST.Host).WithField("port", s.config.REST.Port).Infof("starting REST server...")

	s.OnConnection()
	s.OnDisconnect()
	s.OnError()

	go func() {
		if err := s.socketServer.Serve(); err != nil {
			logrus.Infof("socketio listen error: %s\n", err)
		}
	}()
	defer s.socketServer.Close()

	listener, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		logrus.WithField("err", err).Fatal("unable to listen REST")
		panic("unable to listen REST")
	}

	readyChan <- true

	if err = doServe(s.config.REST, s.httpServer, listener); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			logrus.Info("serve REST shutdown success")
			close(s.shutdownChan)
		} else {
			logrus.WithField("err", err).Fatal("unable to serve REST")
			panic("unable to serve REST")
		}
	}

}

func (s *Server) Stop(signal os.Signal) {
	logrus.Info("shutting down serve REST...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		logrus.WithField("err", err).Fatal("unable to graceful shutdown REST")
	}

	select {
	case <-ctx.Done():
		logrus.Fatal("timeout to graceful shutdown REST")
	case <-s.shutdownChan:
		break
	}

	if err := s.redisCli.Close(); err != nil {
		logrus.WithField("err", err).Error("unable to close redis")
	} else {
		logrus.Info("REST redis closed")
	}
}

var weekDays = []string{
	carbon.Sunday,    // 0
	carbon.Monday,    // 1
	carbon.Tuesday,   // 2
	carbon.Wednesday, // 3
	carbon.Thursday,  // 4
	carbon.Friday,    // 5
	carbon.Saturday,  // 6
}

func (s *Server) loadCarbon() error {
	carbon.SetDefault(carbon.Default{
		Timezone:     s.config.Carbon.Timezone,
		WeekStartsAt: weekDays[s.config.Carbon.WeekStartsAt],
	})

	return nil
}

func (s *Server) loadRedis() error {
	options, err := redis.ParseURL(s.config.Redis.Url)
	if err != nil {
		logrus.WithField("error", err).Fatalf("failt to parse redis url")
		return err
	}

	redisClient := redis.NewClient(options)

	//ping
	if err = redisClient.Ping(context.Background()).Err(); err != nil {
		logrus.WithField("error", err).Fatalf("fail to load redis")
		return err
	}

	s.redisCli = redisClient

	logrus.Infof("redis connected")
	return nil
}

func (s *Server) loadPostgres() error {
	if err := utils.Migrate(s.config.Postgres.Dsn); err != nil {
		return err
	}

	db, err := gorm.Open(
		postgres.Open(s.config.Postgres.Dsn),
		&gorm.Config{
			AllowGlobalUpdate: false,
			Logger: logger.Default.LogMode(
				getGormLoggerLevel(s.config.Postgres.LogLevel),
			),
		},
	)
	if err != nil {
		return err
	}

	s.db = db

	logrus.Info("postgres connected")
	return nil
}

func (s *Server) loadBytesMatch() error {
	if s.config.BytesMatch.Address == "" {
		return nil
	}

	conn, err := grpc.Dial(s.config.BytesMatch.Address, grpc.WithInsecure())
	if err != nil {
		return errors.Wrap(err, "couldn't connect to bytes match grpc")
	}

	s.bytesMatchClient = bytesmatch.NewBytesMatchClient(conn)

	return nil
}

func (s *Server) loadSmartPay() error {
	if s.config.SmartPay == "" {
		return errors.New("no config for smartpay")
	}

	sp, err := smartpay.New(smartpay.WithConfigFile(s.config.SmartPay))
	if err != nil {
		return errors.Wrap(err, "new smart pay fail")
	}

	s.theSp = sp

	return nil
}

func (s *Server) loadClickPay() error {
	if s.config.ClickPay == "" {
		return errors.New("no config for clickpay")
	}

	cp, err := clickpay.New(clickpay.WithConfigFile(s.config.ClickPay))
	if err != nil {
		return errors.New("new click pay fail :" + err.Error())
	}

	s.theClickPay = cp

	return nil
}

func (s *Server) loadIngredientAnalysis() error {
	if !s.config.EnableIngredientAnalysis {
		return nil
	}

	ia, err := ingredient.New()
	if err != nil {
		return errors.New("init ingredient analysis fail :" + err.Error())
	}

	s.ingredientAnalysis = ia

	return nil
}

func newGrpcConn(hostAndPort, caPath, clientCrt, clientKey string) (*grpc.ClientConn, error) {
	if caPath == "" {
		return grpc.Dial(hostAndPort, grpc.WithInsecure())
	}

	if caPath != "" && (clientCrt == "" || clientKey == "") {
		return nil, errors.Errorf("invalid credentials for:%v", hostAndPort)
	}

	cPool := x509.NewCertPool()
	caCert, err := ioutil.ReadFile(caPath)
	if err != nil {
		return nil, errors.Errorf("invalid CA crt file: %s", caPath)
	}
	if cPool.AppendCertsFromPEM(caCert) != true {
		return nil, errors.Errorf("failed to parse CA crt")
	}

	clientCert, err := tls.LoadX509KeyPair(clientCrt, clientKey)
	if err != nil {
		return nil, errors.Errorf("invalid client crt file: %s", caPath)
	}

	clientTLSConfig := &tls.Config{
		RootCAs:      cPool,
		Certificates: []tls.Certificate{clientCert},
	}
	creds := credentials.NewTLS(clientTLSConfig)

	conn, err := grpc.Dial(hostAndPort, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, errors.Errorf("dial %v failed: %v", hostAndPort, err)
	}

	return conn, nil
}

func doServe(cfg config.ServerREST, httpServer *http.Server, listener net.Listener) error {
	if cfg.CertPath == "" || cfg.KeyPath == "" {
		return httpServer.Serve(listener)
	}
	return httpServer.ServeTLS(listener, cfg.CertPath, cfg.KeyPath)
}

func getGormLoggerLevel(logLevel string) logger.LogLevel {
	switch logLevel {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn":
		return logger.Warn
	case "info":
		return logger.Info
	default:
		return logger.Info
	}
}
