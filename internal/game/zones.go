package game

import "fmt"

func (g *Game) moveToGraveyard(cardInstance *CardInstance, reason string) error {
	player, zone, i, err := g.findCardInZonesFromInstance(cardInstance)
	if err != nil {
		return fmt.Errorf("error finding zone: %w", err)
	}

	var ownerPlayer *PlayerState
	for _, p := range g.Players {
		if p.PlayerID == cardInstance.Owner {
			ownerPlayer = p
			break
		}
	}
	if ownerPlayer == nil {
		return fmt.Errorf("owner %s not found", cardInstance.Owner)
	}
	switch zone {
	case ZoneBoard:
		if i+1 < len(player.Board) {
			player.Board = append(player.Board[:i], player.Board[i+1:]...)
		} else {
			player.Board = player.Board[:i]
		}
		ownerPlayer.Graveyard = append(ownerPlayer.Graveyard, *cardInstance)
	case ZoneHand:
		if i+1 < len(player.Hand) {
			player.Hand = append(player.Hand[:i], player.Hand[i+1:]...)
		} else {
			player.Hand = player.Hand[:i]
		}
		ownerPlayer.Graveyard = append(ownerPlayer.Graveyard, *cardInstance)
	case ZoneGraveyard:
		return fmt.Errorf("unable to move card from graveyard to graveyard")
	case ZoneDeck:
		if i+1 < len(player.Deck) {
			player.Deck = append(player.Deck[:i], player.Deck[i+1:]...)
		} else {
			player.Deck = player.Deck[:i]
		}
		ownerPlayer.Graveyard = append(ownerPlayer.Graveyard, *cardInstance)
	default:
		return fmt.Errorf("unexpected zone %s found", zone)
	}

	g.log("graveyard", cardInstance.Owner, "%s moved to graveyard (%s)", cardInstance.Def.Name, reason)
	return nil
}

func (g *Game) findCardInZonesFromInstance(cardInstance *CardInstance) (*PlayerState, Zone, int, error) {
	for _, player := range g.Players {
		zones := map[Zone][]CardInstance{
			ZoneBoard:     player.Board,
			ZoneHand:      player.Hand,
			ZoneGraveyard: player.Graveyard,
			ZoneDeck:      player.Deck,
		}

		for zone, cards := range zones {
			for i, card := range cards {
				if card.InstanceID == cardInstance.InstanceID {
					return player, zone, i, nil
				}
			}
		}
	}

	return nil, "", -1, fmt.Errorf("card %s (%s) not found in any zone", cardInstance.Def.Name, cardInstance.InstanceID)
}
