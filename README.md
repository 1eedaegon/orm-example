# ORM-EXAMPLE

> Graph 기반 ORM 예시

## 0. Introduction

우리가 자주 사용하는 ORM은 보통 JAVA의 Hibernate, Typescript의 TypeORM 같은 ORM이다.

기존의 ORM들은 Runtime에 Object reflection을 통해 타입 Validation과 Value injection을 구현했는데 2000년도 초반에 구현한 기술인지라 컴파일 언어에도 런타임 주입을 시도한다.

Compile 언어에서 Reflection은 성능상 문제가 있고 Heap 할당이 번번히 일어나기 때문에 Garbage collection의 의해 운영 중에 동작이 멈추기도 한다.

현대 Compile언어들은 Domain 모델로부터 Compile time에 Code generating을 통한 Query/Schema/Validator생성이 충분히 가능하므로 runtime의 성능 손실이 없다.

또한 기존의 ORM은 구조상 n+1문제에 봉착하는데 주로 해결하는 방식은 Relation을 별도로 명시하거나(fetch join), entity의 graph구조를 탐색하거나, n+1의 규모를 작게하거나(Batch size)인데 기존 ORM은 주로 relation을 명시한다. 이건 사실 RDBMS 설계를 들여다보는 행위나 진배없다.

go ent는 entity의 graph관계를 명시하고 graph 탐색 방식으로 n+1 문제에 대해 현명하게 대처하는 구현을 했다.

간단한 CMS 시스템 제작을 예시로 만들어보자.

## 1. Initialize

Create User and Post

```bash
 go run -mod=mod entgo.io/ent/cmd/ent new User Post
```

## 2. Define schema for entities

Write field to post&user on `schema` directory

```go
// Fields of the User.
func (User) Fields() []ent.Field {
   return []ent.Field{
      field.String("name"),
      field.Text("email").
            Unique(),
      field.Time("created_at").
            Default(time.Now),
   }
}

// Edges of the User.
func (User) Edges() []ent.Edge {
   return []ent.Edge{
      edge.To("posts", Post.Type),
   }
}
```

## 3. Generate code

```bash
go generate ./...
```

## 4. View schema

```bash
go run -mod=mod ariga.io/entviz ./ent/schema
```

## 5. Install atlas for migration

```bash
curl -sSf https://atlasgo.sh | sh
```

## 6. Generate migration scripts

```bash
atlas migrate diff add_users_posts \
  --dir "file://ent/migrate/migrations" \
  --to "ent://ent/schema" \
  --dev-url "docker://mysql/8/ent"
```

## 7. Run develop database

```bash
docker run --rm --name entdb -d -p 3306:3306 -e MYSQL_DATABASE=ent -e MYSQL_ROOT_PASSWORD=pass mysql:8
```

## 8. Apply migrations

```bash
atlas migrate apply --dir file://ent/migrate/migrations \
  --url mysql://root:pass@localhost:3306/ent
```

## 9. Seeding

```bash
go get -u github.com/go-sql-driver/mysql
```

```go
package main

import (
    "context"
    "flag"
    "fmt"
    "log"

    "github.com/rotemtam/ent-blog-example/ent"

    _ "github.com/go-sql-driver/mysql"
    "github.com/rotemtam/ent-blog-example/ent/user"
)

func main() {
    // Read the connection string to the database from a CLI flag.
    var dsn string
    flag.StringVar(&dsn, "dsn", "", "database DSN")
    flag.Parse()

    // Instantiate the Ent client.
    client, err := ent.Open("mysql", dsn)
    if err != nil {
        log.Fatalf("failed connecting to mysql: %v", err)
    }
    defer client.Close()

    ctx := context.Background()
    // If we don't have any posts yet, seed the database.
    if !client.Post.Query().ExistX(ctx) {
        if err := seed(ctx, client); err != nil {
            log.Fatalf("failed seeding the database: %v", err)
        }
    }
    // ... Continue with server start.
}

func seed(ctx context.Context, client *ent.Client) error {
    // Check if the user "rotemtam" already exists.
    r, err := client.User.Query().
        Where(
            user.Name("rotemtam"),
        ).
        Only(ctx)
    switch {
    // If not, create the user.
    case ent.IsNotFound(err):
        r, err = client.User.Create().
            SetName("rotemtam").
            SetEmail("r@hello.world").
            Save(ctx)
        if err != nil {
            return fmt.Errorf("failed creating user: %v", err)
        }
    case err != nil:
        return fmt.Errorf("failed querying user: %v", err)
    }
    // Finally, create a "Hello, world" blogpost.
    return client.Post.Create().
        SetTitle("Hello, World!").
        SetBody("This is my first post").
        SetAuthor(r).
        Exec(ctx)
}
```

## Ref:

- https://entgo.io/docs/getting-started/
- https://entgo.io/blog/2023/02/23/simple-cms-with-ent
