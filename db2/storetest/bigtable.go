package storetest

import (
	"context"
	"testing"

	"cloud.google.com/go/bigtable"
	"cloud.google.com/go/bigtable/bttest"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewBigTable(t testing.TB) (*bigtable.Client, *bigtable.AdminClient) {
	srv, err := bttest.NewServer("localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	conn, err := grpc.NewClient(srv.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}

	project, instance := "proj", "instance"
	adminClient, err := bigtable.NewAdminClient(ctx, project, instance, option.WithGRPCConn(conn))
	if err != nil {
		t.Fatal(err)
	}

	client, err := bigtable.NewClientWithConfig(ctx, project, instance, bigtable.ClientConfig{}, option.WithGRPCConn(conn))
	if err != nil {
		t.Fatal(err)
	}

	return client, adminClient
}
