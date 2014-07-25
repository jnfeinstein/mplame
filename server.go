package main

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/gorilla/websocket"
	"github.com/martini-contrib/render"
	"net/http"
	"runtime"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Frame struct {
	Type int
	Data []byte
	Next *Frame
}

func NewFrame(messageType int, data []byte, next *Frame) *Frame {
	f := new(Frame)
	f.Type = messageType
	f.Data = data
	f.Next = nil
	return f
}

func BlankFrame() *Frame {
	return NewFrame(0, nil, nil)
}

type Sender struct {
	Conn *websocket.Conn
}

type Receiver struct {
	Conn  *websocket.Conn
	Frame *Frame
	Ready chan bool
}

func NewReceiver(conn *websocket.Conn, frame *Frame) *Receiver {
	r := new(Receiver)
	r.Conn = conn
	r.Frame = frame
	r.Ready = make(chan bool)
	return r
}

func (r *Receiver) SendFrame(frame *Frame) error {
	if err := r.Conn.WriteMessage(frame.Type, frame.Data); err != nil {
		return err
	}
	r.Frame = frame
	return nil
}

type ReceiverMap map[*Receiver]bool

type Room struct {
	Name      string
	Frame     *Frame
	Sender    *Sender
	Receivers ReceiverMap
}

func NewRoom(name string) *Room {
	r := new(Room)
	r.Name = name
	r.Frame = BlankFrame()
	r.Sender = nil
	r.Receivers = make(ReceiverMap)
	return r
}

func (r *Room) HandleReceiver(receiver *Receiver) {
	defer r.RemoveReceiver(receiver)
	count := 0
	for {
		for receiver.Frame.Next != nil {
			frame := receiver.Frame.Next
			count++
			fmt.Printf("Sending frame \"%d\" to receiver\n", count)
			if err := receiver.SendFrame(frame); err != nil {
				return
			}
			receiver.Frame = frame
		}
		select {
		case <-receiver.Ready:
		case <-time.After(100 * time.Millisecond):
		}
	}
}

func (r *Room) AddReceiver(conn *websocket.Conn) {
	receiver := NewReceiver(conn, r.Frame)
	r.Receivers[receiver] = true
	go r.HandleReceiver(receiver)
	fmt.Printf("%s has %d receivers\n", r.Name, len(r.Receivers))
}

func (r *Room) RemoveReceiver(receiver *Receiver) {
	receiver.Conn.Close()
	delete(r.Receivers, receiver)
	fmt.Printf("%s has %d receivers\n", r.Name, len(r.Receivers))
}

func (r *Room) AppendFrame(frame *Frame) {
	r.Frame.Next = frame
	r.Frame = frame
}

func (r *Room) HandleSender() {
	count := 0
	for {
		messageType, data, err := r.Sender.Conn.ReadMessage()
		if err != nil {
			r.Sender.Conn.Close()
			return
		}
		count++
		fmt.Printf("Got \"%d\" from sender\n", count)
		r.AppendFrame(NewFrame(messageType, data, nil))

		for receiver, _ := range r.Receivers {
			receiver.Ready <- true
		}
	}
}

func (r *Room) AddSender(conn *websocket.Conn) {
	r.Sender = &Sender{conn}
	go r.HandleSender()
}

type RoomMap map[string]*Room

type RoomCollection struct {
	Map RoomMap
}

func (rc RoomCollection) GetRoom(name string) *Room {
	room, ok := rooms.Map[name]
	if !ok {
		fmt.Printf("Adding room \"%s\"\n", name)
		room = NewRoom(name)
		rc.Map[name] = room
	}
	return room
}

var rooms = RoomCollection{Map: make(RoomMap)}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	m := martini.Classic()

	m.Use(render.Renderer(render.Options{
		Layout: "application",
	}))

	m.Get("/", func(r render.Render) {
		r.HTML(200, "receiver", "broadcast")
	})

	m.Get("/:name", func(r render.Render, p martini.Params) {
		r.HTML(200, "receiver", p["name"])
	})

	m.Get("/:name/s", func(r render.Render, p martini.Params) {
		r.HTML(200, "sender", p["name"])
	})

	m.Get("/sock/:room", func(w http.ResponseWriter, r *http.Request, p martini.Params) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		room := rooms.GetRoom(p["room"])
		room.AddReceiver(conn)
	})

	m.Get("/sock/:room/s", func(w http.ResponseWriter, r *http.Request, p martini.Params) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		room := rooms.GetRoom(p["room"])
		room.AddSender(conn)
		fmt.Printf("%s has a new sender\n", p["room"])
	})

	m.Run()
}