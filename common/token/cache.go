package token

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"github.com/tespkg/bytes-be/common/global"
	"github.com/tespkg/bytes-be/svc/staff/model/dao"
	"github.com/tespkg/bytes-be/svc/utils"
	"strconv"
	"time"
)

var RedisUnInitErr = errors.New("redis not init")
var redisKeyPrefix = "bytes-be"

func RefreshUserSecret(dto *GenTokenDto) (string, error) {
	defaultLoginTokenExpireDuration := global.GlobalClientSets.TokenExpireDurationHour

	userId, err := strconv.ParseInt(dto.UserId, 10, 64)
	if err != nil {
		return "", errors.Wrap(err, ">>RefreshUserSecret, userId strconv.ParseInt fail")
	}

	user, err := dao.GetUserById(dto.Session, userId)
	if err != nil {
		return "", errors.Wrap(err, ">>RefreshUserSecret, model.GetUserById fail")
	}

	if !global.GlobalClientSets.EnableSingleLogin {
		if user == nil || user.Id == 0 {
			return "", nil
		}

		if user.Password != nil {
			return string(user.Password), nil
		}
	}

	if global.GlobalClientSets.RedisClient == nil {
		return "", RedisUnInitErr
	}

	secret := utils.RandomString(8)
	redisKey := fmt.Sprintf("%s-user-%s", redisKeyPrefix, dto.UserId)

	//Redis is used for single sign-on when users log in.
	if err = global.GlobalClientSets.RedisClient.Set(context.Background(), redisKey, secret, time.Hour*time.Duration(defaultLoginTokenExpireDuration)).Err(); err != nil {
		return "", errors.Wrap(err, ">>RefreshUserSecret, redis set fail")
	}

	return secret, nil
}

// GetUserSecret get user secret from redis, if nil, generate a new one
func GetUserSecret(dto *GenTokenDto) (string, error) {
	userId, err := strconv.ParseInt(dto.UserId, 10, 64)
	if err != nil {
		return "", errors.Wrap(err, ">>GetUserSecret, strconv.ParseInt fail")
	}

	user, err := dao.GetUserById(dto.Session, userId)
	if err != nil {
		return "", errors.Wrap(err, ">>GetUserSecret, dao.GetUserById fail")
	}

	if !global.GlobalClientSets.EnableSingleLogin {
		if user == nil || user.Id == 0 {
			return jwtSecret, nil
		}

		if user.Password != nil {
			return string(user.Password), nil
		}
	}

	if global.GlobalClientSets.RedisClient == nil {
		return "", RedisUnInitErr
	}

	redisKey := fmt.Sprintf("%s-user-%s", redisKeyPrefix, dto.UserId)
	val, err := global.GlobalClientSets.RedisClient.Get(context.Background(), redisKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return RefreshUserSecret(dto)
		}
		return "", err
	}

	return val, nil
}
