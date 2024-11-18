package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type CityModel struct {
	DB *sqlx.DB
}

func NewCityModel(db *sqlx.DB) *CityModel {
	return &CityModel{DB: db}
}

var (
	ErrRecordNotFound = errors.New("city not found")
	ErrEditConflict   = errors.New("edit conflict")
)

func (m CityModel) Insert(city *City) error {
	query := `
        INSERT INTO cities (name, state)
		VALUES ($1, $2)
		RETURNING id `

	args := []any{
		city.Name,
		city.State}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&city.ID)
	if err != nil {
		return err
	}

	return nil
}

func (m CityModel) Update(city *City) error {
	query := `
        UPDATE cities
		SET name = $1, state = $2
		WHERE id = $3
		RETURNING id`

	args := []any{
		city.Name,
		city.State,
		city.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&city.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			fmt.Println("no rows", err)
			return ErrEditConflict
		default:
			fmt.Println("some error", err)
			return err
		}
	}

	return nil
}

func (m CityModel) Delete(id int64) error {

	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
	DELETE FROM cities
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil

}

func (u CityModel) GetAll(filters Filters) ([]*City, Metadata, error) {

	query := fmt.Sprintf(`
        SELECT count(*) OVER(), id,  name, state
        FROM cities  
        ORDER BY %s %s, id ASC
        LIMIT $1 OFFSET $2`, filters.SortColumn(), filters.SortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{filters.Limit(), filters.Offset()}

	rows, err := u.DB.QueryContext(ctx, query, args...)
	if err != nil {
		fmt.Println("some query error", err)
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0

	cities := []*City{}

	for rows.Next() {
		var city City

		err := rows.Scan(
			&totalRecords,
			&city.ID,
			&city.Name,
			&city.State,
		)
		if err != nil {
			fmt.Println("some row error", err)
			return nil, Metadata{}, err
		}

		cities = append(cities, &city)
	}

	if err = rows.Err(); err != nil {
		fmt.Println("some iter error", err)
		return nil, Metadata{}, err
	}

	metadata := CalculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return cities, metadata, nil

}
