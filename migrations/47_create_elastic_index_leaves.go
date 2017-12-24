package migrations

import (
	"context"

	"github.com/Seklfreak/Robyul2/cache"
	"github.com/Seklfreak/Robyul2/models"
)

func m47_create_elastic_index_leaves() {
	if !cache.HasElastic() {
		return
	}

	elastic := cache.GetElastic()
	exists, err := elastic.IndexExists(models.ElasticIndexLeaves).Do(context.Background())
	if err != nil {
		panic(err)
	}
	if exists {
		return
	}

	leaveMapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"doc": map[string]interface{}{
				"properties": map[string]interface{}{
					"CreatedAt": map[string]interface{}{
						"type": "date",
					},
					"GuildID": map[string]interface{}{
						"type": "text",
					},
					"UserID": map[string]interface{}{
						"type": "text",
					},
				},
			},
		},
	}

	index, err := elastic.CreateIndex(models.ElasticIndexLeaves).BodyJson(leaveMapping).Do(context.Background())
	if err != nil {
		panic(err)
	}
	if !index.Acknowledged {
		cache.GetLogger().WithField("module", "migrations").Error("ElasticSearch index not acknowledged")
	}
}
