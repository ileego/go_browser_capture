package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	conn     *websocket.Conn
	server   *Server
	send     chan []byte
	clientID string
}

type Server struct {
	clients            map[*Client]bool
	broadcast          chan []byte
	register           chan *Client
	unregister         chan *Client
	pendingRequests    map[string]chan *CaptureResponse
	mu                 sync.Mutex
	onRequest          func(req *CaptureRequest, client *Client)
	onResponse         func(resp *CaptureResponse)
	onHeartbeat        func(clientID string)
	onClientConnect    func(client *Client)
	onClientDisconnect func(client *Client)
	registerPlugin     func(client *Client, clientID string) string
	unregisterPlugin   func(clientID string)
	updateHeartbeat    func(clientID string)
	onResponseExtra    func(resp *CaptureResponse)
}

func NewServer() *Server {
	return &Server{
		clients:         make(map[*Client]bool),
		broadcast:       make(chan []byte),
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		pendingRequests: make(map[string]chan *CaptureResponse),
	}
}

func (s *Server) SetOnRequestHandler(handler func(req *CaptureRequest, client *Client)) {
	s.onRequest = handler
}

func (s *Server) SetOnResponseHandler(handler func(resp *CaptureResponse)) {
	s.onResponse = handler
}

func (s *Server) SetOnHeartbeatHandler(handler func(clientID string)) {
	s.onHeartbeat = handler
}

func (s *Server) SetOnClientConnectHandler(handler func(client *Client)) {
	s.onClientConnect = handler
}

func (s *Server) SetOnClientDisconnectHandler(handler func(client *Client)) {
	s.onClientDisconnect = handler
}

func (s *Server) SetPluginHandlers(register func(client *Client, clientID string) string, unregister func(clientID string), updateHeartbeat func(clientID string)) {
	s.registerPlugin = register
	s.unregisterPlugin = unregister
	s.updateHeartbeat = updateHeartbeat
}

func (s *Server) SetOnResponseExtraHandler(handler func(resp *CaptureResponse)) {
	s.onResponseExtra = handler
}

func (s *Server) Run() {
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client] = true
			s.mu.Unlock()
			log.Printf("WebSocket client connected: %s", client.conn.RemoteAddr())

			if s.onClientConnect != nil {
				s.onClientConnect(client)
			}

		case client := <-s.unregister:
			s.mu.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.send)
				log.Printf("WebSocket client disconnected: %s", client.conn.RemoteAddr())
			}
			s.mu.Unlock()

			if s.unregisterPlugin != nil && client.clientID != "" {
				s.unregisterPlugin(client.clientID)
			}

			if s.onClientDisconnect != nil {
				s.onClientDisconnect(client)
			}

		case message := <-s.broadcast:
			s.mu.Lock()
			for client := range s.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(s.clients, client)
				}
			}
			s.mu.Unlock()
		}
	}
}

func (s *Server) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	client := &Client{
		conn:   conn,
		server: s,
		send:   make(chan []byte, 256),
	}

	s.register <- client

	go client.readPump()
	go client.writePump()
}

func (c *Client) readPump() {
	defer func() {
		c.server.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)

	for {
		var msg map[string]interface{}
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}

		if action, ok := msg["action"].(string); ok {
			switch action {
			case "capture":
				var req CaptureRequest
				data, _ := json.Marshal(msg)
				json.Unmarshal(data, &req)
				log.Printf("Received capture request: ID=%s, URL=%s, Selector=%s", req.ID, req.URL, req.Selector)

				if c.server.onRequest != nil {
					c.server.onRequest(&req, c)
				}

			case "response":
				var resp CaptureResponse
				data, _ := json.Marshal(msg)
				json.Unmarshal(data, &resp)
				log.Printf("Received capture response: ID=%s, Success=%v", resp.ID, resp.Success)

				// 从消息中提取 pluginId（如果存在）
				if pluginID, ok := msg["pluginId"].(string); ok {
					resp.PluginID = pluginID
				}

				// 处理 server 自己的 pendingRequests
				c.server.mu.Lock()
				if ch, ok := c.server.pendingRequests[resp.ID]; ok {
					delete(c.server.pendingRequests, resp.ID)
					c.server.mu.Unlock()
					ch <- &resp
				} else {
					c.server.mu.Unlock()
				}

				// 调用 onResponse（通常是 TaskScheduler）
				if c.server.onResponse != nil {
					c.server.onResponse(&resp)
				}

				// 调用额外的响应处理器（用于 Handler）
				if c.server.onResponseExtra != nil {
					c.server.onResponseExtra(&resp)
				}

			case "heartbeat":
				pluginID := ""
				if id, ok := msg["pluginId"].(string); ok {
					pluginID = id
				} else if id, ok := msg["client_id"].(string); ok {
					pluginID = id
				}
				log.Printf("Received heartbeat from plugin: %s", pluginID)

				if c.server.onHeartbeat != nil {
					c.server.onHeartbeat(pluginID)
				}

				if c.server.updateHeartbeat != nil {
					c.server.updateHeartbeat(pluginID)
				}

				c.sendHeartbeatResponse(pluginID)

			case "register":
				pluginID := ""
				if id, ok := msg["pluginId"].(string); ok {
					pluginID = id
				}
				log.Printf("Received plugin registration: %s", pluginID)

				if c.server.registerPlugin != nil {
					if pluginID != "" {
						c.clientID = pluginID
					}
					registeredID := c.server.registerPlugin(c, pluginID)
					c.clientID = registeredID
					log.Printf("Plugin registered with ID: %s", c.clientID)
				}

				resp := map[string]interface{}{
					"action":    "registered",
					"pluginId":  c.clientID,
					"timestamp": time.Now().Format(time.RFC3339),
				}
				data, _ := json.Marshal(resp)
				c.send <- data
			}
		}
	}
}

func (c *Client) sendHeartbeatResponse(clientID string) {
	resp := map[string]interface{}{
		"action":    "heartbeat",
		"client_id": clientID,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	data, _ := json.Marshal(resp)
	c.send <- data
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for message := range c.send {
		c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			break
		}
	}
}

func (c *Client) SendResponse(resp *CaptureResponse) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	c.send <- data
	return nil
}

func (c *Client) SendMessage(data []byte) {
	c.send <- data
}

func (s *Server) BroadcastMessage(message []byte) {
	s.broadcast <- message
}

func (s *Server) SendToClient(client *Client, message []byte) {
	client.send <- message
}

func (s *Server) GetClientCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.clients)
}

func (s *Server) RegisterPendingRequest(requestID string, ch chan *CaptureResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pendingRequests[requestID] = ch
}

func (s *Server) RemovePendingRequest(requestID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.pendingRequests, requestID)
}

func (s *Server) GetPendingRequest(requestID string) (chan *CaptureResponse, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ch, ok := s.pendingRequests[requestID]
	return ch, ok
}

