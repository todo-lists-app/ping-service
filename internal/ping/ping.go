package ping

import (
	"context"
	"errors"
	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/todo-lists-app/ping-service/internal/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type Ping struct {
	config.Config
	context.Context
	UserID string
}

type Result struct {
	UserID string    `bson:"userid"`
	Time   time.Time `bson:"time"`
}

func NewPingService(ctx context.Context, cfg config.Config, id string) *Ping {
	return &Ping{
		Config:  cfg,
		Context: ctx,
		UserID:  id,
	}
}

func (p *Ping) Ping() error {
	client, err := config.GetMongoClient(p.Context, p.Config)
	if err != nil {
		return logs.Errorf("error getting mongo client: %v", err)
	}
	defer func() {
		if err := client.Disconnect(p.Context); err != nil {
			_ = logs.Errorf("error disconnecting mongo client: %v", err)
		}
	}()

	prevExists, err := p.pingExists()
	if err != nil {
		return logs.Errorf("error getting prev ping: %v", err)
	}
	if !prevExists {
		if _, err := client.Database(p.Config.Mongo.Database).Collection(p.Config.Mongo.Collections.Ping).InsertOne(p.Context, &bson.M{
			"userid": p.UserID,
			"time":   time.Now(),
		}); err != nil {
			return logs.Errorf("error inserting ping: %v", err)
		}
		return nil
	}

	if _, err := client.Database(p.Config.Mongo.Database).Collection(p.Config.Mongo.Collections.Ping).UpdateOne(p.Context, &bson.M{
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
	client, err := config.GetMongoClient(p.Context, p.Config)
	if err != nil {
		return false, logs.Errorf("error getting mongo client: %v", err)
	}
	defer func() {
		if err := client.Disconnect(p.Context); err != nil {
			_ = logs.Errorf("error disconnecting mongo client: %v", err)
		}
	}()

	res := Result{}
	if err := client.Database(p.Config.Mongo.Database).Collection(p.Config.Mongo.Collections.Ping).FindOne(p.Context, &bson.M{
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
	client, err := config.GetMongoClient(p.Context, p.Config)
	if err != nil {
		return nil, logs.Errorf("error getting mongo client: %v", err)
	}
	defer func() {
		if err := client.Disconnect(p.Context); err != nil {
			_ = logs.Errorf("error disconnecting mongo client: %v", err)
		}
	}()

	res := Result{}
	if err := client.Database(p.Config.Mongo.Database).Collection(p.Config.Mongo.Collections.Ping).FindOne(p.Context, &bson.M{
		"userid": p.UserID,
	}).Decode(&res); err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return &res, logs.Errorf("error finding ping: %v", err)
		}
		return &res, nil
	}

	return &res, nil
}
