package xmlpath_test

import (
	"bytes"
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jcfrperu/goaneco-rest-ensured/xmlpath"
)

const sampleXML = `<?xml version="1.0" encoding="UTF-8"?>
<bookstore name="Grand Central" active="true" rating="4.95">
    <book id="1" category="COOKING">
        <title lang="en">Everyday Italian</title>
        <author>Giada De Laurentiis</author>
        <year>2005</year>
        <price>30.00</price>
        <stock>15</stock>
    </book>
    <book id="2" category="CHILDREN">
        <title lang="en">Harry Potter</title>
        <author>J K. Rowling</author>
        <year>2005</year>
        <price>29.99</price>
        <stock>42</stock>
    </book>
</bookstore>`

type XMLBook struct {
	XMLName  xml.Name `xml:"book"`
	ID       string   `xml:"id,attr"`
	Category string   `xml:"category,attr"`
	Title    string   `xml:"title"`
	Author   string   `xml:"author"`
	Year     int      `xml:"year"`
	Price    float64  `xml:"price"`
	Stock    int      `xml:"stock"`
}

func TestXmlPath_BasicQueries(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	xp, err := xmlpath.From(sampleXML)
	must.NoError(err)
	must.NotNil(xp)

	// Attribute query
	is.Equal("Grand Central", xp.GetString("//bookstore/@name"))
	is.True(xp.GetBool("//bookstore/@active"))
	is.InDelta(4.95, xp.GetFloat("//bookstore/@rating"), 0.001)

	// Child element queries
	is.Equal("Everyday Italian", xp.GetString("//book[@id='1']/title"))
	is.Equal("Giada De Laurentiis", xp.GetString("//book[@id='1']/author"))
	is.Equal(2005, xp.GetInt("//book[@id='1']/year"))
	is.Equal(int64(2005), xp.GetInt64("//book[@id='1']/year"))
	is.InDelta(30.00, xp.GetFloat("//book[@id='1']/price"), 0.001)
	is.Equal(15, xp.GetInt("//book[@id='1']/stock"))

	// Second book
	is.Equal("Harry Potter", xp.GetString("//book[@id='2']/title"))
	is.Equal(42, xp.GetInt("//book[@id='2']/stock"))

	// Exists
	is.True(xp.Exists("//bookstore"))
	is.True(xp.Exists("//book[@id='1']"))
	is.False(xp.Exists("//book[@id='999']"))
}

func TestXmlPath_ListQueries(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	xp, err := xmlpath.From(sampleXML)
	must.NoError(err)

	// String list
	titles := xp.GetStringList("//book/title")
	is.Equal([]string{"Everyday Italian", "Harry Potter"}, titles)

	// Int list
	stocks := xp.GetIntList("//book/stock")
	is.Equal([]int{15, 42}, stocks)

	// Float list
	prices := xp.GetFloatList("//book/price")
	is.Len(prices, 2)
	is.InDelta(30.00, prices[0], 0.001)
	is.InDelta(29.99, prices[1], 0.001)

	// Nodes query
	nodes := xp.GetNodes("//book")
	is.Len(nodes, 2)

	// Empty list for non-matching expression
	is.Empty(xp.GetStringList("//nonexistent"))
	is.Empty(xp.GetIntList("//nonexistent"))
	is.Empty(xp.GetFloatList("//nonexistent"))
	is.Empty(xp.GetNodes("//nonexistent"))
}

func TestXmlPath_ObjectDeserialization(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	xp, err := xmlpath.From(sampleXML)
	must.NoError(err)

	// GetObject
	var book1 XMLBook
	err = xp.GetObject("//book[@id='1']", &book1)
	must.NoError(err)
	is.Equal("1", book1.ID)
	is.Equal("COOKING", book1.Category)
	is.Equal("Everyday Italian", book1.Title)
	is.Equal("Giada De Laurentiis", book1.Author)
	is.Equal(15, book1.Stock)

	// GetObjectTyped generic helper
	book2, err := xmlpath.GetObjectTyped[XMLBook](xp, "//book[@id='2']")
	must.NoError(err)
	is.Equal("2", book2.ID)
	is.Equal("CHILDREN", book2.Category)
	is.Equal("Harry Potter", book2.Title)
	is.Equal(42, book2.Stock)

	// Non-matching path
	var missing XMLBook
	err = xp.GetObject("//book[@id='999']", &missing)
	is.ErrorIs(err, xmlpath.ErrInvalidXPath)
}

func TestXmlPath_ConstructorsAndErrors(t *testing.T) {
	t.Parallel()
	is := assert.New(t)
	must := require.New(t)

	// FromBytes
	xpBytes, err := xmlpath.FromBytes([]byte(sampleXML))
	must.NoError(err)
	is.Equal("Everyday Italian", xpBytes.GetString("//book[1]/title"))

	// FromReader
	xpReader, err := xmlpath.FromReader(bytes.NewReader([]byte(sampleXML)))
	must.NoError(err)
	is.Equal("Harry Potter", xpReader.GetString("//book[2]/title"))

	// FromReader nil
	_, err = xmlpath.FromReader(nil)
	is.Error(err)

	// Invalid XML
	_, err = xmlpath.From("<<<not valid xml>>>")
	is.Error(err)

	// OutputXML
	out := xpBytes.OutputXML("//book[@id='1']/title")
	is.Contains(out, "<title")
	is.Contains(out, "Everyday Italian")

	// Config
	cfg := xmlpath.DefaultConfig()
	xpWithCfg := xpBytes.Using(cfg)
	is.Equal("Grand Central", xpWithCfg.GetString("//bookstore/@name"))
}

func TestXmlPath_NilSafety(t *testing.T) {
	t.Parallel()
	is := assert.New(t)

	var nilXP *xmlpath.XmlPath

	is.Equal("", nilXP.GetString("//book"))
	is.Equal(0, nilXP.GetInt("//book"))
	is.Equal(int64(0), nilXP.GetInt64("//book"))
	is.Equal(0.0, nilXP.GetFloat("//book"))
	is.False(nilXP.GetBool("//book"))
	is.Empty(nilXP.GetStringList("//book"))
	is.Empty(nilXP.GetIntList("//book"))
	is.Empty(nilXP.GetFloatList("//book"))
	is.Empty(nilXP.GetNodes("//book"))
	is.Nil(nilXP.GetNode("//book"))
	is.False(nilXP.Exists("//book"))
	is.Equal("", nilXP.OutputXML("//book"))
	is.Nil(nilXP.Using(xmlpath.DefaultConfig()))

	var b XMLBook
	is.ErrorIs(nilXP.GetObject("//book", &b), xmlpath.ErrNilXmlPath)
	_, err := xmlpath.GetObjectTyped[XMLBook](nilXP, "//book")
	is.ErrorIs(err, xmlpath.ErrNilXmlPath)
}
