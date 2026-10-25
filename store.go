package main

import (
	"database/sql"
	"errors"
	"fmt"
	"go/types"
	"os"

	_ "modernc.org/sqlite"
)

const DBFILE = "app.db"

func setUpDB() (*sql.DB, error) {
	// Reset the database before setting up
	if err := os.Remove(DBFILE); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	// Setup the sqlite database
	db, err := sql.Open("sqlite", DBFILE+"?_pragma=foreign_keys(1)")
	if err != nil {
		return nil, err
	}

	// name below should also include the package name
	q := map[string]string{
		"apis": `CREATE TABLE apis (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		doc TEXT NOT NULL,

		UNIQUE (name)
	)`,
		"concrete_types": `CREATE TABLE concrete_types (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		doc TEXT NOT NULL,

		UNIQUE (name)
	)`,
		"abstract_types": `CREATE TABLE abstract_types (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		doc TEXT NOT NULL,

		UNIQUE (name)
	)`,
		"implements": `CREATE TABLE implements (
		id INTEGER PRIMARY KEY,
		concrete_type TEXT NOT NULL,
		abstract_type TEXT NOT NULL,
		
		UNIQUE (concrete_type, abstract_type)
	)`,
	}
	for table, query := range q {
		if _, err := db.Exec(query); err != nil {
			return nil, err
		}
		fmt.Printf("Created table %s\n", table)
	}
	return db, nil
}

func insertAll[T any](tx *sql.Tx, query string, rows []T, args func(T) []any) error {
	for _, r := range rows {
		_, err := tx.Exec(query, args(r)...)
		if err != nil {
			return fmt.Errorf("query %s failed :%w", query, err)
		}
	}
	return nil
}

func loadDB(index *Index) error {
	db, err := setUpDB()
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to open transaction to load concrete types: %w", err)
	}
	defer tx.Rollback()

	// load concrete types
	if err := insertAll(
		tx, `INSERT INTO concrete_types (name, doc) VALUES (?,?)`, index.Concrete,
		func(row *types.TypeName) []any { return []any{row.Type().String(), index.Docs[row]} },
	); err != nil {
		return err
	}

	// load abstract types
	if err := insertAll(
		tx, `INSERT INTO abstract_types (name, doc) VALUES (?,?)`, index.Abstract,
		func(row *types.TypeName) []any { return []any{row.Type().String(), index.Docs[row]} },
	); err != nil {
		return err
	}

	// load function apis
	if err := insertAll(
		tx, `INSERT INTO apis (name, doc) VALUES (?,?)`, index.Funcs,
		func(row *types.Func) []any { return []any{row.FullName(), index.Docs[row]} },
	); err != nil {
		return err
	}

	//load implements data
	if err := insertAll(
		tx, `INSERT INTO implements (concrete_type, abstract_type) VALUES (?,?)`, index.Implements,
		func(im Implementation) []any { return []any{im.Concrete, im.Abstract} },
	); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit error: %w", err)
	}
	fmt.Println("Successfully loaded all data structures")
	return nil
}
