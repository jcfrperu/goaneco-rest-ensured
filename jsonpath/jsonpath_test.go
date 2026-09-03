package jsonpath_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/jsonpath"
)

const sampleJSON = `{
	"store": {
		"name": "Downtown Books",
		"location": "Main St",
		"isOpen": true,
		"rating": 4.85,
		"book": [
			{
				"category": "reference",
				"author": "Nigel Rees",
				"title": "Sayings of the Century",
				"price": 8.95,
				"stock": 12
			},
			{
				"category": "fiction",
				"author": "Evelyn Waugh",
				"title": "Sword of Honour",
				"price": 12.99,
				"stock": 5
			}
		],
		"bicycle": {
			"color": "red",
			"price": 19.95
		}
	},
	"tags": ["retail", "books", "coffee"]
}`

type Book struct {
	Category string  `json:"category"`
	Author   string  `json:"author"`
	Title    string  `json:"title"`
	Price    float64 `json:"price"`
	Stock    int     `json:"stock"`
}

func TestJsonPath_BasicQueries(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	jp := jsonpath.From(sampleJSON)
	must.NotNil(jp)

	// String query
	is.Equal("Downtown Books", jp.GetString("store.name"))
	is.Equal("Main St", jp.GetString("store.location"))

	// Boolean query
	is.True(jp.GetBool("store.isOpen"))

	// Float query
	is.InDelta(4.85, jp.GetFloat("store.rating"), 0.001)

	// Array access
	is.Equal("Nigel Rees", jp.GetString("store.book.0.author"))
	is.Equal(12, jp.GetInt("store.book.0.stock"))
	is.Equal(int64(12), jp.GetInt64("store.book.0.stock"))
	is.InDelta(12.99, jp.GetFloat("store.book.1.price"), 0.001)

	// Exists
	is.True(jp.Exists("store.name"))
	is.True(jp.Exists("store.book.0.author"))
	is.False(jp.Exists("store.nonexistent"))
}

func TestJsonPath_ListQueries(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	jp := jsonpath.From(sampleJSON)

	// String list
	tags := jp.GetStringList("tags")
	is.Equal([]string{"retail", "books", "coffee"}, tags)

	// Array of results
	books := jp.GetList("store.book")
	is.Len(books, 2)

	// Int list from stock
	stocks := jp.GetIntList("store.book.#.stock")
	is.Equal([]int{12, 5}, stocks)

	// Float list from prices
	prices := jp.GetFloatList("store.book.#.price")
	is.Len(prices, 2)
	is.InDelta(8.95, prices[0], 0.001)
	is.InDelta(12.99, prices[1], 0.001)

	// Map extraction
	bicycleMap := jp.GetMap("store.bicycle")
	is.Contains(bicycleMap, "color")
	is.Equal("red", bicycleMap["color"].String())

	// Non-array returns empty slice
	is.Empty(jp.GetStringList("store.name"))
	is.Empty(jp.GetIntList("store.name"))
	is.Empty(jp.GetFloatList("store.name"))
	is.Empty(jp.GetList("store.name"))
	is.Empty(jp.GetMap("store.name"))
}

func TestJsonPath_ObjectDeserialization(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	jp := jsonpath.From(sampleJSON)

	// Deserialization via GetObject
	var firstBook Book
	err := jp.GetObject("store.book.0", &firstBook)
	must.NoError(err)
	is.Equal("Nigel Rees", firstBook.Author)
	is.Equal(12, firstBook.Stock)

	// Deserialization via GetObjectTyped generic helper
	secondBook, err := jsonpath.GetObjectTyped[Book](jp, "store.book.1")
	must.NoError(err)
	is.Equal("Evelyn Waugh", secondBook.Author)
	is.Equal("Sword of Honour", secondBook.Title)

	// Non-existent path returns error
	var missing Book
	err = jp.GetObject("store.book.99", &missing)
	is.ErrorIs(err, jsonpath.ErrInvalidPath)
}

func TestJsonPath_ConstructorsAndConfig(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	// FromBytes
	jpBytes := jsonpath.FromBytes([]byte(sampleJSON))
	is.Equal("Downtown Books", jpBytes.GetString("store.name"))

	// FromReader
	jpReader, err := jsonpath.FromReader(bytes.NewReader([]byte(sampleJSON)))
	must.NoError(err)
	is.Equal("Downtown Books", jpReader.GetString("store.name"))

	// FromReader nil
	_, err = jsonpath.FromReader(nil)
	is.Error(err)

	// Config
	cfg := jsonpath.DefaultConfig()
	jpWithCfg := jpBytes.Using(cfg)
	is.Equal("Downtown Books", jpWithCfg.GetString("store.name"))

	// Pretty
	prettyStr := jpBytes.Pretty()
	is.Contains(prettyStr, "Downtown Books")
}

func TestJsonPath_NilSafety(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var nilJP *jsonpath.JsonPath

	is.Equal("", nilJP.GetString("foo"))
	is.Equal(0, nilJP.GetInt("foo"))
	is.Equal(int64(0), nilJP.GetInt64("foo"))
	is.Equal(0.0, nilJP.GetFloat("foo"))
	is.False(nilJP.GetBool("foo"))
	is.Empty(nilJP.GetList("foo"))
	is.Empty(nilJP.GetStringList("foo"))
	is.Empty(nilJP.GetIntList("foo"))
	is.Empty(nilJP.GetFloatList("foo"))
	is.Empty(nilJP.GetMap("foo"))
	is.False(nilJP.Exists("foo"))
	is.Equal("", nilJP.Pretty())
	is.Nil(nilJP.Using(jsonpath.DefaultConfig()))

	var b Book
	is.ErrorIs(nilJP.GetObject("store.book.0", &b), jsonpath.ErrNilJsonPath)
	_, err := jsonpath.GetObjectTyped[Book](nilJP, "store.book.0")
	is.ErrorIs(err, jsonpath.ErrNilJsonPath)
}
