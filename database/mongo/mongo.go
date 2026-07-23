package mongo

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	"github.com/ao9911/go-matrix/util/xtime"
)

type Config struct {
	Addrs         []string
	Username      string
	Password      string
	MaxPool       uint64
	ReplicaSet    string
	ConnTimeout   xtime.Duration
	MaxIdletime   xtime.Duration
	SocketTimeout xtime.Duration
}

func NewMongo(c *Config) (client *mongo.Client) {
	var err error
	clientOptions := options.Client()
	if c.Username != "" && c.Password != "" {
		clientOptions.SetAuth(options.Credential{Username: c.Username, Password: c.Password})
	}
	if c.ReplicaSet != "" {
		clientOptions.SetReplicaSet(c.ReplicaSet)
	}
	clientOptions.SetHosts(c.Addrs)
	clientOptions.SetConnectTimeout(time.Duration(c.ConnTimeout))
	clientOptions.SetMaxPoolSize(c.MaxPool)
	clientOptions.SetCompressors([]string{"zstd"}).SetZstdLevel(3)
	clientOptions.SetMaxConnIdleTime(time.Duration(c.MaxIdletime))
	clientOptions.SetSocketTimeout(time.Duration(c.SocketTimeout))
	clientOptions.SetHeartbeatInterval(time.Second)
	clientOptions.SetReadPreference(readpref.SecondaryPreferred())
	clientOptions.SetReadConcern(readconcern.Local())
	journal := true
	clientOptions.SetWriteConcern(&writeconcern.WriteConcern{W: 1, Journal: &journal})
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalf("mongo.Connect error %v", err)
		return
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatalf("client.Ping error %v", err)
	}
	return
}
