package master

import (
	"context"
	"go-cron/internal/common"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type JobLogMgr struct {
	client        *mongo.Client
	logCollection *mongo.Collection
}

var (
	G_jobLogMgr *JobLogMgr
)

func (jobLogMrg *JobLogMgr) JobLogList(name string, page, pageSize int64) (jobLogList []*common.JobLog, err error) {
	var (
		findOptions *options.FindOptions
		cursor      *mongo.Cursor
	)

	jobLogList = make([]*common.JobLog, 0)

	findOptions = options.Find().SetSort(bson.M{"startTime": -1}).SetSkip((page - 1) * pageSize).SetLimit(pageSize)
	if cursor, err = jobLogMrg.logCollection.Find(context.TODO(), bson.M{"jobName": name}, findOptions); err != nil {
		return
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var jobLog common.JobLog
		if err = cursor.Decode(&jobLog); err != nil {
			continue
		}
		jobLogList = append(jobLogList, &jobLog)
	}

	return
}

func InitJobLogMgr() (err error) {
	var (
		client    *mongo.Client
		jobLogMrg *JobLogMgr
	)

	if client, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(G_config.MongoUri)); err != nil {
		return
	}

	jobLogMrg = &JobLogMgr{
		client:        client,
		logCollection: client.Database("cron").Collection("log"),
	}

	G_jobLogMgr = jobLogMrg

	return
}
