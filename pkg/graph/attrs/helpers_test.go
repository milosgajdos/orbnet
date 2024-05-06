package attrs

import (
	"image/color"
	"reflect"
	"testing"
	"time"
)

var (
	testDate  = time.Date(2021, time.November, 10, 23, 0, 0, 0, time.UTC)
	testColor = color.RGBA{R: 245, G: 154, B: 240}
)

const (
	fooStr   = "foo"
	barStr   = "bar"
	dateStr  = "2021-11-10T23:00:00Z"
	colorStr = "#f59af0"
)

type Foo struct{}

func (f Foo) String() string { return fooStr }

type Bar struct{}

func (b Bar) GoString() string { return barStr }

func TestNewCopyFrom(t *testing.T) {
	a := map[string]interface{}{
		"foo":   Foo{},
		"rand":  10,
		"date":  testDate,
		"color": testColor,
	}

	a2 := CopyFrom(a)

	if !reflect.DeepEqual(a, a2) {
		t.Errorf("expected %v, got: %v", a, a2)
	}
}

func TestAttrsToStringMap(t *testing.T) {
	a := map[string]interface{}{
		"foo":   Foo{},
		"rand":  10,
		"date":  testDate,
		"color": testColor,
	}

	exp := map[string]string{
		"foo":   fooStr,
		"date":  dateStr,
		"color": colorStr,
	}

	res := ToStringMap(a)

	if !reflect.DeepEqual(res, exp) {
		t.Fatalf("expected: %v, got: %v", exp, res)
	}
}

func TestIsStringly(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		a   interface{}
		ok  bool
		val string
	}{
		{Foo{}, true, fooStr},
		{10, false, ""},
		{"string", true, "string"},
		{Bar{}, true, barStr},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run("Attr", func(t *testing.T) {
			t.Parallel()
			ok, val := isStringly(tc.a)
			if ok != tc.ok {
				t.Errorf("expected stringly: %v, got: %v", tc.ok, ok)
			}
			if val != tc.val {
				t.Errorf("expected val %s, got: %s", tc.val, val)
			}
		})
	}
}

func TestToString(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		k   string
		a   interface{}
		exp string
	}{
		{"color", testColor, colorStr},
		{"foo", "bar", ""},
		{"date", testDate, dateStr},
		{"weight", "foo", ""},
		{"weight", 1.2, "1.200000"},
		{"name", "somename", "somename"},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run("Attr", func(t *testing.T) {
			t.Parallel()
			val := ToString(tc.k, tc.a)
			if val != tc.exp {
				t.Errorf("expected val: %s, got: %s", tc.exp, val)
			}
		})
	}
}
