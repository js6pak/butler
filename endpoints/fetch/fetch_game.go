package fetch

import (
	"github.com/go-errors/errors"
	"github.com/itchio/butler/buse"
	"github.com/itchio/butler/buse/messages"
	"github.com/itchio/butler/database/hades"
	"github.com/itchio/butler/database/models"
	itchio "github.com/itchio/go-itchio"
)

func FetchGame(rc *buse.RequestContext, params *buse.FetchGameParams) (*buse.FetchGameResult, error) {
	consumer := rc.Consumer

	if params.GameID == 0 {
		return nil, errors.New("gameId must be non-zero")
	}

	_, client := rc.ProfileClient(params.ProfileID)

	sendDBGame := func() error {
		game := models.GameByID(rc.DB(), params.GameID)
		if game != nil {
			err := messages.FetchGameYield.Notify(rc, &buse.FetchGameYieldNotification{Game: game})
			if err != nil {
				return errors.Wrap(err, 0)
			}
		}
		return nil
	}

	err := sendDBGame()
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	consumer.Debugf("Querying API...")
	gameRes, err := client.GetGame(&itchio.GetGameParams{
		GameID: params.GameID,
	})
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	c := HadesContext(rc)

	err = c.Save(rc.DB(), &hades.SaveParams{
		Record: gameRes.Game,
	})
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	err = sendDBGame()
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	res := &buse.FetchGameResult{}
	return res, nil
}
