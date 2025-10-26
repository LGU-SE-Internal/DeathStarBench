package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strconv"

	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/tracing"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	Username string `bson:"username"`
	Password string `bson:"password"`
}

func initializeDatabase(url string) (*mongo.Client, func()) {
	tracing.Log.Info().Msg("Generating test data...")

	newUsers := []interface{}{}

	for i := 0; i <= 500; i++ {
		suffix := strconv.Itoa(i)

		password := ""
		for j := 0; j < 10; j++ {
			password += suffix
		}
		sum := sha256.Sum256([]byte(password))

		newUsers = append(newUsers, User{
			fmt.Sprintf("Cornell_%x", suffix),
			fmt.Sprintf("%x", sum),
		})
	}

	uri := fmt.Sprintf("mongodb://%s", url)
	tracing.Log.Info().Msgf("Attempting connection to %v", uri)

	opts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		tracing.Log.Panic().Msg(err.Error())
	}
	tracing.Log.Info().Msg("Successfully connected to MongoDB")

	collection := client.Database("user-db").Collection("user")
	_, err = collection.InsertMany(context.TODO(), newUsers)
	if err != nil {
		tracing.Log.Fatal().Msg(err.Error())
	}
	tracing.Log.Info().Msg("Successfully inserted test data into user DB")

	return client, func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			tracing.Log.Fatal().Msg(err.Error())
		}
	}
}
