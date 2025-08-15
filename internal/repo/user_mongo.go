package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/luxixing/fx-gin-scaffold/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// userMongoRepository implements UserRepository for MongoDB
type userMongoRepository struct {
	collection *mongo.Collection
}

// NewUserMongoRepository creates a new MongoDB-based user repository
func NewUserMongoRepository(db *mongo.Database) domain.UserRepository {
	collection := db.Collection("users")
	
	// Create indexes
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		// Email unique index
		emailIndex := mongo.IndexModel{
			Keys:    bson.M{"email": 1},
			Options: options.Index().SetUnique(true),
		}
		
		_, err := collection.Indexes().CreateOne(ctx, emailIndex)
		if err != nil {
			// Log error but don't fail - indexes might already exist
			fmt.Printf("Warning: Failed to create email index: %v\n", err)
		}
	}()
	
	return &userMongoRepository{
		collection: collection,
	}
}

// mongoUser represents the User model for MongoDB with proper ID handling
type mongoUser struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Email     string             `bson:"email"`
	Password  string             `bson:"password"`
	Name      string             `bson:"name"`
	Role      string             `bson:"role"`
	Active    bool               `bson:"active"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

// toDomainUser converts mongoUser to domain.User
func (m *mongoUser) toDomainUser() *domain.User {
	return &domain.User{
		ID:        uint(m.ID.Timestamp().Unix()), // Use timestamp as ID for compatibility
		Email:     m.Email,
		Password:  m.Password,
		Name:      m.Name,
		Role:      m.Role,
		Active:    m.Active,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// fromDomainUser converts domain.User to mongoUser
func fromDomainUser(user *domain.User) *mongoUser {
	m := &mongoUser{
		Email:     user.Email,
		Password:  user.Password,
		Name:      user.Name,
		Role:      user.Role,
		Active:    user.Active,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	
	// If ID is provided, try to create ObjectID from it
	if user.ID != 0 {
		// For simplicity, we'll generate a new ObjectID for updates
		// In a real application, you might want to store the ObjectID separately
		m.ID = primitive.NewObjectID()
	}
	
	return m
}

// Create creates a new user
func (r *userMongoRepository) Create(ctx context.Context, user *domain.User) error {
	mongoUser := fromDomainUser(user)
	mongoUser.CreatedAt = time.Now()
	mongoUser.UpdatedAt = time.Now()
	
	result, err := r.collection.InsertOne(ctx, mongoUser)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return domain.ErrUserExists
		}
		return domain.WrapError(err, domain.ErrCodeDatabase, "Failed to create user")
	}
	
	// Set the generated ID back to the user
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		user.ID = uint(oid.Timestamp().Unix())
		user.CreatedAt = mongoUser.CreatedAt
		user.UpdatedAt = mongoUser.UpdatedAt
	}
	
	return nil
}

// GetByID retrieves a user by ID
func (r *userMongoRepository) GetByID(ctx context.Context, id uint) (*domain.User, error) {
	// For MongoDB, we need to find by email or another field since we don't store the uint ID
	// This is a limitation of this example - in practice, you'd store the ID differently
	return nil, domain.NewError(domain.ErrCodeNotFound, "GetByID not implemented for MongoDB - use GetByEmail")
}

// GetByEmail retrieves a user by email
func (r *userMongoRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var mongoUser mongoUser
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&mongoUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrUserNotFound
		}
		return nil, domain.WrapError(err, domain.ErrCodeDatabase, "Failed to get user by email")
	}
	
	return mongoUser.toDomainUser(), nil
}

// Update updates an existing user
func (r *userMongoRepository) Update(ctx context.Context, user *domain.User) error {
	mongoUser := fromDomainUser(user)
	mongoUser.UpdatedAt = time.Now()
	
	update := bson.M{
		"$set": bson.M{
			"name":       mongoUser.Name,
			"role":       mongoUser.Role,
			"active":     mongoUser.Active,
			"updated_at": mongoUser.UpdatedAt,
		},
	}
	
	result, err := r.collection.UpdateOne(ctx, bson.M{"email": user.Email}, update)
	if err != nil {
		return domain.WrapError(err, domain.ErrCodeDatabase, "Failed to update user")
	}
	
	if result.MatchedCount == 0 {
		return domain.ErrUserNotFound
	}
	
	user.UpdatedAt = mongoUser.UpdatedAt
	return nil
}

// Delete soft deletes a user (marks as inactive)
func (r *userMongoRepository) Delete(ctx context.Context, id uint) error {
	// Since we don't have a direct way to find by uint ID, this is a limitation
	// In practice, you'd store the relationship differently
	return domain.NewError(domain.ErrCodeNotFound, "Delete by ID not implemented for MongoDB")
}

// List retrieves users with pagination
func (r *userMongoRepository) List(ctx context.Context, offset, limit int) ([]*domain.User, int64, error) {
	// Count total documents
	total, err := r.collection.CountDocuments(ctx, bson.M{"active": true})
	if err != nil {
		return nil, 0, domain.WrapError(err, domain.ErrCodeDatabase, "Failed to count users")
	}
	
	// Find documents with pagination
	findOptions := options.Find()
	findOptions.SetSkip(int64(offset))
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.M{"created_at": -1})
	
	cursor, err := r.collection.Find(ctx, bson.M{"active": true}, findOptions)
	if err != nil {
		return nil, 0, domain.WrapError(err, domain.ErrCodeDatabase, "Failed to list users")
	}
	defer cursor.Close(ctx)
	
	var mongoUsers []mongoUser
	if err := cursor.All(ctx, &mongoUsers); err != nil {
		return nil, 0, domain.WrapError(err, domain.ErrCodeDatabase, "Failed to decode users")
	}
	
	// Convert to domain users
	users := make([]*domain.User, len(mongoUsers))
	for i, mu := range mongoUsers {
		users[i] = mu.toDomainUser()
	}
	
	return users, total, nil
}

// Search searches users by name or email
func (r *userMongoRepository) Search(ctx context.Context, query string, offset, limit int) ([]*domain.User, int64, error) {
	// Create regex pattern for case-insensitive search
	pattern := primitive.Regex{Pattern: query, Options: "i"}
	filter := bson.M{
		"active": true,
		"$or": []bson.M{
			{"name": pattern},
			{"email": pattern},
		},
	}
	
	// Count total documents
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, domain.WrapError(err, domain.ErrCodeDatabase, "Failed to count search results")
	}
	
	// Find documents with pagination
	findOptions := options.Find()
	findOptions.SetSkip(int64(offset))
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.M{"created_at": -1})
	
	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, domain.WrapError(err, domain.ErrCodeDatabase, "Failed to search users")
	}
	defer cursor.Close(ctx)
	
	var mongoUsers []mongoUser
	if err := cursor.All(ctx, &mongoUsers); err != nil {
		return nil, 0, domain.WrapError(err, domain.ErrCodeDatabase, "Failed to decode search results")
	}
	
	// Convert to domain users
	users := make([]*domain.User, len(mongoUsers))
	for i, mu := range mongoUsers {
		users[i] = mu.toDomainUser()
	}
	
	return users, total, nil
}