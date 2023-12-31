package db

import (
	"context"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"
	"fmt"

	"github.com/gertd/go-pluralize"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func Client() *mongo.Database {
	if client == nil {
		uri := os.Getenv("mongodb")

		serverAPI := options.ServerAPI(options.ServerAPIVersion1)
		opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)
		// Create a new client and connect to the server
		c, err := mongo.Connect(context.TODO(), opts)
		if err != nil {
			panic(err)
		}

		client = c
	}

	return client.Database("neodeliver")
}

func Close() {
	if client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}

		client = nil
	}
}

// ---

func Find(ctx context.Context, o interface{}, filter interface{}, opts ...*options.FindOneOptions) (interface{}, error) {
	c := Client()
	result := o
	err := c.Collection(CollectionName(o)).FindOne(ctx, filter, opts...).Decode(&result)
	return result, err
}

func Count(ctx context.Context, o interface{}, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	c := Client()
	count, err := c.Collection(CollectionName(o)).CountDocuments(ctx, filter, opts...)
	if err != nil {
		return 0, err
	}
	return count, err
}

func Save(ctx context.Context, o interface{}) (*mongo.InsertOneResult, error) {
	c := Client().Collection(CollectionName(o))
	insertResult, err := c.InsertOne(ctx, o)
	return insertResult, err
}

func Update(ctx context.Context, o interface{}, filter interface{}, update interface{}) error {
	c := Client().Collection(CollectionName(o))
	after := options.After
	upsert := true

	// TODO filter out nil fields
	fmt.Println(update)
	update = FilterNilFields(update)
	fmt.Println(update)

	res := c.FindOneAndUpdate(ctx, filter, map[string]interface{}{"$set": update}, &options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
		Upsert:         &upsert,
	})

	if err := res.Err(); err != nil {
		return err
	}

	return res.Decode(o)
}

func Delete(ctx context.Context, o interface{}, filter interface{}) error {
	c := Client().Collection(CollectionName(o))
	_, err := c.DeleteOne(ctx, filter)
	return err
}

func CollectionName(o interface{}) string {
	name := ToSnakeCase(reflect.TypeOf(o).Elem().Name())
	return pluralize.NewClient().Plural(name)
}

// ---

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func FilterNilFields(obj interface{}) interface{} {
	objValue := reflect.ValueOf(obj)
	objType := objValue.Type()
	fmt.Println(objValue, objType)
	
	// Create a new instance of the same type as the input object
	result := reflect.New(objType).Elem()

	// Iterate over the fields of the object
	for i := 0; i < objValue.NumField(); i++ {
		fieldValue := objValue.Field(i)

		// Check if the field value is nil
		if fieldValue.IsNil() {
			continue // Skip nil fields
		}

		// Set the field value in the result object
		result.Field(i).Set(fieldValue)
	}

	return result.Interface()
}
