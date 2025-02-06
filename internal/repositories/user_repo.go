package repositories

import (
	"context"

	"github.com/ronak-89/simple-user-auth/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	UserCollection  *mongo.Collection
	EmailCollection *mongo.Collection
}

func NewUserRepository(client *mongo.Client) *UserRepository {
	return &UserRepository{
		UserCollection:  client.Database("auth").Collection("users"),
		EmailCollection: client.Database("auth").Collection("email"),
	}
}

func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	filter := bson.D{{Key: "email", Value: email}}

	err := r.UserCollection.FindOne(context.TODO(), filter).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) CreatUser(user *models.User) error {
	_, err := r.UserCollection.InsertOne(context.TODO(), user)
	return err
}
