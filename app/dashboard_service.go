package app

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
)

// DashboardService provides real-time analytics data for dashboards
type DashboardService struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan interface{}
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.RWMutex
}

// DashboardMetric represents a real-time metric for dashboards
type DashboardMetric struct {
	Type      string                 `json:"type"`
	Value     interface{}            `json:"value"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// DashboardEvent represents a real-time event for dashboards
type DashboardEvent struct {
	EventType string                 `json:"event_type"`
	UserID    string                 `json:"user_id"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewDashboardService creates a new dashboard service instance
func NewDashboardService() *DashboardService {
	return &DashboardService{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan interface{}, 100),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

// Start begins the dashboard service
func (s *DashboardService) Start() {
	go s.run()
}

// run handles the main service loop
func (s *DashboardService) run() {
	for {
		select {
		case client := <-s.register:
			s.mutex.Lock()
			s.clients[client] = true
			s.mutex.Unlock()
			log.Printf("Dashboard client connected. Total clients: %d", len(s.clients))

		case client := <-s.unregister:
			s.mutex.Lock()
			delete(s.clients, client)
			s.mutex.Unlock()
			log.Printf("Dashboard client disconnected. Total clients: %d", len(s.clients))

		case message := <-s.broadcast:
			s.broadcastToClients(message)
		}
	}
}

// broadcastToClients sends a message to all connected clients
func (s *DashboardService) broadcastToClients(message interface{}) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	for client := range s.clients {
		err := client.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Printf("Error sending message to client: %v", err)
			client.Close()
			delete(s.clients, client)
		}
	}
}

// HandleWebSocket handles WebSocket connections for real-time dashboard
func (s *DashboardService) HandleWebSocket(c *websocket.Conn) {
	// Register the client
	s.register <- c

	// Handle incoming messages from client
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle client message (e.g., subscription to specific metrics)
		s.handleClientMessage(c, message)
	}

	// Unregister the client when done
	s.unregister <- c
}

// handleClientMessage processes messages from dashboard clients
func (s *DashboardService) handleClientMessage(conn *websocket.Conn, message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Error parsing client message: %v", err)
		return
	}

	// Handle subscription requests
	if msgType, ok := msg["type"].(string); ok {
		switch msgType {
		case "subscribe":
			if metric, ok := msg["metric"].(string); ok {
				s.sendMetricUpdate(conn, metric)
			}
		case "ping":
			// Respond to ping with pong
			conn.WriteJSON(map[string]string{"type": "pong"})
		}
	}
}

// sendMetricUpdate sends a specific metric update to a client
func (s *DashboardService) sendMetricUpdate(conn *websocket.Conn, metric string) {
	// Generate mock metric data for now
	// In production, this would query real-time data sources
	metricData := DashboardMetric{
		Type:      metric,
		Value:     generateMockMetricValue(metric),
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"source": "analytics",
		},
	}

	conn.WriteJSON(metricData)
}

// generateMockMetricValue generates mock values for different metrics
func generateMockMetricValue(metric string) interface{} {
	switch metric {
	case "total_events":
		return 1234
	case "active_users":
		return 567
	case "events_per_minute":
		return 89
	case "conversion_rate":
		return 0.045
	default:
		return "N/A"
	}
}

// BroadcastEvent broadcasts an analytics event to all dashboard clients
func (s *DashboardService) BroadcastEvent(event *AnalyticsEvent) {
	dashboardEvent := DashboardEvent{
		EventType: event.EventType,
		UserID:    event.UserID,
		Data: map[string]interface{}{
			"event_id":   event.ID,
			"page":       event.Page,
			"properties": event.Properties,
		},
		Timestamp: event.Timestamp,
	}

	s.broadcast <- dashboardEvent
}

// BroadcastMetric broadcasts a metric update to all dashboard clients
func (s *DashboardService) BroadcastMetric(metric DashboardMetric) {
	s.broadcast <- metric
}

// GetConnectedClientsCount returns the number of connected dashboard clients
func (s *DashboardService) GetConnectedClientsCount() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.clients)
}
