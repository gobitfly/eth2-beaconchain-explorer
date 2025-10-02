package services

import (
	"context"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gobitfly/eth2-beaconchain-explorer/db"
	"github.com/gobitfly/eth2-beaconchain-explorer/dune"
	"github.com/gobitfly/eth2-beaconchain-explorer/metrics"
	"github.com/gobitfly/eth2-beaconchain-explorer/types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/sirupsen/logrus"
)

func cacheCSVPath(queryID int, limit int) string {
	// include limit in filename to avoid collisions between full and limited datasets
	if limit > 0 {
		return filepath.Join(os.TempDir(), fmt.Sprintf("validator_entities_%d_limit_%d.csv", queryID, limit))
	}
	return filepath.Join(os.TempDir(), fmt.Sprintf("validator_entities_%d.csv", queryID))
}

// loadCSVIntoDB truncates and COPY loads validator_entities from the provided CSV reader.
// If the CSV contains multiple rows for the same publickey, only the first occurrence
// (by CSV order) is inserted into the table to avoid PK violations.
func loadCSVIntoDB(ctx context.Context, rdr io.Reader) (int64, error) {
	// Reuse existing writer connection and retrieve underlying pgx.Conn (see db.WriteValidatorStatisticsForDay)
	sqlConn, err := db.WriterDb.Conn(ctx)
	if err != nil {
		return 0, fmt.Errorf("get writer sql conn: %w", err)
	}
	defer sqlConn.Close()

	var inserted int64
	if err := sqlConn.Raw(func(dc interface{}) error {
		pgxConn := dc.(*stdlib.Conn).Conn()

		pgxTx, err := pgxConn.Begin(ctx)
		if err != nil {
			return fmt.Errorf("pgx begin: %w", err)
		}
		defer pgxTx.Rollback(ctx)

		// Truncate the target table and COPY directly into validator_entities.
		// Duplicate pubkeys are skipped by entityCSVSource to avoid PK conflicts.
		if _, err := pgxTx.Exec(ctx, "TRUNCATE validator_entities"); err != nil {
			return fmt.Errorf("truncate validator_entities: %w", err)
		}

		// CSV reader and CopyFrom source
		reader := csv.NewReader(rdr)
		reader.FieldsPerRecord = -1

		src := &entityCSVSource{r: reader}
		cols := []string{"publickey", "entity", "sub_entity"}

		n, err := pgxTx.CopyFrom(ctx, pgx.Identifier{"validator_entities"}, cols, src)
		if err != nil {
			return fmt.Errorf("copy into validator_entities: %w", err)
		}
		inserted = n

		if err := pgxTx.Commit(ctx); err != nil {
			return fmt.Errorf("commit: %w", err)
		}
		return nil
	}); err != nil {
		return 0, err
	}
	return inserted, nil
}

