package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"restaurantms/database"
	"restaurantms/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var tablesCollection *mongo.Collection = database.OpenCollection(database.Client, "tables")

func GetTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		res, err := tablesCollection.Find(context.TODO(), bson.M{})

		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Couldn't find the data"})
			return

		}

		var allTables []bson.M

		if err = res.All(ctx, &allTables); err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, allTables)
	}
}

func GetTablebyID() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		tableID := c.Param("table_id")

		defer cancel()
		var tables models.Table

		err := tablesCollection.FindOne(ctx, bson.M{"table_id": tableID}).Decode(&tables)

		if err != nil {
			msg := fmt.Sprintf("Error while fetching the data, tables")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}

		c.JSON(http.StatusOK, tables)

	}
}

func CreateTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var tables models.Table

		if err := c.BindJSON(&tables); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		validateErr := validate.Struct(tables)
		if validateErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validateErr.Error()})
			return
		}

		tables.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		tables.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		tables.ID = primitive.NewObjectID()
		tables.Table_id = tables.ID.Hex()

		res, insertErr := tablesCollection.InsertOne(ctx, tables)
		if insertErr != nil {
			msg := fmt.Sprintf("ERROR: Failed to create")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, res)
	}
}

func UpdateTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var tables models.Table

		tablesID := c.Param("table_id")

		if err := c.BindJSON(&tables); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		var updateObj primitive.D
		if tables.Number_of_guests != nil {
			updateObj = append(updateObj, bson.E{"number_of_guests", tables.Number_of_guests})
		}

		if tables.Table_number != nil {
			updateObj = append(updateObj, bson.E{"table_number", tables.Table_number})
		}

		tables.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		upsert := true

		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		filter := bson.M{"table_id": tablesID}

		res, err := tablesCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set", updateObj},
			},
			&opt,
		)

		if err != nil {
			msg := fmt.Sprintf("Error, while updating records")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}

		defer cancel()
		c.JSON(http.StatusOK, res)
	}
}
