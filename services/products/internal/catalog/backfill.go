package catalog

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type GradingBackfillResult struct {
	Category  int64
	Condition int64
}

func (s *Store) BackfillGrading(ctx context.Context, apply bool) (GradingBackfillResult, error) {
	result := GradingBackfillResult{}
	for _, field := range []struct {
		name, value string
		count       *int64
	}{
		{"category", "accessory", &result.Category},
		{"condition", "good", &result.Condition},
	} {
		filter := bson.M{field.name: bson.M{"$exists": false}}
		if apply {
			updated, err := s.listings.UpdateMany(ctx, filter, bson.M{"$set": bson.M{field.name: field.value}})
			if err != nil {
				return result, err
			}
			*field.count = updated.ModifiedCount
		} else {
			count, err := s.listings.CountDocuments(ctx, filter)
			if err != nil {
				return result, err
			}
			*field.count = count
		}
	}
	return result, nil
}