// refreshAndLoadValidatorNames refreshes a Dune query and streams the CSV results into the DB using the known schema.
// currently will consume around 600 dune credits for the full file (no row limit)
func refreshAndLoadValidatorNames(ctx context.Context, cfg types.ValidatorTaggerConfig) error {
	// 0) If a local CSV path is provided, import from it and do not delete the file.
	if strings.TrimSpace(cfg.LocalCSVPath) != "" {
		p := strings.TrimSpace(cfg.LocalCSVPath)
		validatorTaggerLogger.WithField("path", p).Info("importing validator_entities from local CSV path")
		f, err := os.Open(p)
		if err != nil {
			return fmt.Errorf("open local csv: %w", err)
		}
		defer f.Close()
		inserted, err := loadCSVIntoDB(ctx, f)
		if err != nil {
			return fmt.Errorf("import local csv failed: %w", err)
		}
		metrics.Counter.WithLabelValues("validator_tagger_import_local").Inc()
		metrics.Counter.WithLabelValues("validator_tagger_import_local_rows_inserted").Add(float64(inserted))
		validatorTaggerLogger.WithFields(logrus.Fields{"rows": inserted, "table": "validator_entities", "source": "local"}).Info("loaded validator entities from local csv via COPY")
		return nil
	}

	path := cacheCSVPath(cfg.Dune.QueryID, cfg.Dune.LimitRows)

	// 1) If a cached CSV exists, try to import from it first to avoid paying Dune costs.
	if fi, err := os.Stat(path); err == nil && fi.Size() > 0 {
		validatorTaggerLogger.WithField("path", path).Info("found cached validator_entities CSV; attempting import")
		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open cached csv: %w", err)
		}
		defer f.Close()
		inserted, err := loadCSVIntoDB(ctx, f)
		if err != nil {
			return fmt.Errorf("import cached csv failed: %w", err)
		}
		metrics.Counter.WithLabelValues("validator_tagger_import_cache_hit").Inc()
		metrics.Counter.WithLabelValues("validator_tagger_import_cache_rows_inserted").Add(float64(inserted))
		_ = os.Remove(path)
		validatorTaggerLogger.WithFields(logrus.Fields{"rows": inserted, "table": "validator_entities", "source": "cache"}).Info("loaded validator entities from cached csv via COPY")
		return nil
	}

	// 2) No cache: fetch from Dune, store to a stable temp file, then import; keep file on failure.
	client := dune.NewClient(cfg.Dune.ApiKey)
	csvBody, err := client.FetchCSVWithLimit(ctx, cfg.Dune.QueryID, cfg.Dune.Timeout, cfg.Dune.LimitRows)
	if err != nil {
		return fmt.Errorf("dune fetch csv: %w", err)
	}
	// Write to temp file atomically
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("ensure temp dir: %w", err)
	}
	tmpFile, err := os.CreateTemp(dir, "validator_entities_*.csv")
	if err != nil {
		return fmt.Errorf("create temp csv: %w", err)
	}
	_, copyErr := io.Copy(tmpFile, csvBody)
	closeBodyErr := csvBody.Close()
	closeTmpErr := tmpFile.Close()
	if copyErr != nil {
		return fmt.Errorf("write temp csv: %w", copyErr)
	}
	if closeBodyErr != nil {
		validatorTaggerLogger.WithError(closeBodyErr).Warn("error closing dune csv body")
	}
	if closeTmpErr != nil {
		return fmt.Errorf("close temp csv: %w", closeTmpErr)
	}
	if err := os.Rename(tmpFile.Name(), path); err != nil {
		return fmt.Errorf("finalize temp csv: %w", err)
	}
	validatorTaggerLogger.WithField("path", path).Info("cached dune csv to temp file")

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open cached csv for import: %w", err)
	}
	defer f.Close()
	inserted, err := loadCSVIntoDB(ctx, f)
	if err != nil {
		// Keep the cached file for the next run
		return fmt.Errorf("import fresh csv failed (cached for retry at %s): %w", path, err)
	}
	metrics.Counter.WithLabelValues("validator_tagger_import_cache_miss").Inc()
	metrics.Counter.WithLabelValues("validator_tagger_import_dune_rows_inserted").Add(float64(inserted))
	_ = os.Remove(path)
	validatorTaggerLogger.WithFields(logrus.Fields{"rows": inserted, "table": "validator_entities", "source": "dune"}).Info("loaded validator entities from dune csv via COPY")
	return nil
}

type entityCSVSource struct {
	r          *csv.Reader
	headerDone bool
	idxPubkey  int
	idxEntity  int
	idxSubEnt  int
	cur        []any
	err        error
	seen       map[string]struct{}
}

func (s *entityCSVSource) Next() bool {
	for {
		if !s.headerDone {
			headers, err := s.r.Read()
			if err == io.EOF {
				s.err = nil
				return false
			}
			if err != nil {
				s.err = fmt.Errorf("read csv header: %w", err)
				return false
			}
			indexOf := func(name string) int {
				name = strings.ToLower(strings.TrimSpace(name))
				for i, h := range headers {
					if strings.ToLower(strings.TrimSpace(h)) == name {
						return i
					}
				}
				return -1
			}
			s.idxPubkey = indexOf("pubkey")
			s.idxEntity = indexOf("entity")
			s.idxSubEnt = indexOf("sub_entity")
			if s.idxPubkey < 0 || s.idxEntity < 0 || s.idxSubEnt < 0 {
				s.err = fmt.Errorf("csv does not contain required headers for validator_entities schema")
				return false
			}
			s.headerDone = true
			if s.seen == nil {
				s.seen = make(map[string]struct{}, 65536)
			}
			continue
		}

		rec, err := s.r.Read()
		if err == io.EOF {
			s.err = nil
			return false
		}
		if err != nil {
			s.err = err
			return false
		}
		if len(rec) == 0 {
			continue
		}
		pk := strings.TrimSpace(rec[s.idxPubkey])
		h := strings.TrimPrefix(strings.TrimPrefix(pk, "0x"), "0X")
		b, err := hex.DecodeString(h)
		if err != nil {
			validatorTaggerLogger.WithField("pubkey", pk).Warnf("invalid pubkey; skipping: %v", err)
			continue
		}
		// skip duplicates: keep only the first occurrence of each publickey
		key := string(b)
		if _, exists := s.seen[key]; exists {
			continue
		}
		s.seen[key] = struct{}{}
		// normalize empty or placeholder strings to SQL NULL by using nil values
		toNull := func(s string) any {
			t := strings.TrimSpace(s)
			lt := strings.ToLower(t)
			if t == "" || lt == "<nil>" || lt == "null" || lt == "na" || lt == "n/a" {
				return nil
			}
			return t
		}
		s.cur = []any{
			b,
			toNull(rec[s.idxEntity]),
			toNull(rec[s.idxSubEnt]),
		}
		return true
	}
}

func (s *entityCSVSource) Values() ([]any, error) { return s.cur, nil }
func (s *entityCSVSource) Err() error             { return s.err }
