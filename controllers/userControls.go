package controllers

import (
	"context"
	"log"
	"net/http"
	"restaurantms/database"
	"restaurantms/helpers"
	"restaurantms/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var usersCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

func GetUserbyID() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User
		userID := c.Param("user_id")
		err := usersCollection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&user)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while fetching user records"})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		recordsPerPage, err := strconv.Atoi(c.Query("recordsPerPage"))
		if err != nil || recordsPerPage < 1 {
			recordsPerPage = 10
		}

		pages, err1 := strconv.Atoi(c.Query("page"))
		if err1 != nil || pages < 1 {
			pages = 1
		}

		startIndexes := (pages - 1) * recordsPerPage
		startIndexes, err = strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{{"$match", bson.D{{}}}}
		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{"_id", 0},
				{"total_count", 1},
				{"user_items", bson.D{{"$slice", []interface{}{"$data", startIndexes, recordsPerPage}}}},
			}}}

		res, err := usersCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage, projectStage})

		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while listing user items"})
		}
		var allUsers []bson.M
		if err = res.All(ctx, &allUsers); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allUsers)
	}
}

func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var user models.User

		// Convert JSON coming from the postman to something that go could understand
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		//Validating data based on user struct
		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}
		// checking if the email has already been used by another user
		count, err := usersCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusContinue, gin.H{"Error": "Conflict, Email has already been used"})
			return
		}

		// hashing the password
		pass := HashPass(*user.Password)
		user.Password = &pass
		// also checking if phone number has already been used
		count, err = usersCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusContinue, gin.H{"Error": "Conflict, Number has already been used"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Email or phone already exists"})
			return
		}

		// some more details for user object, created_at, updated_at, ID
		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()

		// generate token and refresh token(generate all token function)
		token, refreshToken, _ := helpers.GenerateAllToken(*user.Email, *user.First_name, *user.Last_name, *&user.User_id)
		user.Token = &token
		user.Refresh_Token = &refreshToken

		//if all ok, we insert user in the database

		res, insertErr := usersCollection.InsertOne(ctx, user)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Can't create the user"})
			return
		}
		defer cancel()
		// return user with status ok
		c.JSON(http.StatusOK, res)
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		var foundUser models.User

		// convert the login data JSON data into golang format
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
			return
		}
		// find a user with relevant email and check if user exists
		err := usersCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "User not found"})
			return
		}
		//	then validate password
		passIsvalid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		if passIsvalid != true {
			c.JSON(http.StatusUnauthorized, gin.H{"Error": msg})
			return
		}
		// 	if ok, then generate tokens
		token, refreshToken, _ := helpers.GenerateAllToken(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, *&foundUser.User_id)

		// update token - token and refresh token
		helpers.UpdateAllTokens(token, refreshToken, foundUser.User_id)

		// return status ok, if successful
		c.JSON(http.StatusOK, foundUser)
	}
}

func HashPass(password string) string {
	// This function will be used in the signup while creating user
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword string, providePass string) (bool, string) {

	// this function will be used in login to verify if the credentials are correct or not
	check := true
	msg := ""
	if err := bcrypt.CompareHashAndPassword([]byte(providePass), []byte(userPassword)); err != nil {
		msg = "Invalid Credentials"
		check = false
	}
	return check, msg
}
