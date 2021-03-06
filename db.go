package server

import (
	"database/sql"
	"log"
	"os"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"

	"github.com/rs/xid"
)

// Query Defines the Product interface, that receives a context
type Queries interface {
	QueryAll() ([]Product, error)
	QueryRow(id string) (Product, error)
	CreateRow(name string, price int) (Product, error)
	DeleteRow(id string) (string, error)
}

// DB defines a database type
type DB struct {
	db *sql.DB
}

// QueryAll elements from db
func (sqlDB *DB) QueryAll() ([]Product, error) {
	statement := sq.
		Select("id, name, price").
		From("products").
		RunWith(sqlDB.db)

	rows, err := statement.Query()
	if err != nil {
		return nil, err
	}
	return Products(rows)
}

// QueryRow queries a single row
func (sqlDB *DB) QueryRow(id string) (Product, error) {
	statement := sq.
		Select("id, name, price").
		From("products").
		Where("id", id)

	product := Product{}
	err := statement.Scan(&product.ID, &product.Name, &product.Price)
	if err != nil {
		return Product{}, err
	}
	return product, nil
}

// CreateRow creates a new row in the database
func (sqlDB *DB) CreateRow(name string, price int) (Product, error) {
	guid := xid.New()
	var productID string
	statement := sq.
		Insert("products").
		Columns("id", "name", "price").
		Values(guid, name, price).
		RunWith(sqlDB.db).
		PlaceholderFormat(sq.Dollar)

	_, err := statement.Exec()
	if err != nil {
		log.Println("error on create row", err)
		return Product{}, err
	}

	log.Printf("current product id %s \n", productID)
	return Product{ID: guid.String(), Name: name, Price: price}, nil
}

// DeleteRow deletes a row in the database
func (sqlDB *DB) DeleteRow(id string) (string, error) {
	statement := sq.Delete("products").Where("id = $1", id).RunWith(sqlDB.db)
	r, err := statement.Exec()
	if err != nil {
		return "", err
	}
	dr, _ := r.RowsAffected()
	log.Printf("removed row %d \n", dr)
	return id, nil
}

// get url based on MODE env variable
func Url() string {
	mode := os.Getenv("MODE")

	// if database should be fetch from internal docker network (by docker-compose)
	if mode == "DOCKER" {
		dockerUrl := os.Getenv("DATABASE_DOCKER_URL")
		return dockerUrl
	}

	// if database should be fetch from localhost
	dockerUrl := os.Getenv("DATABASE_DEV_URL")
	return dockerUrl
}

// Start starts a database connection
func Start() (DB, error) {
	c, err := pgx.ParseConfig(Url())
	if err != nil {
		log.Println("error here")
		log.Fatal(err)
	}
	connStr := stdlib.RegisterConnConfig(c)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return DB{}, err
	}
	return DB{db: db}, nil
}
