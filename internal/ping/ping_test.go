package ping

import (
	"context"
	"errors"
	ConfigBuilder "github.com/keloran/go-config"
	mungo "github.com/keloran/go-config/mongo"
	"github.com/stretchr/testify/assert"
	"github.com/todo-lists-app/ping-service/internal/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"testing"
	"time"
)

type MockMongoOperations struct {
	shouldError bool
	exists      bool
}

func (m *MockMongoOperations) GetMongoClient(ctx context.Context, config mungo.Mongo) error {
	if m.shouldError {
		return errors.New("mock error")
	}
	return nil
}

func (m *MockMongoOperations) Disconnect(ctx context.Context) error {
	return nil
}

func (m *MockMongoOperations) InsertOne(ctx context.Context, document interface{}) (interface{}, error) {
	if m.shouldError {
		return nil, errors.New("mock error")
	}
	return nil, nil
}

func (m *MockMongoOperations) UpdateOne(ctx context.Context, filter interface{}, update interface{}) (interface{}, error) {
	if m.shouldError {
		return nil, errors.New("mock error")
	}
	return nil, nil
}

func (m *MockMongoOperations) FindOne(ctx context.Context, filter interface{}) *mongo.SingleResult {
	doc := bson.D{{"userid", "test"}, {"time", time.Now()}}

	if m.shouldError {
		// This should be adjusted based on how you handle errors in the FindOne method.
		return mongo.NewSingleResultFromDocument(nil, errors.New("mock error"), bson.DefaultRegistry)
	}
	if m.exists {
		// Mocked result
		return mongo.NewSingleResultFromDocument(doc, nil, bson.DefaultRegistry)
	}
	// Simulate a "not found" scenario

	return mongo.NewSingleResultFromDocument(doc, mongo.ErrNoDocuments, bson.DefaultRegistry)
}

func TestPing(t *testing.T) {
	ctx := context.Background()
	cfg := config.Config{
		Config: ConfigBuilder.Config{
			Mongo: mungo.Mongo{
				Database:    "testDB",
				Collections: map[string]string{"ping": "testCollection"},
			},
		},
	}

	t.Run("Ping with no previous ping", func(t *testing.T) {
		pingService := NewPingService(ctx, cfg, "test", &MockMongoOperations{exists: false})

		err := pingService.Ping()
		assert.Nil(t, err)
	})

	t.Run("Ping with previous ping", func(t *testing.T) {
		pingService := NewPingService(ctx, cfg, "test", &MockMongoOperations{exists: true})

		err := pingService.Ping()
		assert.Nil(t, err)
	})

	t.Run("Ping with error", func(t *testing.T) {
		pingService := NewPingService(ctx, cfg, "test", &MockMongoOperations{shouldError: true})

		err := pingService.Ping()
		assert.NotNil(t, err)
	})
}

// You can similarly write tests for pingExists and GetPing methods by extending the MockMongoOperations and adding more test cases.
