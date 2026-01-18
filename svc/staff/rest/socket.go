package rest

import (
	"context"
	"errors"
	"fmt"
	"github.com/tespkg/bytes-be/common/global"
	"github.com/tespkg/bytes-be/common/token"
	"net/http"
	"strconv"
	"tespkg.in/kit/log"

	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/polling"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
)

const (
	NamespaceRoot = "/"

	RoleCustomer = "customer"

	EventChatConnectReply = "recv-connect"
)

// Easier to get running with CORS
var allowOriginFunc = func(r *http.Request) bool {
	return true
}

func initSocketIOServer() *socketio.Server {
	socketIOServer := socketio.NewServer(&engineio.Options{
		Transports: []transport.Transport{
			&polling.Transport{
				CheckOrigin: allowOriginFunc,
			},
			&websocket.Transport{
				CheckOrigin: allowOriginFunc,
			},
		},
	})

	return socketIOServer
}

func (s *Server) OnConnection() {
	s.socketServer.OnConnect(NamespaceRoot, func(c socketio.Conn) error {
		url := c.URL()

		log.Infof("namespace %s, conn %s connected\n", NamespaceRoot, c.ID())

		accessToken := url.Query().Get("access_token")
		role := url.Query().Get("role")
		var entityId string
		switch role {
		case RoleCustomer:
			ctx, err := token.VerifyToken(context.Background(), s.db, accessToken)
			if err != nil {
				log.Errorf("token verify failed, token=%s, errors: %s\n", accessToken, err.Error())
				c.Emit(EventChatConnectReply, socketIOResponseErrorWithCode(ResponseSocketIOCode401, err))
			}

			claims := ctx.Value(token.ClaimsCtx).(*token.UserClaims)
			userId, _ := strconv.ParseInt(claims.Subject, 10, 64)

			entityId = strconv.FormatInt(userId, 10)
		default:
			c.Emit(EventChatConnectReply, socketIOResponseErrorWithCode(ResponseSocketIOCode401, errors.New("invalid role")))
			return nil
		}

		existedConn := global.GlobalClientSets.Broadcaster.GetEntityConn(role, entityId)
		if existedConn != nil {
			global.GlobalClientSets.Broadcaster.RemoveEntityByRoleAndId(role, entityId)
			existedConn.Close()
		}

		log.Infof(fmt.Sprintf(`add new entity role %s,entityId %s`, role, entityId))

		global.GlobalClientSets.Broadcaster.AddEntity(role, entityId, c)
		c.Emit(EventChatConnectReply, socketIOResponseSuccess(nil))
		return nil
	})
}

func (s *Server) OnDisconnect() {
	s.socketServer.OnDisconnect(NamespaceRoot, func(c socketio.Conn, reason string) {
		log.Infof("%s's conn %s closed, reason: %s\n", NamespaceRoot, c.ID(), reason)

		if c != nil {
			global.GlobalClientSets.Broadcaster.RemoveByConnId(c.ID())
		}
	})
}

func (s *Server) OnError() {
	s.socketServer.OnError(NamespaceRoot, func(c socketio.Conn, e error) {
		log.Infof("%s's conn %s meet error: %s\n", NamespaceRoot, c.ID(), e.Error())

		if c != nil {
			global.GlobalClientSets.Broadcaster.RemoveByConnId(c.ID())
		}
	})
}

const (
	ResponseSocketIOSuccess = "success"
	ResponseSocketIOError   = "error"
	ResponseSocketIOCode401 = "401"
	ResponseSocketIOCode500 = "500"
)

type SocketIOResponse struct {
	Code string      `json:"code"`
	Msg  *string     `json:"msg"`
	Data interface{} `json:"data"`
}

func socketIOResponseSuccess(data interface{}) *SocketIOResponse {
	return &SocketIOResponse{
		Code: ResponseSocketIOSuccess,
		Data: data,
	}
}

func socketIOResponseError(err error) *SocketIOResponse {
	return &SocketIOResponse{
		Code: ResponseSocketIOError,
		Msg:  NewString(err.Error()),
	}
}

func socketIOResponseErrorWithCode(code string, err error) *SocketIOResponse {
	return &SocketIOResponse{
		Code: code,
		Msg:  NewString(err.Error()),
	}
}

func NewString(s string) *string {
	return &s
}
