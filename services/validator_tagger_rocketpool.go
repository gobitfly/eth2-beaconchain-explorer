package services

import (
	"context"
	"fmt"

	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/sirupsen/logrus"
)

// enrichRocketPoolSubEntity upserts validator_entities for Rocket Pool using a single INSERT ... SELECT with join.
func enrichRocketPoolSubEntity(ctx context.Context) error {
	res, err := db.WriterDb.Exec(`
			INSERT INTO validator_entities (publickey, entity, sub_entity)
			SELECT DISTINCT ON (rpm.pubkey) rpm.pubkey, 'Rocket Pool', '0x' || ENCODE(rpm.node_address, 'hex')
			FROM rocketpool_minipools rpm
			ON CONFLICT (publickey) DO UPDATE
			SET entity = 'Rocket Pool',
			    sub_entity = EXCLUDED.sub_entity
		`)
	if err != nil {
		return fmt.Errorf("upsert validator_entities (Rocket Pool): %w", err)
	}
	if n, err2 := res.RowsAffected(); err2 == nil {
		validatorTaggerLogger.WithFields(logrus.Fields{"affected": n}).Info("upserted Rocket Pool validator_entities rows")
	}
	return nil
}
