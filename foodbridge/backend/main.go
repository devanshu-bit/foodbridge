package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// Models – proper BSON tags added
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Email     string             `bson:"email" json:"email"`
	Password  string             `bson:"password,omitempty" json:"password,omitempty"`
	Role      string             `bson:"role" json:"role"`
	Phone     string             `bson:"phone" json:"phone"`
	Address   string             `bson:"address" json:"address"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

type Donation struct {
	ID              primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	RestaurantID    primitive.ObjectID  `bson:"restaurant_id" json:"restaurant_id"`
	RestaurantName  string              `bson:"restaurant_name" json:"restaurant_name"`
	RestaurantPhone string              `bson:"restaurant_phone,omitempty" json:"restaurant_phone,omitempty"`
	FoodName        string              `bson:"food_name" json:"food_name"`
	Quantity        int                 `bson:"quantity" json:"quantity"`
	PickupTime      time.Time           `bson:"pickup_time" json:"pickup_time"`
	PickupLocation  string              `bson:"pickup_location" json:"pickup_location"`
	Status          string              `bson:"status" json:"status"`
	NGOID           *primitive.ObjectID `bson:"ngo_id,omitempty" json:"ngo_id,omitempty"`
	NGOName         string              `bson:"ngo_name,omitempty" json:"ngo_name,omitempty"`
	NGOPhone        string              `bson:"ngo_phone,omitempty" json:"ngo_phone,omitempty"`
	VolunteerID     *primitive.ObjectID `bson:"volunteer_id,omitempty" json:"volunteer_id,omitempty"`
	VolunteerName   string              `bson:"volunteer_name,omitempty" json:"volunteer_name,omitempty"`
	VolunteerPhone  string              `bson:"volunteer_phone,omitempty" json:"volunteer_phone,omitempty"`
	CreatedAt       time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time           `bson:"updated_at" json:"updated_at"`
}

var (
	userCollection     *mongo.Collection
	donationCollection *mongo.Collection
	jwtSecret          = []byte("foodbridge-secret-2024")
)

type Claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	UserName string `json:"user_name"`
	jwt.RegisteredClaims
}

func connectDB() *mongo.Client {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal("MongoDB connection error:", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("MongoDB ping error:", err)
	}
	fmt.Println("Connected to MongoDB!")
	return client
}

func generateToken(userID, email, role, userName string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Email:    email,
		Role:     role,
		UserName: userName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
			c.Abort()
			return
		}
		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("user_name", claims.UserName)
		c.Next()
	}
}

func adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// Helper to get user phone
func getUserPhone(userID primitive.ObjectID) string {
	var user User
	err := userCollection.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return ""
	}
	return user.Phone
}

// Handlers
func register(c *gin.Context) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Role     string `json:"role"`
		Phone    string `json:"phone"`
		Address  string `json:"address"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing User
	err := userCollection.FindOne(context.Background(), bson.M{"email": input.Email}).Decode(&existing)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	user := User{
		ID:        primitive.NewObjectID(),
		Name:      input.Name,
		Email:     input.Email,
		Password:  string(hashedPassword),
		Role:      input.Role,
		Phone:     input.Phone,
		Address:   input.Address,
		CreatedAt: time.Now(),
	}

	_, err = userCollection.InsertOne(context.Background(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	token, _ := generateToken(user.ID.Hex(), user.Email, user.Role, user.Name)
	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"user": gin.H{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"role":       user.Role,
			"phone":      user.Phone,
			"address":    user.Address,
			"created_at": user.CreatedAt,
		},
	})
}

func login(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user User
	err := userCollection.FindOne(context.Background(), bson.M{"email": input.Email}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	token, _ := generateToken(user.ID.Hex(), user.Email, user.Role, user.Name)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"role":       user.Role,
			"phone":      user.Phone,
			"address":    user.Address,
			"created_at": user.CreatedAt,
		},
	})
}

