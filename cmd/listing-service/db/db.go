package db

import (
	"context"
	"fmt"
	"log"
	"time"

	configs "github.com/matteeyao/listing-service/configs"
	"github.com/matteeyao/listing-service/graph/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DB struct {
	client *mongo.Client
}

func ConnectDB() *DB {
	client, err := mongo.NewClient(options.Client().ApplyURI(configs.EnvMongoURI()))
	if err != nil {
		log.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	//ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB")
	return &DB{client: client}
}

func colHelper(db *DB, collectionName string) *mongo.Collection {
	return db.client.Database("listings-db").Collection(collectionName)
}

func (db *DB) CreateListing(input *model.NewListing) (*model.Listing, error) {
	collection := colHelper(db, "listing")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := collection.InsertOne(ctx, input)

	if err != nil {
		return nil, err
	}

	listing := &model.Listing{
		ID:          res.InsertedID.(primitive.ObjectID).Hex(),
		OwnerID:     input.OwnerID,
		Description: input.Description,
		Location:    input.Location,
		CreatedAt:   time.Now(),
		Status:      model.StatusNotStarted,
	}

	return listing, err
}

func (db *DB) CreateOwner(input *model.NewOwner) (*model.Owner, error) {
	collection := colHelper(db, "owner")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := collection.InsertOne(ctx, input)

	if err != nil {
		return nil, err
	}

	owner := &model.Owner{
		ID:    res.InsertedID.(primitive.ObjectID).Hex(),
		Name:  input.Name,
		Email: input.Email,
		Phone: input.Phone,
	}

	return owner, err
}

func (db *DB) GetOwners() ([]*model.Owner, error) {
	collection := colHelper(db, "owner")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	var owners []*model.Owner
	defer cancel()

	res, err := collection.Find(ctx, bson.M{})

	if err != nil {
		return nil, err
	}

	defer res.Close(ctx)
	for res.Next(ctx) {
		var singleOwner *model.Owner
		if err = res.Decode(&singleOwner); err != nil {
			log.Fatal(err)
		}
		owners = append(owners, singleOwner)
	}

	return owners, err
}

func (db *DB) GetListings() ([]*model.Listing, error) {
	collection := colHelper(db, "listing")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	var listings []*model.Listing
	defer cancel()

	res, err := collection.Find(ctx, bson.M{})

	if err != nil {
		return nil, err
	}

	defer res.Close(ctx)
	for res.Next(ctx) {
		var singleListing *model.Listing
		if err = res.Decode(&singleListing); err != nil {
			log.Fatal(err)
		}
		listings = append(listings, singleListing)
	}

	return listings, err
}

func (db *DB) GetOwner(ID string) (*model.Owner, error) {
	collection := colHelper(db, "owner")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	var owner *model.Owner
	defer cancel()

	objId, _ := primitive.ObjectIDFromHex(ID)

	err := collection.FindOne(ctx, bson.M{"_id": objId}).Decode(&owner)

	return owner, err
}

func (db *DB) GetListing(ID string) (*model.Listing, error) {
	collection := colHelper(db, "listing")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	var listing *model.Listing
	defer cancel()

	objId, _ := primitive.ObjectIDFromHex(ID)

	err := collection.FindOne(ctx, bson.M{"_id": objId}).Decode(&listing)

	return listing, err
}
