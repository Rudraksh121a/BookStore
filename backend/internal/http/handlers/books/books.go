package books

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Rudraksh121a/BookStore/internal/config"
	"github.com/Rudraksh121a/BookStore/internal/storage/mongodb"
	"github.com/Rudraksh121a/BookStore/internal/types"
	jwtutil "github.com/Rudraksh121a/BookStore/internal/utils/jwt"
	"github.com/Rudraksh121a/BookStore/internal/utils/response"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func New() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Server is Healthy"))
	}
}

func Register(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user types.User

		// Decode request body
		err := json.NewDecoder(r.Body).Decode(&user)
		if errors.Is(err, io.EOF) {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body")))
			return
		}

		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		// Validate user data
		if err := validator.New().Struct(user); err != nil {
			validateError := err.(validator.ValidationErrors)
			response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateError))
			return
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(fmt.Errorf("failed to hash password")))
			return
		}
		user.Password = string(hashedPassword)

		// Connect to database
		mongoDB, err := mongodb.NewMongo(cfg)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(fmt.Errorf("database connection failed: %v", err)))
			return
		}
		defer mongoDB.Client.Disconnect(context.TODO())

		// Initialize database indexes
		if err := mongoDB.Init(); err != nil {
			// Log but continue - index might already exist
		}

		// Create user
		err = mongoDB.CreateUser(&user)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		// Send success response
		response.WriteJson(w, http.StatusCreated, map[string]string{
			"message": "user registered successfully",
			"email":   user.Email,
		})
	}
}

func Login(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var loginReq struct {
			Email    string `json:"email" validate:"required,email"`
			Password string `json:"password" validate:"required"`
		}

		// Decode request body
		err := json.NewDecoder(r.Body).Decode(&loginReq)
		if errors.Is(err, io.EOF) {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body")))
			return
		}

		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		// Validate login data
		if err := validator.New().Struct(loginReq); err != nil {
			validateError := err.(validator.ValidationErrors)
			response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateError))
			return
		}

		// Connect to database
		mongoDB, err := mongodb.NewMongo(cfg)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(fmt.Errorf("database connection failed")))
			return
		}
		defer mongoDB.Client.Disconnect(context.TODO())

		// Get user by email
		user, err := mongoDB.GetUserByEmail(loginReq.Email)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				response.WriteJson(w, http.StatusUnauthorized, response.GeneralError(fmt.Errorf("invalid email or password")))
				return
			}
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(fmt.Errorf("failed to retrieve user")))
			return
		}

		// Verify password
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password))
		if err != nil {
			response.WriteJson(w, http.StatusUnauthorized, response.GeneralError(fmt.Errorf("invalid email or password")))
			return
		}

		// Generate JWT token
		token, err := jwtutil.GenerateJWT(user.ID.Hex(), cfg)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(fmt.Errorf("failed to generate token")))
			return
		}

		// Send success response with token
		response.WriteJson(w, http.StatusOK, map[string]interface{}{
			"message": "login successful",
			"token":   token,
			"user": map[string]string{
				"id":    user.ID.Hex(),
				"email": user.Email,
				"name":  user.Name,
			},
		})
	}
}

func CreateBook(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.WriteJson(w, http.StatusUnauthorized, response.GeneralError(fmt.Errorf("authorization header required")))
			return
		}

		// Extract token (format: "Bearer <token>")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.WriteJson(w, http.StatusUnauthorized, response.GeneralError(fmt.Errorf("invalid authorization header format")))
			return
		}
		tokenString := parts[1]

		// Verify JWT token
		token, err := jwtutil.VerifyJWT(tokenString, cfg)
		if err != nil {
			response.WriteJson(w, http.StatusUnauthorized, response.GeneralError(fmt.Errorf("invalid or expired token")))
			return
		}

		// Extract user ID from token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			response.WriteJson(w, http.StatusUnauthorized, response.GeneralError(fmt.Errorf("invalid token claims")))
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			response.WriteJson(w, http.StatusUnauthorized, response.GeneralError(fmt.Errorf("user_id not found in token")))
			return
		}

		// Decode book data
		var book types.Book
		err = json.NewDecoder(r.Body).Decode(&book)
		if errors.Is(err, io.EOF) {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body")))
			return
		}

		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		// Validate book data
		if err := validator.New().Struct(book); err != nil {
			validateError := err.(validator.ValidationErrors)
			response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateError))
			return
		}

		// Set metadata
		now := time.Now().Format(time.RFC3339)
		book.CreatedBy = userID
		book.CreatedAt = now
		book.UpdatedAt = now

		// Connect to database
		mongoDB, err := mongodb.NewMongo(cfg)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(fmt.Errorf("database connection failed")))
			return
		}
		defer mongoDB.Client.Disconnect(context.TODO())

		// Create book
		err = mongoDB.CreateBook(&book)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(fmt.Errorf("failed to create book")))
			return
		}

		// Send success response
		response.WriteJson(w, http.StatusCreated, map[string]interface{}{
			"message": "book created successfully",
			"book": map[string]interface{}{
				"id":         book.ID.Hex(),
				"title":      book.Title,
				"author":     book.Author,
				"genre":      book.Genre,
				"created_at": book.CreatedAt,
			},
		})
	}
}
