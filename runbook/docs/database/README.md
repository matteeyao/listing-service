# Create a GraphQL-powered project management endpoint in Golang and MongoDB

## Database logic

* In the `db.go` file:

```go
package configs
import (
	"context"
	"fmt"
	"log"
	"project-mngt-golang-graphql/graph/model"
	"time"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DB struct {
	client *mongo.Client
}

func ConnectDB() *DB {
	client, err := mongo.NewClient(options.Client().ApplyURI(EnvMongoURI()))
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
	return db.client.Database("projectMngt").Collection(collectionName)
}

func (db *DB) CreateProject(input *model.NewProject) (*model.Project, error) {
	collection := colHelper(db, "project")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := collection.InsertOne(ctx, input)

	if err != nil {
		return nil, err
	}

	project := &model.Project{
		ID:          res.InsertedID.(primitive.ObjectID).Hex(),
		OwnerID:     input.OwnerID,
		Name:        input.Name,
		Description: input.Description,
		Status:      model.StatusNotStarted,
	}

	return project, err
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

func (db *DB) GetProjects() ([]*model.Project, error) {
	collection := colHelper(db, "project")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	var projects []*model.Project
	defer cancel()

	res, err := collection.Find(ctx, bson.M{})

	if err != nil {
		return nil, err
	}

	defer res.Close(ctx)
	for res.Next(ctx) {
		var singleProject *model.Project
		if err = res.Decode(&singleProject); err != nil {
			log.Fatal(err)
		}
		projects = append(projects, singleProject)
	}

	return projects, err
}

func (db *DB) SingleOwner(ID string) (*model.Owner, error) {
	collection := colHelper(db, "owner")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	var owner *model.Owner
	defer cancel()

	objId, _ := primitive.ObjectIDFromHex(ID)

	err := collection.FindOne(ctx, bson.M{"_id": objId}).Decode(&owner)

	return owner, err
}

func (db *DB) SingleProject(ID string) (*model.Project, error) {
	collection := colHelper(db, "project")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	var project *model.Project
	defer cancel()

	objId, _ := primitive.ObjectIDFromHex(ID)

	err := collection.FindOne(ctx, bson.M{"_id": objId}).Decode(&project)

	return project, err
}
```

* The snippet above does the following:

  * Imports the required dependencies

  * Create a `DB` struct w/ a `client` field to access MongoDB

  * Creates a `ConnectDB` function that first configures the client to use the correct URI and check for errors

    * Secondly, we defined a timeout of 10 seconds we wanted to use when trying to connect

    * Thirdly, check if there is an error while connecting to the database and cancel the connection if the connecting period exceeds 10 seconds

    * Finally, we pinged the database to test our connection and returned a pointer to the `DB` struct

  * Creates a `colHelper` function to create a collection

  * Creates a `CreateProject` function that takes the `DB` struct as a pointer receiver, and returns either the created `Project` or `Error`

    * Inside the function, we also created a `project` collection, defined a timeout of 10 seconds when inserting data into the collection, and used the `InsertOne` function to insert the `input`

  * Creates a `GetOwners` function that takes the `DB` struct as a pointer receiver, and returns either the list of `Owners` or `Error`

    * The function follows the previous steps by getting the list of owners using the `Find` function

    * We also read the returned list optimally using the `Next` attribute method to loop through the returned list of owners

  * Creates a `GetProjects` function that takes the `DB` struct as a pointer receiver, and returns either the list of `Projects` or `Error`

    * The function follows the previous steps by getting the list of projects using the `Find` function

    * We also read the returned list optimally using the `Next` attribute method to loop through the returned list of projects

  * Creates a `GetOwner` function that takes the `DB` struct as a pointer receiver, and returns either the matched `Owner` using the `FindOne` function or `Error`

  * Creates a `GetProject` function that takes the `DB` struct as a pointer receiver, and returns either the matched `Project` using the `FindOne` function or `Error`

## Updating the application logic

* Next, we need to update the application logic w/ the database functions. To do this, we need to update the `schema.resolvers.go` file as shown below: 

```go
package graph
// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
import (
    "context"
    "project-mngt-golang-graphql/configs" //add this
    "project-mngt-golang-graphql/graph/generated"
    "project-mngt-golang-graphql/graph/model"
)

//add this
var (
    db = configs.ConnectDB()
)

// CreateProject is the resolver for the createProject field.
func (r *mutationResolver) CreateProject(ctx context.Context, input model.NewProject) (*model.Project, error) {
    //modify here
    project, err := db.CreateProject(&input)
    return project, err
}

// CreateOwner is the resolver for the createOwner field.
func (r *mutationResolver) CreateOwner(ctx context.Context, input model.NewOwner) (*model.Owner, error) {
    //modify here
    owner, err := db.CreateOwner(&input)
    return owner, err
}

// Owners is the resolver for the owners field.
func (r *queryResolver) Owners(ctx context.Context) ([]*model.Owner, error) {
    //modify here
    owners, err := db.GetOwners()
    return owners, err
}

// Projects is the resolver for the projects field.
func (r *queryResolver) Projects(ctx context.Context) ([]*model.Project, error) {
    //modify here
    projects, err := db.GetProjects()
    return projects, err
}

// Owner is the resolver for the owner field.
func (r *queryResolver) Owner(ctx context.Context, input *model.FetchOwner) (*model.Owner, error) {
    //modify here
    owner, err := db.SingleOwner(input.ID)
    return owner, err
}

// Project is the resolver for the project field.
func (r *queryResolver) Project(ctx context.Context, input *model.FetchProject) (*model.Project, error) {
    //modify here
    project, err := db.SingleProject(input.ID)
    return project, err
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
```

* The snippet above does the following:

  * Imports the required dependency

  * Creates a `db` variable to initialize the MongoDB using `ConnectDB` function

  * Modifies the `CreateProject`, `CreateOwner`, `Owners`, `Projects`, `Owner`, and `Project` function using their corresponding function from the database logic

* Finally, we need to modify the generated model IDs in the `models_gen.go` file w/ a `bson:"_id"` struct tags

  * We use the struct tags to reformat the JSON `_id` returned by MongoDB:

```go
//The remaining part of the code goes here

type FetchOwner struct {
    ID string `json:"id" bson:"_id"` //modify here
}

type FetchProject struct {
    ID string `json:"id" bson:"_id"` //modify here
}

type NewOwner struct {
    //code goes here
}

type NewProject struct {
    //code goes here
}

type Owner struct {
    ID    string `json:"_id" bson:"_id"` //modify here
    Name  string `json:"name"`
    Email string `json:"email"`
    Phone string `json:"phone"`
}

type Project struct {
    ID          string `json:"_id" bson:"_id"` //modify here
    OwnerID     string `json:"ownerId"`
    Name        string `json:"name"`
    Description string `json:"description"`
    Status      Status `json:"status"`
}

//The remaining part of the code goes here
```

## Graphql mutations and queries

```graphql
mutation {
    createOwner (
        input: {name: "test", email: "test@gmail.com", phone: "9198398900"}
    ) {
        _id
        name
        email
        phone
    }
}
```

```graphql
mutation {
    createListing(
        input: {ownerId: "6405867e3e0a950708bd56f8", description: "New test listing", location: "Coppell, TX 75039", status: NOT_STARTED}
    ) {
        _id
        ownerId
        description
        location
        createdAt
        status
    }
}
```
