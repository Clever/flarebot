package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeTillNextTopicChange(t *testing.T) {
	pt, _ := time.LoadLocation("America/Los_Angeles")
	// simple case
	now := time.Date(2016, 7, 1, 0, 0, 0, 0, pt)
	expected := time.Date(2016, 7, 4, 12, 0, 0, 0, pt).Sub(now)
	actual := timeTillNextTopicChange(now)
	assert.Equal(t, expected, actual)

	// another simple case with time != 00:00:00
	now = time.Date(2016, 7, 1, 11, 15, 0, 0, pt)
	expected = time.Date(2016, 7, 4, 12, 0, 0, 0, pt).Sub(now)
	actual = timeTillNextTopicChange(now)
	assert.Equal(t, expected, actual)

	// Monday before noon
	now = time.Date(2016, 7, 4, 11, 15, 0, 0, pt)
	expected = time.Date(2016, 7, 4, 12, 0, 0, 0, pt).Sub(now)
	actual = timeTillNextTopicChange(now)
	assert.Equal(t, expected, actual)

	// Monday after noon
	now = time.Date(2016, 7, 4, 13, 30, 0, 0, pt)
	expected = time.Date(2016, 7, 11, 12, 0, 0, 0, pt).Sub(now)
	actual = timeTillNextTopicChange(now)
	assert.Equal(t, expected, actual)

	// Monday is across month boundary
	now = time.Date(2016, 7, 29, 13, 30, 0, 0, pt)
	expected = time.Date(2016, 8, 1, 12, 0, 0, 0, pt).Sub(now)
	actual = timeTillNextTopicChange(now)
	assert.Equal(t, expected, actual)

	// Monday is across year boundary
	now = time.Date(2016, 12, 29, 13, 30, 0, 0, pt)
	expected = time.Date(2017, 1, 2, 12, 0, 0, 0, pt).Sub(now)
	actual = timeTillNextTopicChange(now)
	assert.Equal(t, expected, actual)
}

func TestSwapNextTeam(t *testing.T) {
	topic := "TO FIRE A FLARE:  @flarebot fire a flare <p0/p1/p2> <reason>. If you don’t know what team to page, page #oncall-secure-sync."
	expected := "TO FIRE A FLARE:  @flarebot fire a flare <p0/p1/p2> <reason>. If you don’t know what team to page, page #oncall-instant-login."
	actual := swapNextTeam(topic)
	assert.Equal(t, expected, actual)

	topic = "TO FIRE A FLARE:  @flarebot fire a flare <p0/p1/p2> <reason>. If you don’t know what team to page, page #oncall-instant-login."
	expected = "TO FIRE A FLARE:  @flarebot fire a flare <p0/p1/p2> <reason>. If you don’t know what team to page, page #oncall-infra."
	actual = swapNextTeam(topic)
	assert.Equal(t, expected, actual)

	topic = "TO FIRE A FLARE:  @flarebot fire a flare <p0/p1/p2> <reason>. If you don’t know what team to page, page #oncall-infra."
	expected = "TO FIRE A FLARE:  @flarebot fire a flare <p0/p1/p2> <reason>. If you don’t know what team to page, page #oncall-ip-and-de."
	actual = swapNextTeam(topic)
	assert.Equal(t, expected, actual)

	topic = "TO FIRE A FLARE:  @flarebot fire a flare <p0/p1/p2> <reason>. If you don’t know what team to page, page #oncall-ip-and-de."
	expected = "TO FIRE A FLARE:  @flarebot fire a flare <p0/p1/p2> <reason>. If you don’t know what team to page, page #oncall-secure-sync."
	actual = swapNextTeam(topic)
	assert.Equal(t, expected, actual)
}
