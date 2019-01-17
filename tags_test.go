package gremgoser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTagParsing(t *testing.T) {
	assert := assert.New(t)

	// test proper graph tags
	name, opts := parseTag("graph,string")
	assert.Equal("graph", name)
	_opts := tagOptions("string")
	assert.Equal(_opts, opts)

	// test no tags
	name, opts = parseTag("")
	assert.Equal("", name)
	_opts = tagOptions("")
	assert.Equal(_opts, opts)
}

func TestTagContains(t *testing.T) {
	assert := assert.New(t)

	// test proper graph tags
	_, opts := parseTag("graph,string,foo")
	c := opts.Contains("string")
	assert.Equal(true, c)
	c = opts.Contains("foo")
	assert.Equal(true, c)
	c = opts.Contains("bar")
	assert.Equal(false, c)

	// test no tags
	_, opts = parseTag("")
	c = opts.Contains("string")
	assert.Equal(false, c)
}
