package global

import (
	"encoding/json"
	"fmt"
	"sync"

	socketio "github.com/googollee/go-socket.io"
	"tespkg.in/kit/log"
)

const (
	BroadcastRoleUser        = "user"
	BroadcastRoleMerchant    = "merchant"
	BroadcastRoleBranch      = "branch"
	BroadcastRoleDriver      = "driver"
	BroadcastRoleFreshBranch = "fresh_branch"
)

/*
Broadcast
memory based store, to keep user and merchant socket connection
*/
type Broadcast struct {
	connEntities map[string]string        // map value is 'user-<user_id>' or 'merchant-<merchant_id>'
	entityConns  map[string]socketio.Conn // map key is 'user-<user_id>' or 'merchant-<merchant_id>'
	lock         sync.RWMutex
}

/*
NewBroadcast
creates a new broadcast adapter
*/
func NewBroadcast() *Broadcast {
	return &Broadcast{
		entityConns:  make(map[string]socketio.Conn),
		connEntities: make(map[string]string),
	}
}

/*
AddUser
Add adds the given connection to the broadcast
*/
func (bc *Broadcast) AddUser(userId int64, connection socketio.Conn) {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	key := fmt.Sprintf("%s-%d", BroadcastRoleUser, userId)
	bc.entityConns[key] = connection
	bc.connEntities[connection.ID()] = key
}

/*
AddEntity
Add adds the given connection to the broadcast
*/
func (bc *Broadcast) AddEntity(role string, entityId string, connection socketio.Conn) {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	key := fmt.Sprintf("%s-%s", role, entityId)
	bc.entityConns[key] = connection
	bc.connEntities[connection.ID()] = key
}

/*
AddMerchant
Add adds the given connection to the broadcast
*/
func (bc *Broadcast) AddMerchant(merchantId int64, connection socketio.Conn) {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	key := fmt.Sprintf("%s-%d", BroadcastRoleMerchant, merchantId)
	bc.entityConns[key] = connection
	bc.connEntities[connection.ID()] = key
}

func (bc *Broadcast) GetUserConn(userId int64) socketio.Conn {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	key := fmt.Sprintf("%s-%d", BroadcastRoleUser, userId)
	conn, ok := bc.entityConns[key]
	if ok {
		return conn
	}
	return nil
}

func (bc *Broadcast) GetEntityConn(entityRole string, entityId string) socketio.Conn {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	key := fmt.Sprintf("%s-%s", entityRole, entityId)
	conn, ok := bc.entityConns[key]
	if ok {
		return conn
	}
	return nil
}

func (bc *Broadcast) GetEntityConnByKeys(keys []string) []socketio.Conn {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	keysInfo, _ := json.Marshal(keys)
	log.Infof("keys: %s", string(keysInfo))

	entityConnsInfo, _ := json.Marshal(bc.entityConns)
	log.Infof("entityConns: %s", string(entityConnsInfo))

	var conns []socketio.Conn
	for _, v := range keys {
		conn, ok := bc.entityConns[v]
		if ok {
			conns = append(conns, conn)
		}
	}

	return conns
}

func (bc *Broadcast) GetMerchantConn(merchantId int64) socketio.Conn {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	key := fmt.Sprintf("%s-%d", BroadcastRoleMerchant, merchantId)
	conn, ok := bc.entityConns[key]
	if ok {
		return conn
	}
	return nil
}

// RemoveUserById removes the given connection by userId
func (bc *Broadcast) RemoveUserById(userId int64) {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	key := fmt.Sprintf("%s-%d", BroadcastRoleUser, userId)
	if conn, ok := bc.entityConns[key]; ok {
		delete(bc.connEntities, conn.ID())
		delete(bc.entityConns, key)
	}
}

// RemoveEntityByRoleAndId removes the given connection by role and id
func (bc *Broadcast) RemoveEntityByRoleAndId(role string, entityId string) {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	key := fmt.Sprintf("%s-%s", role, entityId)
	if conn, ok := bc.entityConns[key]; ok {
		delete(bc.connEntities, conn.ID())
		delete(bc.entityConns, key)
	}
}

// RemoveMerchantById removes the given connection by merchantId
func (bc *Broadcast) RemoveMerchantById(merchantId int64) {
	bc.lock.Lock()
	defer bc.lock.Unlock()

	key := fmt.Sprintf("%s-%d", BroadcastRoleMerchant, merchantId)
	if conn, ok := bc.entityConns[key]; ok {
		delete(bc.connEntities, conn.ID())
		delete(bc.entityConns, key)
	}
}

func (ub *Broadcast) RemoveByConnId(connId string) {
	ub.lock.Lock()
	defer ub.lock.Unlock()

	if uuid, ok := ub.connEntities[connId]; ok {
		log.Infof("remove conn %s of %s fro Broadcaster\n", connId, uuid)
		delete(ub.connEntities, connId)
		delete(ub.entityConns, uuid)
	}
}
