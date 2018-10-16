package color

import (
	"testing"

	check "gopkg.in/check.v1"
)

var _ = check.Suite(new(colorSuit))

type colorSuit struct{}

func TestColor(t *testing.T) {
	check.TestingT(t)
}

func (s *colorSuit) TestColor(c *check.C) {
	var (
		txt = "abc"
	)

	c.Assert(len(Cyan(txt)), check.Equals, 12)
	c.Assert(len(Yellow(txt)), check.Equals, 12)
	c.Assert(len(Green(txt)), check.Equals, 12)
	c.Assert(len(Magenta(txt)), check.Equals, 12)
	c.Assert(len(Red(txt)), check.Equals, 12)
	c.Assert(len(Blue(txt)), check.Equals, 12)

	c.Assert(len(IntenseCyan(txt)), check.Equals, 14)
	c.Assert(len(IntenseYellow(txt)), check.Equals, 14)
	c.Assert(len(IntenseGreen(txt)), check.Equals, 14)
	c.Assert(len(IntenseMagenta(txt)), check.Equals, 14)
	c.Assert(len(IntenseRed(txt)), check.Equals, 14)
	c.Assert(len(IntenseBlue(txt)), check.Equals, 14)

	n := len(RandColor(txt))
	c.Assert(n == 12 || n == 14, check.Equals, true)

	for i := 1; i <= len(colorVals); i++ {
		f := SeqColorFunc()
		c.Assert(len(f(txt)) == 12 || len(f(txt)) == 14, check.Equals, true)
	}
}
