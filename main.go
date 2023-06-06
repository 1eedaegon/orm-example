package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/go-sql-driver/mysql"

	"github.com/1eedaegon/orm-example/ent"
	"github.com/1eedaegon/orm-example/ent/post"
	"github.com/1eedaegon/orm-example/ent/user"
)

var (
	//go:embed templates/*
	resources embed.FS
	tmpl      = template.Must(template.ParseFS(resources, "templates/*.html"))
)

type server struct {
	client *ent.Client
}

func NewServer(client *ent.Client) *server {
	return &server{client: client}
}

// = HTTP index handler =
// 모든 post를 가져와서 SSR by template
func (s *server) index(w http.ResponseWriter, r *http.Request) {
	// post 전부를 가져온다. edge에 의해 author를 가져올 수 있다.
	posts, err := s.client.Post.
		Query().
		WithAuthor().
		Order(ent.Desc(post.FieldCreatedAt)).
		All(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// SSR이 실패하면 Internal server error를 반환한다.
	if err := tmpl.Execute(w, posts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *server) add(w http.ResponseWriter, r *http.Request) {
	author, err := s.client.User.Query().Only(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.client.Post.Create().
		SetTitle(r.FormValue("title")).
		SetBody(r.FormValue("body")).
		SetAuthor(author).
		Exec(r.Context()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// Lightweight http router: chi v5
// chi has two built-in middlewares: Logger, Recorverer
func NewRouter(srv *server) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/", srv.index)
	r.Post("/add", srv.add)
	return r
}

func seed(ctx context.Context, client *ent.Client) error {
	r, err := client.User.Query().
		Where(user.Name("1eedaegon")).
		Only(ctx)
	switch {
	case ent.IsNotFound(err):
		r, err = client.User.Create().
			SetName("1eedaegon").
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
	// if doesn't exist, execute seeding process
	if !client.Post.Query().ExistX(ctx) {
		if err := seed(ctx, client); err != nil {
			log.Fatalf("Failed seeding to mysql %v", err)
		}
	}
	srv := NewServer(client)
	r := NewRouter(srv)
	log.Fatal(http.ListenAndServe(":8080", r))
}
