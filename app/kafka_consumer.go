package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

// KafkaConsumerService handles consuming events from Kafka topics
type KafkaConsumerService struct {
	consumer sarama.Consumer
	topics   []string
	handlers map[string]EventHandler
	mu       sync.RWMutex
	running  bool
	ctx      context.Context
	cancel   context.CancelFunc
}

// EventHandler defines the interface for handling different types of events
type EventHandler func(ctx context.Context, event *CrossServiceEvent) error

// NewKafkaConsumerService creates a new Kafka consumer service
func NewKafkaConsumerService(brokers []string, topics []string) (*KafkaConsumerService, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	service := &KafkaConsumerService{
		consumer: consumer,
		topics:   topics,
		handlers: make(map[string]EventHandler),
		ctx:      ctx,
		cancel:   cancel,
	}

	// Register default handlers
	service.registerDefaultHandlers()

	return service, nil
}

// RegisterHandler registers an event handler for a specific event type
func (s *KafkaConsumerService) RegisterHandler(eventType string, handler EventHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[eventType] = handler
}

// registerDefaultHandlers registers default handlers for common event types
func (s *KafkaConsumerService) registerDefaultHandlers() {
	// Billing events
	s.RegisterHandler("billing.user.subscription.created", s.handleBillingEvent)
	s.RegisterHandler("billing.user.subscription.updated", s.handleBillingEvent)
	s.RegisterHandler("billing.user.subscription.cancelled", s.handleBillingEvent)
	s.RegisterHandler("billing.payment.completed", s.handleBillingEvent)
	s.RegisterHandler("billing.payment.failed", s.handleBillingEvent)

	// Auth events
	s.RegisterHandler("auth.user.login", s.handleAuthEvent)
	s.RegisterHandler("auth.user.logout", s.handleAuthEvent)
	s.RegisterHandler("auth.user.registered", s.handleAuthEvent)
	s.RegisterHandler("auth.user.password.changed", s.handleAuthEvent)

	// Payment events
	s.RegisterHandler("payments.transaction.completed", s.handlePaymentEvent)
	s.RegisterHandler("payments.transaction.failed", s.handlePaymentEvent)
	s.RegisterHandler("payments.refund.processed", s.handlePaymentEvent)

	// Generic analytics events
	s.RegisterHandler("analytics.page.view", s.handleAnalyticsEvent)
	s.RegisterHandler("analytics.user.action", s.handleAnalyticsEvent)
	s.RegisterHandler("analytics.conversion", s.handleAnalyticsEvent)
}

// Start begins consuming messages from Kafka topics
func (s *KafkaConsumerService) Start() error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("consumer service is already running")
	}
	s.running = true
	s.mu.Unlock()

	log.Printf("Starting Kafka consumer for topics: %v", s.topics)

	for _, topic := range s.topics {
		go s.consumeTopic(topic)
	}

	return nil
}

// Stop stops the consumer service
func (s *KafkaConsumerService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	s.running = false
	s.cancel()

	if err := s.consumer.Close(); err != nil {
		log.Printf("Error closing Kafka consumer: %v", err)
	}

	log.Println("Kafka consumer service stopped")
}

// consumeTopic consumes messages from a specific topic
func (s *KafkaConsumerService) consumeTopic(topic string) {
	partitionConsumer, err := s.consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Printf("Failed to start consuming from topic %s: %v", topic, err)
		return
	}
	defer partitionConsumer.Close()

	log.Printf("Started consuming from topic: %s", topic)

	for {
		select {
		case <-s.ctx.Done():
			log.Printf("Stopping consumption from topic: %s", topic)
			return
		case msg := <-partitionConsumer.Messages():
			s.handleMessage(msg)
		case err := <-partitionConsumer.Errors():
			log.Printf("Error consuming from topic %s: %v", topic, err)
		}
	}
}

// handleMessage processes a single Kafka message
func (s *KafkaConsumerService) handleMessage(msg *sarama.ConsumerMessage) {
	log.Printf("Received message from topic %s: %s", msg.Topic, string(msg.Value))

	var event CrossServiceEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return
	}

	// Set correlation ID if not present
	if event.CorrelationID == "" {
		event.CorrelationID = uuid.New().String()
	}

	// Route to appropriate handler
	s.routeEvent(&event)
}

// routeEvent routes an event to the appropriate handler
func (s *KafkaConsumerService) routeEvent(event *CrossServiceEvent) {
	s.mu.RLock()
	handler, exists := s.handlers[event.EventType]
	s.mu.RUnlock()

	if !exists {
		log.Printf("No handler registered for event type: %s", event.EventType)
		return
	}

	// Execute handler in goroutine to avoid blocking
	go func() {
		if err := handler(s.ctx, event); err != nil {
			log.Printf("Error handling event %s: %v", event.EventType, err)
		}
	}()
}

// handleBillingEvent handles billing-related events
func (s *KafkaConsumerService) handleBillingEvent(ctx context.Context, event *CrossServiceEvent) error {
	log.Printf("Processing billing event: %s for user: %s", event.EventType, event.UserID)

	// Extract billing data
	amount, _ := event.Data["amount"].(float64)
	description, _ := event.Data["description"].(string)

	// Create billing event
	billingEvent := NewBillingEvent(event.UserID, event.EventType, amount, description)

	// Store or process the billing event
	log.Printf("Created billing event: %s with amount: %.2f", billingEvent.ID, billingEvent.Amount)

	return nil
}

// handleAuthEvent handles authentication-related events
func (s *KafkaConsumerService) handleAuthEvent(ctx context.Context, event *CrossServiceEvent) error {
	log.Printf("Processing auth event: %s for user: %s", event.EventType, event.UserID)

	// Process authentication events (e.g., track user sessions, security metrics)
	log.Printf("Processed auth event: %s", event.EventType)

	return nil
}

// handlePaymentEvent handles payment-related events
func (s *KafkaConsumerService) handlePaymentEvent(ctx context.Context, event *CrossServiceEvent) error {
	log.Printf("Processing payment event: %s for user: %s", event.EventType, event.UserID)

	// Process payment events (e.g., track transaction volumes, success rates)
	log.Printf("Processed payment event: %s", event.EventType)

	return nil
}

// handleAnalyticsEvent handles analytics-related events
func (s *KafkaConsumerService) handleAnalyticsEvent(ctx context.Context, event *CrossServiceEvent) error {
	log.Printf("Processing analytics event: %s for user: %s", event.EventType, event.UserID)

	// Process analytics events (e.g., aggregate metrics, update dashboards)
	log.Printf("Processed analytics event: %s", event.EventType)

	return nil
}
