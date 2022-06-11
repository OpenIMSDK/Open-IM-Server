package db

import (
	"Open_IM/pkg/common/config"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"strings"

	//"Open_IM/pkg/common/log"
	"Open_IM/pkg/utils"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/options"

	//	"context"
	//	"fmt"
	"github.com/garyburd/redigo/redis"
	"gopkg.in/mgo.v2"
	"time"

	"context"
	//"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	//	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB DataBases

type DataBases struct {
	MysqlDB     mysqlDB
	mgoSession  *mgo.Session
	redisPool   *redis.Pool
	mongoClient *mongo.Client
}

func key(dbAddress, dbName string) string {
	return dbAddress + "_" + dbName
}

func init() {
	//log.NewPrivateLog(constant.LogFileName)
	var mongoClient *mongo.Client
	var err1 error
	//mysql init
	initMysqlDB()
	// mongo init
	// "mongodb://sysop:moon@localhost/records"
	uri := "mongodb://sample.host:27017/?maxPoolSize=20&w=majority"
	if config.Config.Mongo.DBUri != "" {
		// example: mongodb://$user:$password@mongo1.mongo:27017,mongo2.mongo:27017,mongo3.mongo:27017/$DBDatabase/?replicaSet=rs0&readPreference=secondary&authSource=admin&maxPoolSize=$DBMaxPoolSize
		uri = config.Config.Mongo.DBUri
	} else {
		if config.Config.Mongo.DBPassword != "" && config.Config.Mongo.DBUserName != "" {
			uri = fmt.Sprintf("mongodb://%s:%s@%s/%s?maxPoolSize=%d", config.Config.Mongo.DBUserName, config.Config.Mongo.DBPassword, config.Config.Mongo.DBAddress[0],
				config.Config.Mongo.DBDatabase, config.Config.Mongo.DBMaxPoolSize)
		} else {
			uri = fmt.Sprintf("mongodb://%s/%s/?maxPoolSize=%d",
				config.Config.Mongo.DBAddress[0], config.Config.Mongo.DBDatabase,
				config.Config.Mongo.DBMaxPoolSize)
		}
	}
	mongoClient, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		fmt.Println(" mongo.Connect  failed, try ", utils.GetSelfFuncName(), err.Error(), uri)
		time.Sleep(time.Duration(30) * time.Second)
		mongoClient, err1 = mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
		if err1 != nil {
			fmt.Println(" mongo.Connect retry failed, panic", err.Error(), uri)
			panic(err1.Error())
		}
	}
	fmt.Println("0", utils.GetSelfFuncName(), "mongo driver client init success: ", uri)
	// mongodb create index
	if err := createMongoIndex(mongoClient, cSendLog, false, "send_id", "-send_time"); err != nil {
		fmt.Println("send_id", "-send_time", "index create failed", err.Error())
	}
	if err := createMongoIndex(mongoClient, cChat, true, "uid"); err != nil {
		fmt.Println("uid", " index create failed", err.Error())
	}
	if err := createMongoIndex(mongoClient, cWorkMoment, true, "-create_time", "work_moment_id"); err != nil {
		fmt.Println("-create_time", "work_moment_id", "index create failed", err.Error())
	}
	if err := createMongoIndex(mongoClient, cWorkMoment, true, "work_moment_id"); err != nil {
		fmt.Println("work_moment_id", "index create failed", err.Error())
	}

	if err := createMongoIndex(mongoClient, cWorkMoment, false, "user_id", "-create_time"); err != nil {
		fmt.Println("user_id", "-create_time", "index create failed", err.Error())
	}

	if err := createMongoIndex(mongoClient, cTag, false, "user_id", "-create_time"); err != nil {
		fmt.Println("user_id", "-create_time", "index create failed", err.Error())
	}
	if err := createMongoIndex(mongoClient, cTag, true, "tag_id"); err != nil {
		fmt.Println("user_id", "-create_time", "index create failed", err.Error())
	}
	fmt.Println("create index success")
	DB.mongoClient = mongoClient

	// redis pool init
	DB.redisPool = &redis.Pool{
		MaxIdle:     config.Config.Redis.DBMaxIdle,
		MaxActive:   config.Config.Redis.DBMaxActive,
		IdleTimeout: time.Duration(config.Config.Redis.DBIdleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial(
				"tcp",
				config.Config.Redis.DBAddress,
				redis.DialReadTimeout(time.Duration(1000)*time.Millisecond),
				redis.DialWriteTimeout(time.Duration(1000)*time.Millisecond),
				redis.DialConnectTimeout(time.Duration(1000)*time.Millisecond),
				redis.DialDatabase(config.Config.Redis.DBDatabase),
				redis.DialPassword(config.Config.Redis.DBPassWord),
			)
		},
	}
}

func createMongoIndex(client *mongo.Client, collection string, isUnique bool, keys ...string) error {
	db := client.Database(config.Config.Mongo.DBDatabase).Collection(collection)
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)

	indexView := db.Indexes()
	keysDoc := bsonx.Doc{}

	// 复合索引
	for _, key := range keys {
		if strings.HasPrefix(key, "-") {
			keysDoc = keysDoc.Append(strings.TrimLeft(key, "-"), bsonx.Int32(-1))
		} else {
			keysDoc = keysDoc.Append(key, bsonx.Int32(1))
		}
	}

	// 创建索引
	index := mongo.IndexModel{
		Keys: keysDoc,
	}
	if isUnique == true {
		index.Options = options.Index().SetUnique(true)
	}
	result, err := indexView.CreateOne(
		context.Background(),
		index,
		opts,
	)
	if err != nil {
		return utils.Wrap(err, result)
	}
	return nil
}
