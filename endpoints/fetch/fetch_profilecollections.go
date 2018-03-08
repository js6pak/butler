package fetch

import (
	"github.com/go-errors/errors"
	"github.com/itchio/butler/buse"
	"github.com/itchio/butler/buse/messages"
	"github.com/itchio/butler/database/hades"
	"github.com/itchio/butler/database/models"
	"github.com/itchio/go-itchio"
)

func FetchProfileCollections(rc *buse.RequestContext, params *buse.FetchProfileCollectionsParams) (*buse.FetchProfileCollectionsResult, error) {
	profile, client := rc.ProfileClient(params.ProfileID)

	c := HadesContext(rc)

	sendDBCollections := func() error {
		err := c.Preload(rc.DB(), &hades.PreloadParams{
			Record: profile,
			Fields: []hades.PreloadField{
				{Name: "ProfileCollections", OrderBy: `"position" ASC`},
				{Name: "ProfileCollections.Collection"},
			},
		})
		if err != nil {
			return errors.Wrap(err, 0)
		}

		profileCollections := profile.ProfileCollections

		var collectionIDs []int64
		collectionsByIDs := make(map[int64]*itchio.Collection)
		for _, pc := range profileCollections {
			c := pc.Collection
			collectionIDs = append(collectionIDs, c.ID)
			collectionsByIDs[c.ID] = c
		}

		var rows []struct {
			itchio.CollectionGame
			itchio.Game
		}
		err = rc.DB().Raw(`
			SELECT collection_games.*, games.*
			FROM collections
			JOIN collection_games ON collection_games.collection_id = collections.id
			JOIN games ON games.id = collection_games.game_id
			WHERE collections.id IN (?)
			AND collection_games.game_id IN (
				SELECT game_id
				FROM collection_games
				WHERE collection_games.collection_id = collections.id
				ORDER BY "position" ASC
				LIMIT 8
			)
		`, collectionIDs).Scan(&rows).Error
		if err != nil {
			return errors.Wrap(err, 0)
		}

		for _, row := range rows {
			c := collectionsByIDs[row.CollectionGame.CollectionID]
			cg := row.CollectionGame
			game := row.Game
			cg.Game = &game
			c.CollectionGames = append(c.CollectionGames, &cg)
		}

		if len(profileCollections) > 0 {
			yn := &buse.FetchProfileCollectionsYieldNotification{}
			yn.Offset = 0
			yn.Total = int64(len(profileCollections))

			for _, pc := range profileCollections {
				yn.Items = append(yn.Items, pc.Collection)
			}

			err = messages.FetchProfileCollectionsYield.Notify(rc, yn)
			if err != nil {
				return errors.Wrap(err, 0)
			}
		}
		return nil
	}

	err := sendDBCollections()
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	collRes, err := client.ListMyCollections()
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	profile.ProfileCollections = nil
	for i, c := range collRes.Collections {
		for j, g := range c.Games {
			c.CollectionGames = append(c.CollectionGames, &itchio.CollectionGame{
				Position: int64(j),
				Game:     g,
			})
		}
		c.Games = nil

		profile.ProfileCollections = append(profile.ProfileCollections, &models.ProfileCollection{
			// Other fields are set when saving the association
			Collection: c,
			Position:   int64(i),
		})
	}

	err = c.Save(rc.DB(), &hades.SaveParams{
		Record: profile,
		Assocs: []string{"ProfileCollections"},

		PartialJoins: []string{"CollectionGames"},
	})
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	err = sendDBCollections()
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	res := &buse.FetchProfileCollectionsResult{}
	return res, nil
}
