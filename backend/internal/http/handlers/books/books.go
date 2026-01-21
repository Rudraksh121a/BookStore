package books

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/Rudraksh121a/BookStore/internal/config"
	"github.com/Rudraksh121a/BookStore/internal/storage/mongodb"
	"github.com/Rudraksh121a/BookStore/internal/types"
	"github.com/Rudraksh121a/BookStore/internal/utils/response"
	"github.com/go-playground/validator/v10"
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
