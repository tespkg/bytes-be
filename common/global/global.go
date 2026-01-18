package global

import (
	"github.com/redis/go-redis/v9"
	"tespkg.in/go-genproto/sespb"
	"tespkg.in/go-genproto/smspb"
)

var GlobalClientSets ClientSet

type ClientSet struct {
	RedisClient             *redis.Client
	EnableSingleLogin       bool
	SMSClient               smspb.SMSClient
	SESClient               sespb.SESClient
	EmailSender             string
	TokenExpireDurationHour int
	JwtSignedSecret         string
	Broadcaster             *Broadcast
}

func SetUp(clientSet ClientSet) {
	if clientSet.RedisClient != nil {
		GlobalClientSets.RedisClient = clientSet.RedisClient
	}

	if clientSet.SMSClient != nil {
		GlobalClientSets.SMSClient = clientSet.SMSClient
	}

	if clientSet.SESClient != nil {
		GlobalClientSets.SESClient = clientSet.SESClient
	}

	GlobalClientSets.EnableSingleLogin = clientSet.EnableSingleLogin
	GlobalClientSets.EmailSender = clientSet.EmailSender
	GlobalClientSets.TokenExpireDurationHour = clientSet.TokenExpireDurationHour
	GlobalClientSets.Broadcaster = clientSet.Broadcaster
	GlobalClientSets.JwtSignedSecret = clientSet.JwtSignedSecret
}
