package customws

import (
	"encoding/json"
	"fmt"
	"sync"

	structsws "github.com/davidM20/micro-service-backend-go.git/pkg/customws/types"
	"github.com/davidM20/micro-service-backend-go.git/pkg/tools"
	"github.com/gorilla/websocket"
)

var mapChannels map[int64]chan []byte
var checkinMap map[string]chan []byte
var mu sync.Mutex

type Process[T any] struct {
}

func Init[T any]() *Process[T] {
	return &Process[T]{}
}

// initial (Constructor)
func (r *Process[any]) Start() {
	mapChannels = make(map[int64]chan []byte)
	checkinMap = make(map[string]chan []byte)

}

func (r *Process[any]) DeleteUser(userId int64) {

	mu.Lock()
	if ch, ok := mapChannels[userId]; ok {
		close(ch)
		delete(mapChannels, userId)
	} else {
		tools.Rojo("El canal con el ID proporcionado no existe en el mapa: Niuye.,sm7")
	}

	mu.Unlock()
}

func (r *Process[any]) OflineUser(userId int64) bool {

	mu.Lock()
	_, ok := mapChannels[userId]
	mu.Unlock()

	if ok {
		return true
	} else {
		return false
	}
}

type UserWS[T any] struct {
	Conn      *websocket.Conn
	Me        T
	Connected int
	Process[T]
	Pid []byte
}

type Op struct {
	Pid    string
	Status string
}

func (r *UserWS[T]) SendMessageToFriend(friendId int64, message any, pid string) (bool, error) {
	Checking := make(chan []byte, 1)

	// Bloqueo antes de modificar el mapa
	mu.Lock()
	checkinMap[pid] = Checking
	mu.Unlock()

	bytes, err := json.Marshal(message)
	if err != nil {
		tools.Rojo("nsiasodhq", err)
		return false, err
	}

	mu.Lock()
	ch, ok := mapChannels[friendId]
	mu.Unlock()

	if ok {
		select {
		case ch <- bytes:
			// tools.Naranja("Mensaje enviado correctamente por el canal para : ", friendId)
		default:
			tools.Rojo("El canal está bloqueado.", nil)
			mu.Lock()
			delete(checkinMap, pid)
			mu.Unlock()
			return false, nil
		}
	} else {
		tools.Rojo("el canal no existe.", nil)
		mu.Lock()
		delete(checkinMap, pid)
		mu.Unlock()
		return false, fmt.Errorf("%s", err)
	}

	for range Checking {
		mu.Lock()
		delete(checkinMap, pid)
		mu.Unlock()
		return true, nil
	}

	return false, nil
}

func (r *UserWS[T]) ResponseMe(message any) error {

	bytes, err := json.Marshal(message)
	if err != nil {
		tools.Rojo("nsiasodhq", err)
		return err
	}
	mu.Lock()

	err = r.Conn.WriteMessage(websocket.TextMessage, bytes)
	if err != nil {
		tools.Rojo("kjgdiuasdAS", err)
		return err
	}
	mu.Unlock()

	return nil
}

func (r *UserWS[T]) ResponseError(Apiorigin string) error {

	res := structsws.Apiresponse{
		ApiOrigin: Apiorigin,
		Status:    "error",
	}

	bytes, err := json.Marshal(res)
	if err != nil {
		tools.Rojo("nsiasodhq", err)
		return err
	}
	mu.Lock()
	err = r.Conn.WriteMessage(websocket.TextMessage, bytes)
	if err != nil {
		tools.Rojo("kjgdiuasdAS", err)
		return err
	}
	mu.Unlock()

	return nil
}

func (r *UserWS[T]) ResponseMeBytes(message []byte) error {

	mu.Lock()
	err := r.Conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		tools.Rojo("kjgdiuasdAS", err)
		return err
	}
	mu.Unlock()

	return nil
}

func (r *Process[T]) Process(id int64, u *UserWS[T], process func(msg *[]byte, c *UserWS[T]) any, prevCode func(c *UserWS[T]), killCode func(c *UserWS[T])) {

	canal := make(chan []byte, 1)
	mu.Lock()
	mapChannels[id] = canal
	mu.Unlock()

	go prevCode(u)

	// case send to friend
	go func(c *chan []byte) {
		for msg := range *c {
			err := u.ResponseMeBytes(msg)
			if err != nil {
				tools.Rojo("kjbaskjdbkew", err)
			}
		}
	}(&canal)

	go func() {

		for {
			_, message, err := u.Conn.ReadMessage()
			if err != nil {
				tools.Rojo("ndasdlsandl2", err)
				r.DeleteUser(id)
				killCode(u)
				return
			}

			size := len(message)
			if size == 0 {
				continue
			}

			if message[0] == 0x01 {

				tools.Azul("process")
				message2 := message[1:]
				res := process(&message2, u)
				if !isNotNumericOrError(res) {
					err = u.ResponseMe(res)
					if err != nil {
						tools.Rojo("hjdsjdfsade3", err)
					}
				} else {
					if res != 1 {
						tools.Rojo("kjsdjsasad646kjsuduisadu", res)
					}
				}
			}

			if message[0] == 0x02 {
				tools.Naranja("subprocess")

				m := (message)[7:]
				var op Op
				err = json.Unmarshal(m, &op)
				if err != nil {
					tools.Rojo("abdjbdA", err)
				}

				// TODO: Falta agregar que en caso de error se elimine el canal con el pid
				checkinMap[op.Pid] <- message

			}

		}
	}()

}

func GetTotalUsers() int {
	return len(mapChannels)
}

func isNotNumericOrError(s interface{}) bool {
	switch s.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	case error:
		return true
	default:
		return false
	}
}
