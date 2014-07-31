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

type ChatterMap map[*websocket.Conn]bool

type Room struct {
	Name      string
	Frame     *Frame
	Sender    *Sender
	Receivers ReceiverMap
	Chatters  ChatterMap
}

func NewRoom(name string) *Room {
	r := new(Room)
	r.Name = name
	r.Frame = BlankFrame()
	r.Sender = nil
	r.Receivers = make(ReceiverMap)
	r.Chatters = make(ChatterMap)
	return r
}

func (r *Room) HandleChatter(conn *websocket.Conn) {
	defer r.RemoveChatter(conn)
	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		fmt.Printf("%s: new chat\n", r.Name)
		for c, _ := range r.Chatters {
			if c != conn {
				if err := c.WriteMessage(messageType, data); err != nil {
					r.RemoveChatter(c)
				}
			}
		}
	}
}

func (r *Room) AddChatter(conn *websocket.Conn) {
	fmt.Printf("%s: has %d chatters\n", r.Name, len(r.Chatters))
	r.Chatters[conn] = true
	go r.HandleChatter(conn)
}

func (r *Room) RemoveChatter(conn *websocket.Conn) {
	if _, ok := r.Chatters[conn]; ok {
		conn.Close()
		delete(r.Chatters, conn)
		fmt.Printf("%s: has %d chatters\n", r.Name, len(r.Chatters))
	}
}

func (r *Room) HandleReceiver(receiver *Receiver) {
	defer r.RemoveReceiver(receiver)
	for {
		for receiver.Frame.Next != nil {
			frame := receiver.Frame.Next
			fmt.Printf("%s: sending frame to receiver\n", r.Name)
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
	fmt.Printf("%s: has %d receivers\n", r.Name, len(r.Receivers))
}

func (r *Room) RemoveReceiver(receiver *Receiver) {
	if _, ok := r.Receivers[receiver]; ok {
		receiver.Conn.Close()
		delete(r.Receivers, receiver)
		fmt.Printf("%s has %d receivers\n", r.Name, len(r.Receivers))
	}
}

func (r *Room) AppendFrame(frame *Frame) {
	r.Frame.Next = frame
	r.Frame = frame
}

func (r *Room) CheckSender() {
	for {
		<-time.After(1 * time.Second)
		if err := r.Sender.Conn.WriteMessage(1, []byte{1, 2, 3}); err != nil {
			fmt.Printf("%s: sender disconnected\n", r.Name)
			r.RemoveSender()
			return
		}
	}
}

func (r *Room) HandleSender() {
	for {
		messageType, data, err := r.Sender.Conn.ReadMessage()
		if err != nil {
			return
		}

		fmt.Printf("%s: new frame from sender\n", r.Name)
		r.AppendFrame(NewFrame(messageType, data, nil))

		for receiver, _ := range r.Receivers {
			receiver.Ready <- true
		}
	}
}

func (r *Room) HasSender() bool {
	return r.Sender != nil
}

func (r *Room) AddSender(conn *websocket.Conn) {
	fmt.Printf("%s: new sender\n", r.Name)
	r.Sender = &Sender{conn}
	go r.HandleSender()
	go r.CheckSender()
}

func (r *Room) RemoveSender() {
	if r.Sender != nil {
		r.Sender.Conn.Close()
		r.Sender = nil
	}
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

func (rc RoomCollection) GetRoomNames() []string {
	names := make([]string, len(rc.Map))
	idx := 0
	for name, _ := range rc.Map {
		names[idx] = name
		idx++
	}
	return names
}

var rooms = RoomCollection{Map: make(RoomMap)}

type ReceiverParams struct {
	Name      string
	HasSender bool
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	m := martini.Classic()

	m.Use(render.Renderer(render.Options{
		Layout: "application",
	}))

	m.Get("/", func(r render.Render) {
		r.HTML(200, "landing", rooms.GetRoomNames())
	})

	m.Get("/favicon.ico", func() {

	})

	m.Get("/:name", func(r render.Render, p martini.Params) {
		name := p["name"]
		room := rooms.GetRoom(name)
		r.HTML(200, "receiver", ReceiverParams{name, room.HasSender()})
	})

	m.Get("/:name/s", func(r render.Render, p martini.Params) {
		name := p["name"]
		room := rooms.GetRoom(name)
		if room.HasSender() {
			r.HTML(200, "receiver", ReceiverParams{name, true})
		} else {
			r.HTML(200, "sender", name)
		}
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
	})

	m.Get("/sock/:room/c", func(w http.ResponseWriter, r *http.Request, p martini.Params) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		room := rooms.GetRoom(p["room"])
		room.AddChatter(conn)
	})

	m.Run()
}
