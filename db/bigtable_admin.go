package db

import (
	"context"
	"eth2-exporter/utils"
	"log"

	gcp_bigtable "cloud.google.com/go/bigtable"
)

type BigtableAdmin struct {
	client *gcp_bigtable.AdminClient
}

type CreateTables struct {
	Name    string
	ColFams []CreateFamily
}

type CreateFamily struct {
	Name   string
	Policy gcp_bigtable.GCPolicy
}

var BigAdminClient *BigtableAdmin

func MustInitBigtableAdmin(ctx context.Context, project, instance string) {
	admin, err := gcp_bigtable.NewAdminClient(ctx, project, instance)
	if err != nil {
		log.Fatalf("Could not create admin client: %v", err)
	}

	bta := &BigtableAdmin{
		client: admin,
	}

	BigAdminClient = bta
}

func (admin *BigtableAdmin) createTables(tables []CreateTables) error {
	ctx := context.Background()

	tableList, err := admin.client.Tables(ctx)
	if err != nil {
		log.Printf("Could not fetch table list")
		return err
	}

	for _, table := range tables {
		if !utils.SliceContains(tableList, table.Name) {
			log.Printf("Creating table %s", table)
			if err := admin.client.CreateTable(ctx, table.Name); err != nil {
				log.Printf("Could not create table %s", table.Name)
				return err
			}
		}

		tblInfo, err := admin.client.TableInfo(ctx, table.Name)
		if err != nil {
			log.Printf("Could not read info for table %s", table.Name)
			return err
		}
		for _, colfam := range table.ColFams {
			if !utils.SliceContains(tblInfo.Families, colfam.Name) {
				if err := admin.client.CreateColumnFamily(ctx, table.Name, colfam.Name); err != nil {
					log.Printf("Could not create column family %s: %v", colfam.Name, err)
					return err
				}
			}
		}
	}
	return nil
}

func (admin *BigtableAdmin) deleteTables(tables []CreateTables) error {
	ctx := context.Background()
	for _, table := range tables {
		if err := admin.client.DeleteTable(ctx, table.Name); err != nil {
			log.Printf("Could not delete table %s err %s", table, err)
			return err
		} else {
			log.Printf("Deleted Table: %v", table.Name)
		}
	}
	return nil
}
