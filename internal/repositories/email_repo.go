package repositories

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type EmailRepository struct {
	EmailCollection *mongo.Collection
}

func NewEmailRepository(client *mongo.Client) *EmailRepository {
	return &EmailRepository{
		EmailCollection: client.Database("auth").Collection("email"),
	}
}

func (r *EmailRepository) CreateEmail(email string, otp string) error {

	_, err := r.EmailCollection.InsertOne(context.TODO(), bson.D{{Key: "email", Value: email}, {Key: "otp", Value: otp}})
	return err

}
