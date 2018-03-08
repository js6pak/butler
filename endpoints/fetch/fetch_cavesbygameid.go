package fetch

import (
	"github.com/itchio/butler/buse"
	"github.com/itchio/butler/database/models"
)

func FetchCavesByGameID(rc *buse.RequestContext, params *buse.FetchCavesByGameIDParams) (*buse.FetchCavesByGameIDResult, error) {
	caves := models.CavesByGameID(rc.DB(), params.GameID)
	models.PreloadCaves(rc.DB(), caves)

	var formattedCaves []*buse.Cave
	for _, c := range caves {
		formattedCaves = append(formattedCaves, formatCave(rc.DB(), c))
	}

	res := &buse.FetchCavesByGameIDResult{
		Caves: formattedCaves,
	}
	return res, nil
}
