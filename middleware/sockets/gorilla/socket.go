package gorilla

import (
	"errors"
	"github.com/fasthttp-contrib/websocket"
	json "github.com/json-iterator/go"
	"github.com/spf13/cast"
	"net"
	"net/http"
	"sync"
	"time"
	"xdf/common"
	"xdf/common/log"
	"xdf/middleware"
	"xdf/model"

	"github.com/kataras/neffos"

	gorilla "github.com/gorilla/websocket"
)

// Socket completes the `neffos.Socket` interface,
// it describes the underline websocket connection.
var Clients sync.Map
var AdminClients sync.Map
var AdminClientsSlice []uint

const PingMessageType = gorilla.PingMessage
const TextMessage = gorilla.TextMessage

var Mu sync.Mutex

type BroadcastMessageType struct {
	To      uint
	Msg     interface{}
	MsgType int //1
}

type Socket struct {
	UnderlyingConn *gorilla.Conn
	request        *http.Request

	client bool

	mu sync.Mutex
}

func GetOneAdminClient(cr *model.ChatRecord) error {
	if cr.To != 0 {
		con, ok := AdminClients.Load(cr.To)
		if ok {
			c, ok := con.(*Socket)
			if ok {
				msg, err := json.Marshal(cr)
				if err != nil {
					log.Error(err)
					return err
				}
				err = c.UnderlyingConn.WriteMessage(gorilla.TextMessage, msg)
				if err == nil {
					return nil
				} else {
					AdminClients.Delete(cr.To)
				}
			}
		}
	}
	for {
		if len(AdminClientsSlice) == 0 {
			return errors.New("暂无客服在线")
		}
		// 重新轮询一个客服
		id := AdminClientsSlice[0]
		con, ok := AdminClients.Load(id)
		if ok {
			//该客服不存在 删除
			c, ok := con.(*Socket)
			if ok {
				cr.To = id
				msg, err := json.Marshal(cr)
				if err != nil {
					log.Error(err)
					return err
				}
				err = c.UnderlyingConn.WriteMessage(gorilla.TextMessage, msg)
				if err == nil {
					Mu.Lock()
					AdminClientsSlice = append(AdminClientsSlice[:0], AdminClientsSlice[1:]...)
					AdminClientsSlice = append(AdminClientsSlice, id)
					Mu.Unlock()
					// 再给用户发一条信息,告诉他是哪个客服
					conUser, ok := Clients.Load(cr.From)
					if ok {
						cUser, ok := conUser.(*Socket)
						if ok {
							crTuUser := &model.ChatRecord{
								From:     id,
								To:       cr.From,
								Type:     1,
								Content:  "客服已经连接成功，请稍等",
								UserName: cr.UserName,
								MsgType:  1,
							}
							msg, err := json.Marshal(crTuUser)
							if err != nil {
								log.Error(err)
								return err
							}
							cUser.UnderlyingConn.WriteMessage(gorilla.TextMessage, msg)
						}
					}
					break
				}
			}
		}
		Mu.Lock()
		AdminClientsSlice = append(AdminClientsSlice[:0], AdminClientsSlice[1:]...)
		Mu.Unlock()
		AdminClients.Delete(id)
	}
	return nil
}

func newSocket(underline *gorilla.Conn, request *http.Request, client bool) (*Socket, error) {
	s := &Socket{
		UnderlyingConn: underline,
		request:        request,
		client:         client,
	}
	query := request.URL.Query()
	token := query.Get("token")
	m, ok := middleware.ParseToken(token)
	if !ok {
		log.Error("token错误")
		s.UnderlyingConn.Close()
		return nil, errors.New("token错误")
	}
	aid := cast.ToUint(m["aid"])
	if aid < 1 {
		log.Error("id错误")
		s.UnderlyingConn.Close()
		return nil, errors.New("id错误")
	}
	tp := cast.ToInt(m["type"])
	if tp == 1 {
		Clients.Store(aid, s)
	} else {
		tempU := &model.AdminUser{}
		tempU.GetUserById(aid)
		roleID := tempU.RoleID
		if roleID == 2 {
			log.Println("插入客服")
			if !common.IsExistInArray(aid, AdminClientsSlice) {
				Mu.Lock()
				log.Println("插入客服成功")
				AdminClientsSlice = append(AdminClientsSlice, aid)
				Mu.Unlock()
			}
			AdminClients.Store(aid, s)
		}
	}
	return s, nil
}

// NetConn returns the underline net connection.
func (s *Socket) NetConn() net.Conn {
	return s.UnderlyingConn.UnderlyingConn()
}

func (s *Socket) Send(uid uint, msg []byte) {
	tmpGroupMap := make(map[uint]interface{})
	Clients.Range(func(k, v interface{}) bool {
		id, ok := k.(uint)
		if ok {
			tmpGroupMap[id] = &v
		}
		return true
	})
	client, ok := Clients.Load(uid)
	if ok {
		c, ok := client.(*Socket)
		if ok {
			err := c.UnderlyingConn.WriteMessage(websocket.TextMessage, msg)
			log.Println(err)
			if err != nil {
				log.Error(err)
			}
			log.Println("发送成功")
		}
	} else {
		log.Println("没找到对方", uid)
	}
}

// Request returns the http request value.
func (s *Socket) Request() *http.Request {
	return s.request
}

// ReadData reads binary or text messages from the remote connection.
func (s *Socket) ReadData(timeout time.Duration) ([]byte, neffos.MessageType, error) {
	for {
		if timeout > 0 {
			s.UnderlyingConn.SetReadDeadline(time.Now().Add(timeout))
		}

		opCode, data, err := s.UnderlyingConn.ReadMessage()
		if err != nil {
			return nil, 0, err
		}

		if opCode != gorilla.BinaryMessage && opCode != gorilla.TextMessage {
			// if gorilla.IsUnexpectedCloseError(err, gorilla.CloseGoingAway) ...
			continue
		}

		return data, neffos.MessageType(opCode), err
	}
}

// WriteBinary sends a binary message to the remote connection.
func (s *Socket) WriteBinary(body []byte, timeout time.Duration) error {
	return s.write(body, gorilla.BinaryMessage, timeout)
}

// WriteText sends a text message to the remote connection.
func (s *Socket) WriteText(body []byte, timeout time.Duration) error {
	return s.write(body, gorilla.TextMessage, timeout)
}

func (s *Socket) write(body []byte, opCode int, timeout time.Duration) error {
	if timeout > 0 {
		s.UnderlyingConn.SetWriteDeadline(time.Now().Add(timeout))
	}

	s.mu.Lock()
	err := s.UnderlyingConn.WriteMessage(opCode, body)
	s.mu.Unlock()

	return err
}
