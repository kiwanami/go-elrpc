package elrpc

import "testing"

func testCompare(t *testing.T, obj interface{}, expected string) {
	res, err := Encode(obj)
	if err != (error)(nil) {
		perr := err.(*EncodeError)
		//pp.Println(perr)
		t.Error(perr.Msg)
		return
	}
	sres := string(res)
	//pp.Println(sres)
	if sres != expected {
		//pp.Println(sres)
		t.Errorf("Not equal exp:[%s] -> result:[%s]", expected, sres)
	}
}

func TestEncoderPrimitives(t *testing.T) {
	testCompare(t, "hello!", `"hello!"`)
	testCompare(t, nil, `nil`)
	testCompare(t, 1, `1`)
	testCompare(t, -1, `-1`)
	testCompare(t, 1.23, `1.23`)
	testCompare(t, 1.2345678901, `1.2345678901`)
}

func TestEncoderMap1(t *testing.T) {
	m1 := map[string]string{
		"test": "test1",
		"111":  "222",
	}
	testCompare(t, m1, `(("111" . "222") ("test" . "test1"))`)
}

type testStruct struct {
	a1 int
	a2 string
	a3 float64
}

func TestStruct1(t *testing.T) {
	v1 := testStruct{1, "test value", 0.002}
	testCompare(t, v1, `((a1 . 1) (a2 . "test value") (a3 . 0.002))`)
	v2 := &testStruct{2, "OK", 12.345}
	testCompare(t, v2, `((a1 . 2) (a2 . "OK") (a3 . 12.345))`)
}

type testPtrStruct struct {
	b1 int
	b2 *testStruct
}

func TestStruct2(t *testing.T) {
	v1 := testStruct{3, "PTR?", 5.432}
	v2 := &testPtrStruct{
		b1: 1234, b2: &v1,
	}
	testCompare(t, v2, `((b1 . 1234) (b2 . ((a1 . 3) (a2 . "PTR?") (a3 . 5.432))))`)
}

func TestEncoderArray1(t *testing.T) {
	a1 := [4]int{1, 2, 3, 4}
	testCompare(t, a1, `(1 2 3 4)`)
}

func TestEncoderSlice1(t *testing.T) {
	a1 := [4]int{1, 2, 3, 4}
	s1 := a1[1:]
	testCompare(t, s1, `(2 3 4)`)
}

func TestEncoderComplex1(t *testing.T) {
	a1 := []interface{}{
		1, []interface{}{
			2, []interface{}{
				3, 4,
			},
		},
	}
	testCompare(t, a1, `(1 (2 (3 4)))`)

	a2 := map[string]interface{}{
		"a": []int{1, 2, 3},
		"b": map[string]interface{}{
			"c": []int{4, 5, 6},
		},
	}
	//testCompare(t, a2, `((a 1 2 3) (b (c 4 5 6)))`)
	testCompare(t, a2, `(("a" . (1 2 3)) ("b" . (("c" . (4 5 6)))))`)
}

func TestUnicode1(t *testing.T) {
	a1 := "unicode \u2026\u2027 normal"
	testCompare(t, a1, `"unicode …‧ normal"`)
	a2 := "unicode \U0001f607 emoji"
	testCompare(t, a2, `"unicode \U0001f607 emoji"`)
	a3 := "unicode 日本語 japanese"
	testCompare(t, a3, `"unicode 日本語 japanese"`)
}
