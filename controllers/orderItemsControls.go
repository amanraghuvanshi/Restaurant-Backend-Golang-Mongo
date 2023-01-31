package controllers

import (
	"context"
	"fmt"
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

type OrderItemPack struct {
	Table_id    *string
	Order_items []models.OrderItem
}

var orderItemsCollection *mongo.Collection = database.OpenCollection(database.Client, "orderItem")

// This function gets all the records
func GetOrderItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		res, err := orderCollection.Find(context.TODO(), bson.M{})

		defer cancel()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error encountered while fetching records"})
			return
		}
		var allOrdersItems []bson.M

		if err = res.All(ctx, &allOrdersItems); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error encountered while parsing all orderItems"})
		}

		c.JSON(http.StatusOK, allOrdersItems)
	}
}

// This function gets a order specific to the id provided
func GetOrderItemsbyID() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		orderItemsID := c.Param("order_item_id")

		var orderItems models.OrderItem

		err := orderItemsCollection.FindOne(ctx, bson.M{"order_item_id": orderItemsID}).Decode(&orderItems)

		defer cancel()

		if err != nil {
			msg := fmt.Sprintf("Error while getting order")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}

		c.JSON(http.StatusOK, orderItems)
	}
}

// This function provides the order specific to the order_id recieved
func GetOrderItemsbyOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("order_id")
		allOrderItems, err := ItemsByOrder(orderID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error while listing orders"})
			return
		}

		c.JSON(http.StatusOK, allOrderItems)
	}
}

func ItemsByOrder(id string) (OrderItems []primitive.M, err error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	// Here we will match the records based on the key provided
	// This will give us all the records, related to that orderId
	matchStage := bson.D{{"$match", bson.D{{"order_id", id}}}}

	// The lookup function is used for looking up the data, from a particular collection, here we are looking into food, from orderItemsCollection and we are using the food_id, as the localfield. and the table from which we are looking is food collection. And "as" means how the data will be represented
	lookupStage := bson.D{{"$lookup", bson.D{{"from", "food"}, {"localField", "food_id"}, {"foreignFeild", "food_id"}, {"as ", "food"}}}}
	// When we lookup for data, we get that in form of array, now in mongo we can't perform any operation while that data is in array form, so we need to unwind it, or decode it.
	unwindStage := bson.D{{"$unwind", bson.D{{"path", "$food"}, {"preserveNullandEmptyArrays", true}}}}
	//If we set this as false, then it removes all the null and empty arrays.

	// lookup for the orders
	lookupOrderStage := bson.D{{"$lookup", bson.D{{"from", "order"}, {"localField", "order_id"}, {"foreignField", "order_id"}, {"as", "order"}}}}
	unwindOrderStage := bson.D{{"$unwind", bson.D{{"path", "$order"}, {"preserveNullandEmptyArrays", true}}}}

	// Since already looked up in the orders collection, now we have the whole order object with us. And since we have unwinded it, we have the access to the field inside it.
	lookUpTableStage := bson.D{{"$lookup", bson.D{{"from", "tables"}, {"localField", "order.table_id"}, {"foreignField", "table_id"}, {"as", "table"}}}}
	unwindTableStage := bson.D{{"$unwind", bson.D{{"path", "$table"}, {"preserveNullandEmptyArrays", true}}}}

	// This stage manages the field that we will be sending to the next stage. After all these stages there will be a lot of data, and we might not use them all, so we need to sort them out
	projectStage := bson.D{
		{"$project", bson.D{
			{"id", 0}, // this means that ID is not going to the next stage
			{"amount", "$food.price"},
			{"total_count", 1},
			{"food_name", "$food.name"},
			{"food_image", "$food.food_image"},
			{"table_number", "$table_number"},
			{"table_id", "$table.table_id"},
			{"order_id", "$order.order_id"},
			{"price", "$food.price"},
			{"quantity", 1},
		}}}

	// It groups all the data based on the criteria provided
	groupStage := bson.D{{"$group", bson.D{{"_id", bson.D{{"order_id", "$order_id"}, {"table_id", "$table_id"}, {"table_number", "$table_number"}}}, {"payment_due", bson.D{{"$sum", "$amount"}}}, {"total_count", bson.D{{"$sum", 1}}}, {"order_items", bson.D{{"$sum", 1}}}}}}

	projectStage2 := bson.D{
		{"$project", bson.D{
			{"id", 0},
			{"payment_due", 1},
			{"total_count", 1},
			{"table_number", "$id.table_number"},
			{"order_items", 1},
		}}}

	res, err := orderItemsCollection.Aggregate(ctx, mongo.Pipeline{
		matchStage,
		lookupStage,
		unwindStage,
		lookupOrderStage,
		unwindOrderStage,
		lookUpTableStage,
		unwindTableStage,
		projectStage,
		groupStage,
		projectStage2})

	if err != nil {
		panic(err)
	}
	if err = res.All(ctx, &OrderItems); err != nil {
		panic(err)
	}
	defer cancel()
	return OrderItems, err
}

// This function creates a new entity of the orderItemCollection
func CreateOrderItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var orderItemsPack OrderItemPack
		var order models.Order

		if err := c.BindJSON(&orderItemsPack); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error while binding the data"})
			return
		}
		// creating the order date with its, timestamp
		order.Order_Date, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		// we will be using the table id for creating our order
		orderItemsTobeInserted := []interface{}{}
		order.Table_id = orderItemsPack.Table_id
		order_id := OrderItemsOrderCreator(order)

		for _, orderItem := range orderItemsPack.Order_items {
			orderItem.Order_id = order_id
			validationErr := validate.Struct(orderItem)

			if validationErr != nil {
				msg := fmt.Sprintf("Validation falied")
				c.JSON(http.StatusBadRequest, gin.H{"Error": msg})
				return
			}

			orderItem.ID = primitive.NewObjectID()
			orderItem.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			orderItem.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			orderItem.Order_item_id = orderItem.ID.Hex()
			var num = Tofixed(*orderItem.Unit_price, 2)
			orderItem.Unit_price = &num

			orderItemsTobeInserted = append(orderItemsTobeInserted, orderItem)

		}
		insertedOrders, err := orderItemsCollection.InsertMany(ctx, orderItemsTobeInserted)
		if err != nil {
			msg := fmt.Sprintf("Error:Failed to insert records")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}

		defer cancel()
		c.JSON(http.StatusOK, insertedOrders)
	}
}

// Function that updates the specified records
func UpdateOrderItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var orderItems models.OrderItem
		orderItemsID := c.Param("order_item_id")
		filter := bson.M{"order_item_id": orderItemsID}

		var updateObj primitive.D

		if orderItems.Unit_price != nil {
			updateObj = append(updateObj, bson.E{"unit_price", *&orderItems.Unit_price})
		}

		if orderItems.Quantity != nil {
			updateObj = append(updateObj, bson.E{"quantity", *&orderItems.Quantity})
		}

		if orderItems.Food_id != nil {
			updateObj = append(updateObj, bson.E{"food_id", *&orderItems.Food_id})
		}

		orderItems.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		updateObj = append(updateObj, bson.E{"updated_at", orderItems.Updated_at})

		defer cancel()

		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		res, err := orderItemsCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set", updateObj},
			},
			&opt,
		)

		if err != nil {
			msg := fmt.Sprintf("Error while updation")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		}
		c.JSON(http.StatusOK, res)
	}
}
