package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"tgd/cc"
	"tgd/dataloader"
	"tgd/graph/generated"
	"tgd/graph/model"
	"tgd/models"

	"gorm.io/gorm"
)

func (r *mutationResolver) CreateTodo(ctx context.Context, input model.NewTodo) (*models.Todo, error) {
	cc := cc.For(ctx)
	u := models.User{}
	if err := cc.DB().Where("name = ?", input.User).First(&u).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}
		u = models.User{
			Name: input.User,
		}
		if err := cc.DB().Create(&u).Error; err != nil {
			return nil, err
		}
	}
	t := models.Todo{
		Text:   input.Text,
		UserID: u.ID,
	}

	if err := cc.DB().Create(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *queryResolver) Todos(ctx context.Context) ([]*models.Todo, error) {
	cc := cc.For(ctx)
	ts := []*models.Todo{}
	if err := cc.DB().Find(&ts).Error; err != nil {
		return nil, err
	}
	return ts, nil
}

func (r *todoResolver) ID(ctx context.Context, obj *models.Todo) (int, error) {
	return int(obj.ID), nil
}

func (r *todoResolver) User(ctx context.Context, obj *models.Todo) (*models.User, error) {
	return dataloader.For(ctx).UserLoader.Load(obj.UserID)
}

func (r *userResolver) ID(ctx context.Context, obj *models.User) (int, error) {
	return int(obj.ID), nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Todo returns generated.TodoResolver implementation.
func (r *Resolver) Todo() generated.TodoResolver { return &todoResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type (
	mutationResolver struct{ *Resolver }
	queryResolver    struct{ *Resolver }
	todoResolver     struct{ *Resolver }
	userResolver     struct{ *Resolver }
)