func createDonation(c *gin.Context) {
	var input struct {
		FoodName       string `json:"food_name"`
		Quantity       int    `json:"quantity"`
		PickupTime     string `json:"pickup_time"`
		PickupLocation string `json:"pickup_location"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	userName := c.GetString("user_name")
	restaurantID, _ := primitive.ObjectIDFromHex(userID)
	pickupTime, _ := time.Parse(time.RFC3339, input.PickupTime)

	// Fetch restaurant phone
	restaurantPhone := getUserPhone(restaurantID)

	donation := Donation{
		ID:              primitive.NewObjectID(),
		RestaurantID:    restaurantID,
		RestaurantName:  userName,
		RestaurantPhone: restaurantPhone,
		FoodName:        input.FoodName,
		Quantity:        input.Quantity,
		PickupTime:      pickupTime,
		PickupLocation:  input.PickupLocation,
		Status:          "available",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	_, err := donationCollection.InsertOne(context.Background(), donation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create donation"})
		return
	}
	fmt.Println("Donation created by", userName)
	c.JSON(http.StatusCreated, donation)
}

func getAvailableDonations(c *gin.Context) {
	cursor, _ := donationCollection.Find(context.Background(), bson.M{"status": "available"})
	var donations []Donation
	cursor.All(context.Background(), &donations)
	if donations == nil {
		donations = []Donation{}
	}
	c.JSON(http.StatusOK, donations)
}

func getAcceptedDonations(c *gin.Context) {
	cursor, _ := donationCollection.Find(context.Background(), bson.M{"status": "accepted"})
	var donations []Donation
	cursor.All(context.Background(), &donations)
	if donations == nil {
		donations = []Donation{}
	}
	c.JSON(http.StatusOK, donations)
}

func acceptDonation(c *gin.Context) {
	donationID := c.Param("id")
	ngoID := c.GetString("user_id")
	ngoName := c.GetString("user_name")

	objID, _ := primitive.ObjectIDFromHex(donationID)
	ngoObjID, _ := primitive.ObjectIDFromHex(ngoID)

	ngoPhone := getUserPhone(ngoObjID)

	_, err := donationCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID, "status": "available"},
		bson.M{"$set": bson.M{
			"status":     "accepted",
			"ngo_id":     ngoObjID,
			"ngo_name":   ngoName,
			"ngo_phone":  ngoPhone,
			"updated_at": time.Now(),
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept donation"})
		return
	}
	var donation Donation
	donationCollection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&donation)
	fmt.Println("Donation accepted by NGO:", ngoName)
	c.JSON(http.StatusOK, donation)
}

func assignVolunteer(c *gin.Context) {
	donationID := c.Param("id")
	volunteerID := c.GetString("user_id")
	volunteerName := c.GetString("user_name")

	objID, _ := primitive.ObjectIDFromHex(donationID)
	volObjID, _ := primitive.ObjectIDFromHex(volunteerID)

	volunteerPhone := getUserPhone(volObjID)

	_, err := donationCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID, "status": "accepted"},
		bson.M{"$set": bson.M{
			"status":          "assigned",
			"volunteer_id":    volObjID,
			"volunteer_name":  volunteerName,
			"volunteer_phone": volunteerPhone,
			"updated_at":      time.Now(),
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign volunteer"})
		return
	}
	var donation Donation
	donationCollection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&donation)
	fmt.Println("Volunteer assigned:", volunteerName)
	c.JSON(http.StatusOK, donation)
}

func updateDonationStatus(c *gin.Context) {
	donationID := c.Param("id")
	var input struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	objID, _ := primitive.ObjectIDFromHex(donationID)
	donationCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"status": input.Status, "updated_at": time.Now()}},
	)
	var donation Donation
	donationCollection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&donation)
	c.JSON(http.StatusOK, donation)
}

func getUserDonations(c *gin.Context) {
	userID := c.GetString("user_id")
	role := c.GetString("role")
	objID, _ := primitive.ObjectIDFromHex(userID)
	var filter bson.M
	switch role {
	case "restaurant":
		filter = bson.M{"restaurant_id": objID}
	case "ngo":
		filter = bson.M{"ngo_id": objID}
	case "volunteer":
		filter = bson.M{"volunteer_id": objID}
	}
	cursor, _ := donationCollection.Find(context.Background(), filter)
	var donations []Donation
	cursor.All(context.Background(), &donations)
	if donations == nil {
		donations = []Donation{}
	}
	fmt.Printf("User %s (%s) fetched %d donations\n", userID, role, len(donations))
	c.JSON(http.StatusOK, donations)
}

func getAllUsers(c *gin.Context) {
	cursor, _ := userCollection.Find(context.Background(), bson.M{})
	var users []User
	cursor.All(context.Background(), &users)
	for i := range users {
		users[i].Password = ""
	}
	if users == nil {
		users = []User{}
	}
	c.JSON(http.StatusOK, users)
}

func getAllDonationsAdmin(c *gin.Context) {
	cursor, _ := donationCollection.Find(context.Background(), bson.M{})
	var donations []Donation
	cursor.All(context.Background(), &donations)
	if donations == nil {
		donations = []Donation{}
	}
	c.JSON(http.StatusOK, donations)
}

func updateUserAdmin(c *gin.Context) {
	userID := c.Param("id")
	objID, _ := primitive.ObjectIDFromHex(userID)
	var input struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Phone   string `json:"phone"`
		Address string `json:"address"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"name": input.Name, "email": input.Email, "phone": input.Phone, "address": input.Address}},
	)
	c.JSON(http.StatusOK, gin.H{"message": "User updated"})
}

