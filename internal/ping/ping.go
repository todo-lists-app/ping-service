package ping

import (
	"context"
	"errors"
	"github.com/bugfixes/go-bugfixes/logs"
	mungo "github.com/keloran/go-config/mongo"
	"github.com/todo-lists-app/ping-service/internal/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type MongoOperations interface {
	GetMongoClient(ctx context.Context, config mungo.Mongo) error
	Disconnect(ctx context.Context) error
	InsertOne(ctx context.Context, document interface{}) (interface{}, error)
	UpdateOne(ctx context.Context, filter interface{}, update interface{}) (interface{}, error)
	FindOne(ctx context.Context, filter interface{}) *mongo.SingleResult
}

type RealMongoOperations struct {
	Client *mongo.Client
}

func (r *RealMongoOperations) GetMongoClient(ctx context.Context, config mungo.Mongo) error {
	client, err := mungo.GetMongoClient(ctx, config)
	if err != nil {
		return logs.Errorf("error getting mongo client: %v", err)
	}
	r.Client = client
	return nil
}
func (r *RealMongoOperations) Disconnect(ctx context.Context) error {
	return r.Client.Disconnect(ctx)
}
func (r *RealMongoOperations) InsertOne(ctx context.Context, document interface{}) (interface{}, error) {
	return r.Client.Database("ping").Collection("ping").InsertOne(ctx, document)
}
func (r *RealMongoOperations) UpdateOne(ctx context.Context, filter interface{}, update interface{}) (interface{}, error) {
	return r.Client.Database("ping").Collection("ping").UpdateOne(ctx, filter, update)
}
func (r *RealMongoOperations) FindOne(ctx context.Context, filter interface{}) *mongo.SingleResult {
	return r.Client.Database("ping").Collection("ping").FindOne(ctx, filter)
}

type Ping struct {
	config.Config
	context.Context
	UserID string

	MongoOps MongoOperations
}

type Result struct {
	UserID string    `bson:"userid"`
	Time   time.Time `bson:"time"`
}

func NewPingService(ctx context.Context, cfg config.Config, id string, ops MongoOperations) *Ping {
	return &Ping{
		Config:   cfg,
		Context:  ctx,
		UserID:   id,
		MongoOps: ops,
	}
}

func (p *Ping) Ping() error {
	if err := p.MongoOps.GetMongoClient(p.Context, p.Config.Mongo); err != nil {
		return logs.Errorf("error getting mongo client: %v", err)
	}
	defer func() {
		if err := p.MongoOps.Disconnect(p.Context); err != nil {
			_ = logs.Errorf("error disconnecting mongo client: %v", err)
		}
	}()

	prevExists, err := p.pingExists()
	if err != nil {
		return logs.Errorf("error getting prev ping: %v", err)
	}
	if !prevExists {
		if _, err := p.MongoOps.InsertOne(p.Context, &bson.M{
			"userid": p.UserID,
			"time":   time.Now(),
		}); err != nil {
			return logs.Errorf("error inserting ping: %v", err)
		}
		return nil
	}

	if _, err := p.MongoOps.UpdateOne(p.Context, &bson.M{
		"userid": p.UserID,
	}, &bson.M{
		"$set": &bson.M{
			"time": time.Now(),
		},
	}); err != nil {
		return logs.Errorf("error updating ping: %v", err)
	}

	return nil
}

func (p *Ping) pingExists() (bool, error) {
	if err := p.MongoOps.GetMongoClient(p.Context, p.Config.Mongo); err != nil {
		return false, logs.Errorf("error getting mongo client: %v", err)
	}
	defer func() {
		if err := p.MongoOps.Disconnect(p.Context); err != nil {
			_ = logs.Errorf("error disconnecting mongo client: %v", err)
		}
	}()

	res := Result{}
	if err := p.MongoOps.FindOne(p.Context, &bson.M{
		"userid": p.UserID,
	}).Decode(&res); err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return false, logs.Errorf("error finding ping: %v", err)
		}
		return false, nil
	}
	return true, nil
}

func (p *Ping) GetPing() (*Result, error) {
	if err := p.MongoOps.GetMongoClient(p.Context, p.Config.Mongo); err != nil {
		return nil, logs.Errorf("error getting mongo client: %v", err)
	}
	defer func() {
		if err := p.MongoOps.Disconnect(p.Context); err != nil {
			_ = logs.Errorf("error disconnecting mongo client: %v", err)
		}
	}()

	res := Result{}
	if err := p.MongoOps.FindOne(p.Context, &bson.M{
		"userid": p.UserID,
	}).Decode(&res); err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return &res, logs.Errorf("error finding ping: %v", err)
		}
		return &res, nil
	}

	return &res, nil
}
