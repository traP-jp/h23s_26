package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type (
	// users table
	User struct {
		ID              string      `db:"id"` // primary key
		AchieveMissions []uuid.UUID `db:"-"`
	}

	// user_mission_relations table
	UserMissionRelation struct {
		ID        uuid.UUID `db:"id"`         // primary key
		UserID    string    `db:"user_id"`    // foreign key
		MissionID string    `db:"mission_id"` // foreign key
	}
)

func (r *Repository) GetUsers(ctx context.Context) ([]*User, error) {
	users := make([]*User, 0)
	if err := r.db.SelectContext(ctx, &users, "SELECT * FROM users"); err != nil {
		return nil, fmt.Errorf("get users from db: %w", err)
	}

	achieveMissons := make([]*UserMissionRelation, 0)
	if err := r.db.SelectContext(ctx, &achieveMissons, "SELECT * FROM user_mission_relations"); err != nil {
		return nil, fmt.Errorf("get user_mission_relations from db: %w", err)
	}

	for _, user := range users {
		for _, achieveMission := range achieveMissons {
			if user.ID == achieveMission.UserID {
				user.AchieveMissions = append(user.AchieveMissions, achieveMission.ID)
			}
		}
	}

	return users, nil
}