func deleteUserAdmin(c *gin.Context) {
	userID := c.Param("id")
	objID, _ := primitive.ObjectIDFromHex(userID)
	userCollection.DeleteOne(context.Background(), bson.M{"_id": objID})
	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

func deleteDonationAdmin(c *gin.Context) {
	donationID := c.Param("id")
	objID, _ := primitive.ObjectIDFromHex(donationID)
	donationCollection.DeleteOne(context.Background(), bson.M{"_id": objID})
	c.JSON(http.StatusOK, gin.H{"message": "Donation deleted"})
}

func createAdminUser() {
	var existing User
	err := userCollection.FindOne(context.Background(), bson.M{"email": "admin@foodbridge.com"}).Decode(&existing)
	if err == nil {
		return
	}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	admin := User{
		ID:        primitive.NewObjectID(),
		Name:      "Admin",
		Email:     "admin@foodbridge.com",
		Password:  string(hashedPassword),
		Role:      "admin",
		Phone:     "0000000000",
		Address:   "FoodBridge HQ",
		CreatedAt: time.Now(),
	}
	userCollection.InsertOne(context.Background(), admin)
	fmt.Println("Default admin created: admin@foodbridge.com / admin123")
}

func main() {
	client := connectDB()
	db := client.Database("foodbridge")
	userCollection = db.Collection("users")
	donationCollection = db.Collection("donations")

	createAdminUser()

	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"*"}
	r.Use(cors.New(config))

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "FoodBridge API is running"})
	})

	r.POST("/api/auth/register", register)
	r.POST("/api/auth/login", login)

	api := r.Group("/api")
	api.Use(authMiddleware())
	{
		api.POST("/donations", createDonation)
		api.GET("/donations/available", getAvailableDonations)
		api.GET("/donations/accepted", getAcceptedDonations)
		api.GET("/donations/user", getUserDonations)
		api.PUT("/donations/:id/accept", acceptDonation)
		api.PUT("/donations/:id/assign", assignVolunteer)
		api.PUT("/donations/:id/status", updateDonationStatus)

		admin := api.Group("/admin")
		admin.Use(adminMiddleware())
		{
			admin.GET("/users", getAllUsers)
			admin.GET("/donations", getAllDonationsAdmin)
			admin.PUT("/users/:id", updateUserAdmin)
			admin.DELETE("/users/:id", deleteUserAdmin)
			admin.DELETE("/donations/:id", deleteDonationAdmin)
		}
	}

	port := "8080"
	fmt.Println("Server running on http://localhost:" + port)
	r.Run(":" + port)
}