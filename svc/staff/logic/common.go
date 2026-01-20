package logic

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/tespkg/bytes-be/common/global"
	"github.com/tespkg/bytes-be/svc/staff/model/dto"
	"math/rand"
	"strings"
	"tespkg.in/go-genproto/sespb"
	"tespkg.in/go-genproto/smspb"
	"time"
)

func SendVerifyCode(ctx context.Context, req *dto.SendVerifyCodeReq) error {
	code := GetCode()

	key := fmt.Sprintf("bytes_be:verify_code:phone:%s", req.Phone)
	if err := global.GlobalClientSets.RedisClient.Set(ctx, key, code, time.Duration(3)*time.Minute).Err(); err != nil {
		return errors.Wrap(err, ">>SendVerifyCode, redis set fail")
	}

	if req.Phone != "" {
		if err := SendMessage(ctx,
			global.GlobalClientSets.SMSClient,
			global.GlobalClientSets.SESClient,
			&smspb.SM{
				PhoneNumber: req.Phone,
				Message:     fmt.Sprintf(`Your OTP number is %s please do not share or reply to this message.`, code),
			},
			&sespb.Email{},
		); err != nil {
			global.GlobalClientSets.RedisClient.Del(ctx, key)
			return errors.Wrap(err, ">>UserSendCodeToMobile ")
		}
	}

	if req.Email != "" {
		//TODO if send verify code to email, please provide the template from the front end.
	}

	return nil
}

func GetCode() string {
	numeric := [9]byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	r := len(numeric)
	rand.Seed(time.Now().UnixNano())

	var sb strings.Builder
	for i := 0; i < 4; i++ {
		fmt.Fprintf(&sb, "%d", numeric[rand.Intn(r)])
	}
	return sb.String()
}

func SendMessage(ctx context.Context, smsCli smspb.SMSClient, sesCli sespb.SESClient, note *smspb.SM, email *sespb.Email) error {
	//send verify code to phone
	if note.PhoneNumber != "" {
		_, err := smsCli.Send(ctx, &smspb.SendRequest{
			ShortMsg: note,
		})
		if err != nil {
			return err
		}
	}

	//send verify code to email
	if email.To != nil && len(email.To) > 0 && !lo.Contains(email.To, "") {
		_, err := sesCli.Send(ctx, &sespb.SendRequest{
			Email: email,
		})
		if err != nil {
			return errors.Wrap(err, ">>SendMessage, sesCli.Send fail")
		}
	}

	return nil
}
