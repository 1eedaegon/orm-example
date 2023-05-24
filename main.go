package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"

	"github.com/rotemtam/ent-blog-example/ent"
	"github.com/rotemtam/ent-blog-example/ent/user"
)

func main() {
	var dsn string
	flag.StringVar(&dsn, "dsn", "", "database dsn")
	flag.Parse()

	client, err := ent.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed connecting to mysql %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	if !client.Post.Query().ExistX(ctx) {
		if err := seed(ctx, client); err != nil {
			log.Fatalf("Failed seeding to mysql %v", err)
		}
	}
}

func seed(ctx context.Context, client *ent.Client) error {
	r, err := client.User.Query().
		Where(user.Name("rotemtam")).
		Only(ctx)
	switch {
	case ent.IsNotFound(err):
		r, err = client.User.Create().
			SetName("rotemtam").
			SetEmail("r@hello.world").
			Save(ctx)
		if err != nil {
			return fmt.Errorf("Failed creating user %v", err)
		}
	case err != nil:
		return fmt.Errorf("Failed querying user %v", err)
	}
	return client.Post.Create().
		SetTitle("Hello world!").
		SetBody("This is my first post").
		SetAuthor(r).
		Exec(ctx)
}
