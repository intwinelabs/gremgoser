package gremgoser

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

type XXX float32

type Test2 struct {
	Id uuid.UUID `graph:"id,string"`
	A  string    `graph:"a,string"`
	B  int       `graph:"b,number"`
}

type Test struct {
	Id uuid.UUID `graph:"id,string"`
	A  string    `graph:"a,string"`
	B  int       `graph:"b,number"`
	C  int8      `graph:"c,number"`
	D  int16     `graph:"d,number"`
	E  int32     `graph:"e,number"`
	F  int64     `graph:"f,number"`
	G  float32   `graph:"g,number"`
	H  float64   `graph:"h,number"`
	I  uint      `graph:"i,number"`
	J  uint8     `graph:"j,number"`
	K  uint16    `graph:"k,number"`
	L  uint32    `graph:"l,number"`
	M  uint64    `graph:"m,number"`
	N  bool      `graph:"n,bool"`
	AA []string  `graph:"aa,[]string"`
	BB []int     `graph:"bb,[]number"`
	CC []int8    `graph:"cc,[]number"`
	DD []int16   `graph:"dd,[]number"`
	EE []int32   `graph:"ee,[]number"`
	FF []int64   `graph:"ff,[]number"`
	GG []float32 `graph:"gg,[]number"`
	HH []float64 `graph:"hh,[]number"`
	II []uint    `graph:"ii,[]number"`
	JJ []uint8   `graph:"jj,[]number"`
	KK []uint16  `graph:"kk,[]number"`
	LL []uint32  `graph:"ll,[]number"`
	MM []uint64  `graph:"mm,[]number"`
	NN []bool    `graph:"nn,[]bool"`
	X  XXX       `graph:"x,number"`
	XX []XXX     `graph:"xx,[]number"`
	Z  Test2     `graph:"z,struct"`
	ZZ []Test2   `graph:"zz,[]struct"`
}

func TestNewClient(t *testing.T) {
	assert := assert.New(t)

	// Create test server with the mock handler.
	s := httptest.NewServer(http.HandlerFunc(mock))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	user := "foo"
	pass := "bar"
	conf := NewClientConfig(u)
	conf.SetAuthentication(user, pass)
	g, errs := NewClient(conf)
	assert.IsType(&Client{}, g)
	assert.IsType(make(chan error), errs)
	assert.Equal(u, g.conf.URI)
	assert.Equal(time.Duration(5000000000), g.conf.Timeout)
}

func TestExecute(t *testing.T) {
	assert := assert.New(t)

	// Create test server with the mock handler.
	s := httptest.NewServer(http.HandlerFunc(mock))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// test connecting to the mock server
	g, errs := NewClient(NewClientConfig(u))
	assert.IsType(make(chan error), errs)
	assert.NotNil(g)
	assert.IsType(&Client{}, g)

	// setup err channel
	go func(chan error) {
		err := <-errs
		assert.Nil(err)
	}(errs)

	// test query execution
	q := "g.V()"
	resp, err := g.Execute(q, nil, nil)
	assert.Nil(err)
	assert.Nil(resp)
	assert.Equal([]*GremlinData(nil), resp)
}

func TestAddV(t *testing.T) {
	assert := assert.New(t)

	// Create test server with the mock handler.
	s := httptest.NewServer(http.HandlerFunc(mock))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// test connecting to the mock server
	g, errs := NewClient(NewClientConfig(u))
	assert.IsType(make(chan error), errs)
	assert.NotNil(g)
	assert.IsType(&Client{}, g)

	// setup err channel
	go func(chan error) {
		err := <-errs
		assert.Nil(err)
	}(errs)

	// create test struct to pass as interface to AddV
	_tUUID, _ := uuid.Parse("64795211-c4a1-4eac-9e0a-b674ced77461")
	_t := Test{
		Id: _tUUID,
		A:  "aa",
		B:  10,
		C:  20,
		D:  30,
		E:  40,
		F:  50,
		G:  0.06,
		H:  0.07,
		I:  80,
		J:  90,
		K:  100,
		L:  110,
		M:  120,
		N:  true,
		AA: []string{"aa", "aa"},
		BB: []int{10, 10},
		CC: []int8{20, 20},
		DD: []int16{30, 30},
		EE: []int32{40, 40},
		FF: []int64{50, 50},
		GG: []float32{0.06, 0.06},
		HH: []float64{0.07, 0.07},
		II: []uint{80, 80},
		JJ: []uint8{90, 90},
		KK: []uint16{100, 100},
		LL: []uint32{110, 110},
		MM: []uint64{120, 120},
		NN: []bool{true, true},
		X:  XXX(130),
		XX: []XXX{XXX(140), XXX(140)},
		Z: Test2{Id: _tUUID,
			A: "aa",
			B: 10,
		},
		ZZ: []Test2{
			Test2{Id: _tUUID,
				A: "aa",
				B: 10,
			},
			Test2{Id: _tUUID,
				A: "aa",
				B: 10,
			},
		},
	}

	/*_tResp := []*gremgoser.GremlinData{
			&GremlinData{
		{
	  Id: _tUUID,
	  Label: "test",
	  Type: "vertex",
	  Properties: map[string][]gremgoser.GremlinProperty
	    "j": ([]gremgoser.GremlinProperty) (len=1 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) e7235d64-4212-448d-869f-612cf2403b96,
	     Value: (float64) 90
	    }
	   },
	   (string) (len=1) "n": ([]gremgoser.GremlinProperty) (len=1 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 954cc7f9-d655-4123-a66d-e3e665cf7d49,
	     Value: (bool) true
	    }
	   },
	   (string) (len=2) "aa": ([]gremgoser.GremlinProperty) (len=2 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 225ed5a7-b000-4a59-b6c3-332682a5216a,
	     Value: (string) (len=2) "aa"
	    },
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 9cbee039-c5b4-4e75-a1b0-346a47e5dc36,
	     Value: (string) (len=2) "aa"
	    }
	   },
	   (string) (len=1) "b": ([]gremgoser.GremlinProperty) (len=1 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 91df576d-3501-4303-9d89-1c8409ce6ff4,
	     Value: (float64) 10
	    }
	   },
	   (string) (len=2) "hh": ([]gremgoser.GremlinProperty) (len=2 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) b2266aff-1f18-4391-98f8-5ad6a542a2e1,
	     Value: (float64) 0.07
	    },
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 709a88dd-8fc4-4c8d-bd31-61962feff9b2,
	     Value: (float64) 0.07
	    }
	   },
	   (string) (len=1) "e": ([]gremgoser.GremlinProperty) (len=1 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) bf35b756-0640-48c6-9601-aab77c6aa603,
	     Value: (float64) 40
	    }
	   },
	   (string) (len=1) "i": ([]gremgoser.GremlinProperty) (len=1 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 2295b3de-fc5b-42c9-8b04-44b12fbe1346,
	     Value: (float64) 80
	    }
	   },
	   (string) (len=1) "g": ([]gremgoser.GremlinProperty) (len=1 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 2c6860ac-7151-48f9-b866-5b40a3488d1e,
	     Value: (float64) 0.06
	    }
	   },
	   (string) (len=2) "gg": ([]gremgoser.GremlinProperty) (len=2 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) a0a129f4-bfa4-4d1a-bc77-df5b63049197,
	     Value: (float64) 0.06
	    },
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 3c8b51c4-8b40-4d5f-a79d-6a9d74929837,
	     Value: (float64) 0.06
	    }
	   },
	   (string) (len=2) "ii": ([]gremgoser.GremlinProperty) (len=2 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) fe3a148a-4a80-4ca2-851b-5dc473f549e6,
	     Value: (float64) 80
	    },
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 021a9a1a-49c1-4ae3-aa40-3f5c33a12e9f,
	     Value: (float64) 80
	    }
	   },
	   (string) (len=2) "kk": ([]gremgoser.GremlinProperty) (len=2 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 740dca9e-c5d7-40d3-8d68-feb48090a638,
	     Value: (float64) 100
	    },
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) a0587e7d-6bf5-4cba-b911-0507d0469068,
	     Value: (float64) 100
	    }
	   },
	   (string) (len=1) "m": ([]gremgoser.GremlinProperty) (len=1 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) a4b507fe-bb16-4b4a-aa1d-cf922af67cd2,
	     Value: (float64) 120
	    }
	   },
	   (string) (len=2) "mm": ([]gremgoser.GremlinProperty) (len=2 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 0487cc86-ee49-4649-8131-d43610235c40,
	     Value: (float64) 120
	    },
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) b6e9c4c6-8ac4-4124-92cd-e53acf0cfd12,
	     Value: (float64) 120
	    }
	   },
	   (string) (len=1) "c": ([]gremgoser.GremlinProperty) (len=1 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 8e58d327-e06b-44e4-a5d9-75558cdca2dc,
	     Value: (float64) 20
	    }
	   },
	   (string) (len=1) "d": ([]gremgoser.GremlinProperty) (len=1 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 565d7400-e75b-4813-aa39-6c09cae781a8,
	     Value: (float64) 30
	    }
	   },
	   (string) (len=2) "nn": ([]gremgoser.GremlinProperty) (len=2 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 7af1d164-3966-4dde-93fa-511a936601f5,
	     Value: (bool) true
	    },
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 90be7e5c-8bf6-4bfd-bd01-38be1697d9f8,
	     Value: (bool) true
	    }
	   },
	   (string) (len=1) "x": ([]gremgoser.GremlinProperty) (len=1 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 56b71ade-d0aa-416f-8da3-517391fd7ee4,
	     Value: (float64) 130
	    }
	   },
	   (string) (len=2) "cc": ([]gremgoser.GremlinProperty) (len=2 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 8f16c7cd-4125-4d29-b714-5d8f561bb8e4,
	     Value: (float64) 20
	    },
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 84f44c14-f038-47b3-a0b7-9f14dd11ddde,
	     Value: (float64) 20
	    }
	   },
	   (string) (len=1) "h": ([]gremgoser.GremlinProperty) (len=1 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 3e16edbc-5a77-4b97-bdb4-4695996d8915,
	     Value: (float64) 0.07
	    }
	   },
	   (string) (len=2) "jj": ([]gremgoser.GremlinProperty) (len=2 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 84e17824-94be-47e9-b7bb-7e46ed5c065f,
	     Value: (float64) 90
	    },
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 20dffbe0-4c63-4b15-9b5d-43111bd10525,
	     Value: (float64) 90
	    }
	   },
	   (string) (len=1) "l": ([]gremgoser.GremlinProperty) (len=1 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) e7c6ddee-729e-44a4-b977-7d3eafe47497,
	     Value: (float64) 110
	    }
	   },
	   (string) (len=1) "a": ([]gremgoser.GremlinProperty) (len=1 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 15d0a33b-d369-4b61-b162-320ece53cfa1,
	     Value: (string) (len=2) "aa"
	    }
	   },
	   (string) (len=2) "ff": ([]gremgoser.GremlinProperty) (len=2 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 4c2afab4-df69-490b-8cf0-c8311808c0fc,
	     Value: (float64) 50
	    },
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 38e68ad4-b9f8-4256-9638-1e52cdbb989a,
	     Value: (float64) 50
	    }
	   },
	   (string) (len=1) "k": ([]gremgoser.GremlinProperty) (len=1 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) a435d9ad-9f8f-43d6-b108-a2fe1d5a95b9,
	     Value: (float64) 100
	    }
	   },
	   (string) (len=2) "bb": ([]gremgoser.GremlinProperty) (len=2 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) b96f76ed-028a-4e2f-942e-2adf37f5bcb0,
	     Value: (float64) 10
	    },
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 7f010e2c-b764-4601-b190-4b34372203e7,
	     Value: (float64) 10
	    }
	   },
	   (string) (len=2) "ee": ([]gremgoser.GremlinProperty) (len=2 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 8d21dad3-5ed7-4925-bfae-12b345592a36,
	     Value: (float64) 40
	    },
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 6277ad55-a3ed-41bc-8ad2-1cfe6e60938b,
	     Value: (float64) 40
	    }
	   },
	   (string) (len=2) "ll": ([]gremgoser.GremlinProperty) (len=2 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) ff4d7387-b3f7-41ec-9cee-912bb9220545,
	     Value: (float64) 110
	    },
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 519d212c-4774-49a1-bc4e-3715af929c38,
	     Value: (float64) 110
	    }
	   },
	   (string) (len=2) "xx": ([]gremgoser.GremlinProperty) (len=2 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 9cf5c2a7-45eb-4e58-bf5a-9f15186c0819,
	     Value: (float64) 140
	    },
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 122191a7-5437-45ae-9ec6-1a73fee5c996,
	     Value: (float64) 140
	    }
	   },
	   (string) (len=2) "dd": ([]gremgoser.GremlinProperty) (len=2 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 51336fa2-ebf9-4d7a-9ec5-0128a6341ea6,
	     Value: (float64) 30
	    },
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) 337fe24f-8ea3-40ea-b726-e7169883618b,
	     Value: (float64) 30
	    }
	   },
	   (string) (len=1) "f": ([]gremgoser.GremlinProperty) (len=1 cap=4) {
	    (gremgoser.GremlinProperty) {
	     Id: (uuid.UUID) (len=16 cap=16) ed8cf6a7-d585-4575-a08c-cf4aa27f1491,
	     Value: (float64) 50
	    }
	   }
	  }
	 })
	}*/

	resp, err := g.AddV("test", _t)
	spew.Dump(resp)
	assert.Nil(err)
	assert.Equal(nil, resp)
}

func TestUpdateV(t *testing.T) {
	assert := assert.New(t)

	// Create test server with the mock handler.
	s := httptest.NewServer(http.HandlerFunc(mock))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// test connecting to the mock server
	g, errs := NewClient(NewClientConfig(u))
	assert.IsType(make(chan error), errs)
	assert.NotNil(g)
	assert.IsType(&Client{}, g)

	// setup err channel
	go func(chan error) {
		err := <-errs
		assert.Nil(err)
	}(errs)

	// create test struct to pass as interface to AddV
	_tUUID, _ := uuid.Parse("64795211-c4a1-4eac-9e0a-b674ced77461")
	_t := Test{
		Id: _tUUID,
		A:  "aa",
		B:  10,
		C:  20,
		D:  30,
		E:  40,
		F:  50,
		G:  0.06,
		H:  0.07,
		I:  80,
		J:  90,
		K:  100,
		L:  110,
		M:  120,
		N:  true,
		AA: []string{"aa", "aa"},
		BB: []int{10, 10},
		CC: []int8{20, 20},
		DD: []int16{30, 30},
		EE: []int32{40, 40},
		FF: []int64{50, 50},
		GG: []float32{0.06, 0.06},
		HH: []float64{0.07, 0.07},
		II: []uint{80, 80},
		JJ: []uint8{90, 90},
		KK: []uint16{100, 100},
		LL: []uint32{110, 110},
		MM: []uint64{120, 120},
		NN: []bool{true, true},
		X:  XXX(130),
		XX: []XXX{XXX(140), XXX(140)},
		Z: Test2{Id: _tUUID,
			A: "aa",
			B: 10,
		},
		ZZ: []Test2{
			Test2{Id: _tUUID,
				A: "aa",
				B: 10,
			},
			Test2{Id: _tUUID,
				A: "aa",
				B: 10,
			},
		},
	}

	_tResp := []interface{}{
		[]interface{}{
			map[string]interface{}{
				"id":    "c1f7a921-b767-4839-bbdc-6478eb5f3454",
				"label": "test",
				"properties": map[string]interface{}{
					"gg": []interface{}{
						map[string]interface{}{
							"id":    "a0a129f4-bfa4-4d1a-bc77-df5b63049197",
							"value": float64(0.06)},
						map[string]interface{}{
							"id":    "3c8b51c4-8b40-4d5f-a79d-6a9d74929837",
							"value": float64(0.06)}},
					"m": []interface{}{
						map[string]interface{}{
							"value": float64(120),
							"id":    "a4b507fe-bb16-4b4a-aa1d-cf922af67cd2"}},
					"x": []interface{}{
						map[string]interface{}{
							"value": float64(130),
							"id":    "56b71ade-d0aa-416f-8da3-517391fd7ee4"}},
					"xx": []interface{}{
						map[string]interface{}{
							"id":    "9cf5c2a7-45eb-4e58-bf5a-9f15186c0819",
							"value": float64(140)},
						map[string]interface{}{
							"id":    "122191a7-5437-45ae-9ec6-1a73fee5c996",
							"value": float64(140)}},
					"bb": []interface{}{
						map[string]interface{}{
							"id":    "b96f76ed-028a-4e2f-942e-2adf37f5bcb0",
							"value": float64(10)},
						map[string]interface{}{
							"id":    "7f010e2c-b764-4601-b190-4b34372203e7",
							"value": float64(10)}},
					"cc": []interface{}{
						map[string]interface{}{
							"id":    "8f16c7cd-4125-4d29-b714-5d8f561bb8e4",
							"value": float64(20)},
						map[string]interface{}{
							"id":    "84f44c14-f038-47b3-a0b7-9f14dd11ddde",
							"value": float64(20)}},
					"ee": []interface{}{
						map[string]interface{}{
							"value": float64(40),
							"id":    "8d21dad3-5ed7-4925-bfae-12b345592a36"},
						map[string]interface{}{
							"id":    "6277ad55-a3ed-41bc-8ad2-1cfe6e60938b",
							"value": float64(40)}},
					"n": []interface{}{
						map[string]interface{}{
							"id":    "954cc7f9-d655-4123-a66d-e3e665cf7d49",
							"value": true}},
					"aa": []interface{}{
						map[string]interface{}{
							"id":    "225ed5a7-b000-4a59-b6c3-332682a5216a",
							"value": "aa"},
						map[string]interface{}{
							"id":    "9cbee039-c5b4-4e75-a1b0-346a47e5dc36",
							"value": "aa"}},
					"b": []interface{}{
						map[string]interface{}{
							"id":    "91df576d-3501-4303-9d89-1c8409ce6ff4",
							"value": float64(10)}},
					"d": []interface{}{
						map[string]interface{}{
							"id":    "565d7400-e75b-4813-aa39-6c09cae781a8",
							"value": float64(30)}},
					"e": []interface{}{
						map[string]interface{}{
							"id":    "bf35b756-0640-48c6-9601-aab77c6aa603",
							"value": float64(40)}},
					"ff": []interface{}{
						map[string]interface{}{
							"id":    "4c2afab4-df69-490b-8cf0-c8311808c0fc",
							"value": float64(50)},
						map[string]interface{}{
							"id":    "38e68ad4-b9f8-4256-9638-1e52cdbb989a",
							"value": float64(50)}},
					"g": []interface{}{
						map[string]interface{}{
							"id":    "2c6860ac-7151-48f9-b866-5b40a3488d1e",
							"value": float64(0.06)}},
					"ii": []interface{}{
						map[string]interface{}{
							"id":    "fe3a148a-4a80-4ca2-851b-5dc473f549e6",
							"value": float64(80)},
						map[string]interface{}{
							"id":    "021a9a1a-49c1-4ae3-aa40-3f5c33a12e9f",
							"value": float64(80)}},
					"mm": []interface{}{
						map[string]interface{}{
							"id":    "0487cc86-ee49-4649-8131-d43610235c40",
							"value": float64(120)},
						map[string]interface{}{
							"id":    "b6e9c4c6-8ac4-4124-92cd-e53acf0cfd12",
							"value": float64(120)}},
					"nn": []interface{}{
						map[string]interface{}{
							"id":    "7af1d164-3966-4dde-93fa-511a936601f5",
							"value": true},
						map[string]interface{}{
							"id":    "90be7e5c-8bf6-4bfd-bd01-38be1697d9f8",
							"value": true}},
					"a": []interface{}{
						map[string]interface{}{
							"value": "aa",
							"id":    "15d0a33b-d369-4b61-b162-320ece53cfa1"}},
					"dd": []interface{}{
						map[string]interface{}{
							"id":    "51336fa2-ebf9-4d7a-9ec5-0128a6341ea6",
							"value": float64(30)},
						map[string]interface{}{
							"id":    "337fe24f-8ea3-40ea-b726-e7169883618b",
							"value": float64(30)}},
					"f": []interface{}{
						map[string]interface{}{
							"id":    "ed8cf6a7-d585-4575-a08c-cf4aa27f1491",
							"value": float64(50)}},
					"jj": []interface{}{
						map[string]interface{}{
							"id":    "84e17824-94be-47e9-b7bb-7e46ed5c065f",
							"value": float64(90)},
						map[string]interface{}{
							"id":    "20dffbe0-4c63-4b15-9b5d-43111bd10525",
							"value": float64(90)}},
					"k": []interface{}{
						map[string]interface{}{
							"value": float64(100),
							"id":    "a435d9ad-9f8f-43d6-b108-a2fe1d5a95b9"}},
					"ll": []interface{}{
						map[string]interface{}{
							"id":    "ff4d7387-b3f7-41ec-9cee-912bb9220545",
							"value": float64(110)},
						map[string]interface{}{
							"id":    "519d212c-4774-49a1-bc4e-3715af929c38",
							"value": float64(110)}},
					"c": []interface{}{
						map[string]interface{}{
							"id":    "8e58d327-e06b-44e4-a5d9-75558cdca2dc",
							"value": float64(20)}},
					"j": []interface{}{
						map[string]interface{}{
							"id":    "e7235d64-4212-448d-869f-612cf2403b96",
							"value": float64(90)}},
					"l": []interface{}{
						map[string]interface{}{
							"id":    "e7c6ddee-729e-44a4-b977-7d3eafe47497",
							"value": float64(110)}},
					"h": []interface{}{
						map[string]interface{}{
							"id":    "3e16edbc-5a77-4b97-bdb4-4695996d8915",
							"value": float64(0.07)}},
					"hh": []interface{}{
						map[string]interface{}{
							"id":    "b2266aff-1f18-4391-98f8-5ad6a542a2e1",
							"value": float64(0.07)},
						map[string]interface{}{
							"id":    "709a88dd-8fc4-4c8d-bd31-61962feff9b2",
							"value": float64(0.07)}},
					"i": []interface{}{
						map[string]interface{}{
							"id":    "2295b3de-fc5b-42c9-8b04-44b12fbe1346",
							"value": float64(80)}},
					"kk": []interface{}{
						map[string]interface{}{
							"id":    "740dca9e-c5d7-40d3-8d68-feb48090a638",
							"value": float64(100)},
						map[string]interface{}{
							"id":    "a0587e7d-6bf5-4cba-b911-0507d0469068",
							"value": float64(100)}}},
				"type": "vertex"}}}

	resp, err := g.UpdateV(_t)
	assert.Nil(err)
	assert.Equal(_tResp, resp)
}

func TestDropV(t *testing.T) {
	assert := assert.New(t)

	// Create test server with the mock handler.
	s := httptest.NewServer(http.HandlerFunc(mock))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// test connecting to the mock server
	g, errs := NewClient(NewClientConfig(u))
	assert.IsType(make(chan error), errs)
	assert.NotNil(g)
	assert.IsType(&Client{}, g)

	// setup err channel
	go func(chan error) {
		err := <-errs
		assert.Nil(err)
	}(errs)

	// create test struct to pass as interface to AddV
	_tUUID, _ := uuid.Parse("64795211-c4a1-4eac-9e0a-b674ced77461")
	_t := Test{
		Id: _tUUID,
		A:  "aa",
		B:  10,
		C:  20,
		D:  30,
		E:  40,
		F:  50,
		G:  0.06,
		H:  0.07,
		I:  80,
		J:  90,
		K:  100,
		L:  110,
		M:  120,
		N:  true,
		AA: []string{"aa", "aa"},
		BB: []int{10, 10},
		CC: []int8{20, 20},
		DD: []int16{30, 30},
		EE: []int32{40, 40},
		FF: []int64{50, 50},
		GG: []float32{0.06, 0.06},
		HH: []float64{0.07, 0.07},
		II: []uint{80, 80},
		JJ: []uint8{90, 90},
		KK: []uint16{100, 100},
		LL: []uint32{110, 110},
		MM: []uint64{120, 120},
		NN: []bool{true, true},
		X:  XXX(130),
		XX: []XXX{XXX(140), XXX(140)},
		Z: Test2{Id: _tUUID,
			A: "aa",
			B: 10,
		},
		ZZ: []Test2{
			Test2{Id: _tUUID,
				A: "aa",
				B: 10,
			},
			Test2{Id: _tUUID,
				A: "aa",
				B: 10,
			},
		},
	}

	_tResp := []interface{}([]interface{}{[]interface{}{}})
	resp, err := g.DropV(_t)
	assert.Nil(err)
	assert.Equal(_tResp, resp)
}

func TestAddE(t *testing.T) {
	assert := assert.New(t)

	// Create test server with the mock handler.
	s := httptest.NewServer(http.HandlerFunc(mock))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// test connecting to the mock server
	g, errs := NewClient(NewClientConfig(u))
	assert.IsType(make(chan error), errs)
	assert.NotNil(g)
	assert.IsType(&Client{}, g)

	// setup err channel
	go func(chan error) {
		err := <-errs
		assert.Nil(err)
	}(errs)

	_tUUID, _ := uuid.Parse("64795211-c4a1-4eac-9e0a-b674ced77461")
	_t := Test{
		Id: _tUUID,
		A:  "aa",
		B:  10,
		C:  20,
		D:  30,
		E:  40,
		F:  50,
		G:  0.06,
		H:  0.07,
		I:  80,
		J:  90,
		K:  100,
		L:  110,
		M:  120,
		N:  true,
		AA: []string{"aa", "aa"},
		BB: []int{10, 10},
		CC: []int8{20, 20},
		DD: []int16{30, 30},
		EE: []int32{40, 40},
		FF: []int64{50, 50},
		GG: []float32{0.06, 0.06},
		HH: []float64{0.07, 0.07},
		II: []uint{80, 80},
		JJ: []uint8{90, 90},
		KK: []uint16{100, 100},
		LL: []uint32{110, 110},
		MM: []uint64{120, 120},
		NN: []bool{true, true},
		X:  XXX(130),
		XX: []XXX{XXX(140), XXX(140)},
	}

	_t2UUID, _ := uuid.Parse("dafeafc6-63a7-42b2-8ac2-4b85c3e2e37a")
	_t2 := Test{
		Id: _t2UUID,
		A:  "a",
		B:  1,
		C:  2,
		D:  3,
		E:  4,
		F:  5,
		G:  0.6,
		H:  0.7,
		I:  8,
		J:  9,
		K:  10,
		L:  11,
		M:  12,
		N:  true,
		AA: []string{"a", "a"},
		BB: []int{1, 1},
		CC: []int8{2, 2},
		DD: []int16{3, 3},
		EE: []int32{4, 4},
		FF: []int64{5, 5},
		GG: []float32{0.6, 0.6},
		HH: []float64{0.7, 0.7},
		II: []uint{8, 8},
		JJ: []uint8{9, 9},
		KK: []uint16{10, 10},
		LL: []uint32{11, 11},
		MM: []uint64{12, 12},
		NN: []bool{true, true},
		X:  XXX(13),
		XX: []XXX{XXX(14), XXX(14)},
	}

	_resp := []interface{}([]interface{}{
		[]interface{}{
			map[string]interface{}{
				"outVLabel": "test",
				"type":      "edge",
				"id":        "e623ef5c-01f9-44f1-9684-f33c2e6598ee",
				"inV":       "d014ab68-fa70-4a6c-8f11-33fd3eef0112",
				"inVLabel":  "test",
				"label":     "relates",
				"outV":      "e3ff8f7d-0b29-4f4e-854a-affa3544b12a"}}})

	resp, err := g.AddE("relates", _t, _t2)
	assert.Nil(err)
	assert.Equal(_resp, resp)
}

func TestDropE(t *testing.T) {
	assert := assert.New(t)

	// Create test server with the mock handler.
	s := httptest.NewServer(http.HandlerFunc(mock))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// test connecting to the mock server
	g, errs := NewClient(NewClientConfig(u))
	assert.IsType(make(chan error), errs)
	assert.NotNil(g)
	assert.IsType(&Client{}, g)

	// setup err channel
	go func(chan error) {
		err := <-errs
		assert.Nil(err)
	}(errs)

	_tUUID, _ := uuid.Parse("64795211-c4a1-4eac-9e0a-b674ced77461")
	_t := Test{
		Id: _tUUID,
		A:  "aa",
		B:  10,
		C:  20,
		D:  30,
		E:  40,
		F:  50,
		G:  0.06,
		H:  0.07,
		I:  80,
		J:  90,
		K:  100,
		L:  110,
		M:  120,
		N:  true,
		AA: []string{"aa", "aa"},
		BB: []int{10, 10},
		CC: []int8{20, 20},
		DD: []int16{30, 30},
		EE: []int32{40, 40},
		FF: []int64{50, 50},
		GG: []float32{0.06, 0.06},
		HH: []float64{0.07, 0.07},
		II: []uint{80, 80},
		JJ: []uint8{90, 90},
		KK: []uint16{100, 100},
		LL: []uint32{110, 110},
		MM: []uint64{120, 120},
		NN: []bool{true, true},
		X:  XXX(130),
		XX: []XXX{XXX(140), XXX(140)},
	}

	_t2UUID, _ := uuid.Parse("dafeafc6-63a7-42b2-8ac2-4b85c3e2e37a")
	_t2 := Test{
		Id: _t2UUID,
		A:  "a",
		B:  1,
		C:  2,
		D:  3,
		E:  4,
		F:  5,
		G:  0.6,
		H:  0.7,
		I:  8,
		J:  9,
		K:  10,
		L:  11,
		M:  12,
		N:  true,
		AA: []string{"a", "a"},
		BB: []int{1, 1},
		CC: []int8{2, 2},
		DD: []int16{3, 3},
		EE: []int32{4, 4},
		FF: []int64{5, 5},
		GG: []float32{0.6, 0.6},
		HH: []float64{0.7, 0.7},
		II: []uint{8, 8},
		JJ: []uint8{9, 9},
		KK: []uint16{10, 10},
		LL: []uint32{11, 11},
		MM: []uint64{12, 12},
		NN: []bool{true, true},
		X:  XXX(13),
		XX: []XXX{XXX(14), XXX(14)},
	}

	_resp := []interface{}{[]interface{}{}}

	resp, err := g.DropE("relates", _t, _t2)
	assert.Nil(err)
	assert.Equal(_resp, resp)
}

func TestAddEById(t *testing.T) {
	assert := assert.New(t)

	// Create test server with the mock handler.
	s := httptest.NewServer(http.HandlerFunc(mock))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// test connecting to the mock server
	g, errs := NewClient(NewClientConfig(u))
	assert.IsType(make(chan error), errs)
	assert.NotNil(g)
	assert.IsType(&Client{}, g)

	// setup err channel
	go func(chan error) {
		err := <-errs
		assert.Nil(err)
	}(errs)

	_tUUID, _ := uuid.Parse("64795211-c4a1-4eac-9e0a-b674ced77461")
	_t2UUID, _ := uuid.Parse("dafeafc6-63a7-42b2-8ac2-4b85c3e2e37a")

	_resp := []interface{}([]interface{}{
		[]interface{}{
			map[string]interface{}{
				"outVLabel": "test",
				"type":      "edge",
				"id":        "e623ef5c-01f9-44f1-9684-f33c2e6598ee",
				"inV":       "d014ab68-fa70-4a6c-8f11-33fd3eef0112",
				"inVLabel":  "test",
				"label":     "relates",
				"outV":      "e3ff8f7d-0b29-4f4e-854a-affa3544b12a"}}})

	resp, err := g.AddEById("relates", _tUUID, _t2UUID)
	assert.Nil(err)
	assert.Equal(_resp, resp)
}

func TestDropEById(t *testing.T) {
	assert := assert.New(t)

	// Create test server with the mock handler.
	s := httptest.NewServer(http.HandlerFunc(mock))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// test connecting to the mock server
	g, errs := NewClient(NewClientConfig(u))
	assert.IsType(make(chan error), errs)
	assert.NotNil(g)
	assert.IsType(&Client{}, g)

	// setup err channel
	go func(chan error) {
		err := <-errs
		assert.Nil(err)
	}(errs)

	_tUUID, _ := uuid.Parse("64795211-c4a1-4eac-9e0a-b674ced77461")
	_t2UUID, _ := uuid.Parse("dafeafc6-63a7-42b2-8ac2-4b85c3e2e37a")

	_resp := []interface{}{[]interface{}{}}

	resp, err := g.DropEById("relates", _tUUID, _t2UUID)
	assert.Nil(err)
	assert.Equal(_resp, resp)
}

func TestGet(t *testing.T) {
	assert := assert.New(t)

	// Create test server with the mock handler.
	s := httptest.NewServer(http.HandlerFunc(mock))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// test connecting to the mock server
	g, errs := NewClient(NewClientConfig(u))
	assert.IsType(make(chan error), errs)
	assert.NotNil(g)
	assert.IsType(&Client{}, g)

	// setup err channel
	go func(chan error) {
		err := <-errs
		assert.Nil(err)
	}(errs)

	// create test struct to pass as interface to AddV
	_tUUID, _ := uuid.Parse("64795211-c4a1-4eac-9e0a-b674ced77461")
	_tUUID2, _ := uuid.Parse("96f7cacd-01fd-469e-a14c-5178903a39b6")
	_tUUID3, _ := uuid.Parse("4f9a2f8b-3b3a-4028-a1e0-fa55e0dd1543")
	_tUUID4, _ := uuid.Parse("593db951-3456-450a-a779-44f1a4bd330d")
	_t := Test{
		Id: _tUUID,
		A:  "aa",
		B:  10,
		C:  20,
		D:  30,
		E:  40,
		F:  50,
		G:  0.06,
		H:  0.07,
		I:  80,
		J:  90,
		K:  100,
		L:  110,
		M:  120,
		N:  true,
		AA: []string{"aa", "aa"},
		BB: []int{10, 10},
		CC: []int8{20, 20},
		DD: []int16{30, 30},
		EE: []int32{40, 40},
		FF: []int64{50, 50},
		GG: []float32{0.06, 0.06},
		HH: []float64{0.07, 0.07},
		II: []uint{80, 80},
		JJ: []uint8{90, 90},
		KK: []uint16{100, 100},
		LL: []uint32{110, 110},
		MM: []uint64{120, 120},
		NN: []bool{true, true},
		X:  XXX(130),
		XX: []XXX{XXX(140), XXX(140)},
		Z: Test2{Id: _tUUID2,
			A: "aa",
			B: 10,
		},
		ZZ: []Test2{
			Test2{Id: _tUUID3,
				A: "aa",
				B: 10,
			},
			Test2{Id: _tUUID4,
				A: "aa",
				B: 10,
			},
		},
	}

	_ts := []Test{}
	q := fmt.Sprintf("g.V('%s')", _t.Id)
	err := g.Get(q, &_ts)
	assert.Nil(err)
	assert.Equal(1, len(_ts))
	assert.Equal(_t, _ts[0])

	// test error not a slice
	_ts2 := Test{}
	err = g.Get(q, &_ts2)
	_err := errors.New("the passed interface is not a slice")
	assert.Equal(_err, err)

	// test error not a ptr
	_ts3 := []Test{}
	err = g.Get(q, _ts3)
	_err = errors.New("the passed interface is not a ptr")
	assert.Equal(_err, err)
}

func TestDisposed(t *testing.T) {
	assert := assert.New(t)

	// Create test server with the mock handler.
	s := httptest.NewServer(http.HandlerFunc(mock))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// test connecting to the mock server
	g, errs := NewClient(NewClientConfig(u))
	assert.IsType(make(chan error), errs)
	assert.NotNil(g)
	assert.IsType(&Client{}, g)

	// setup err channel
	go func(chan error) {
		err := <-errs
		_err := &net.OpError{}
		_errWs := &websocket.CloseError{}
		assert.True(assert.IsType(_err, err) || assert.IsType(_errWs, err))
	}(errs)

	// dispose connection
	err := g.conn.close()
	assert.Nil(err)
	q := "g.V()"
	_, err = g.Execute(q, nil, nil)
	_err := errors.New("you cannot write on a disposed connection")
	assert.Equal(_err, err)

	// create test struct to pass as interface to AddV
	_tUUID, _ := uuid.Parse("64795211-c4a1-4eac-9e0a-b674ced77461")
	_t := Test{
		Id: _tUUID,
		A:  "aa",
		B:  10,
		C:  20,
		D:  30,
		E:  40,
		F:  50,
		G:  0.06,
		H:  0.07,
		I:  80,
		J:  90,
		K:  100,
		L:  110,
		M:  120,
		N:  true,
		AA: []string{"aa", "aa"},
		BB: []int{10, 10},
		CC: []int8{20, 20},
		DD: []int16{30, 30},
		EE: []int32{40, 40},
		FF: []int64{50, 50},
		GG: []float32{0.06, 0.06},
		HH: []float64{0.07, 0.07},
		II: []uint{80, 80},
		JJ: []uint8{90, 90},
		KK: []uint16{100, 100},
		LL: []uint32{110, 110},
		MM: []uint64{120, 120},
		NN: []bool{true, true},
		X:  XXX(130),
		XX: []XXX{XXX(140), XXX(140)},
	}

	_t2UUID, _ := uuid.Parse("dafeafc6-63a7-42b2-8ac2-4b85c3e2e37a")
	_t2 := Test{
		Id: _t2UUID,
		A:  "a",
		B:  1,
		C:  2,
		D:  3,
		E:  4,
		F:  5,
		G:  0.6,
		H:  0.7,
		I:  8,
		J:  9,
		K:  10,
		L:  11,
		M:  12,
		N:  true,
		AA: []string{"a", "a"},
		BB: []int{1, 1},
		CC: []int8{2, 2},
		DD: []int16{3, 3},
		EE: []int32{4, 4},
		FF: []int64{5, 5},
		GG: []float32{0.6, 0.6},
		HH: []float64{0.7, 0.7},
		II: []uint{8, 8},
		JJ: []uint8{9, 9},
		KK: []uint16{10, 10},
		LL: []uint32{11, 11},
		MM: []uint64{12, 12},
		NN: []bool{true, true},
		X:  XXX(13),
		XX: []XXX{XXX(14), XXX(14)},
	}

	_ts := []Test{}
	q = fmt.Sprintf("g.V('%s')", &_t.Id)
	err = g.Get(q, _ts)
	assert.Equal(_err, err)

	_, err = g.AddV("test", _t)
	assert.Equal(_err, err)

	_, err = g.DropV(_t)
	assert.Equal(_err, err)

	_, err = g.UpdateV(_t)
	assert.Equal(_err, err)

	_, err = g.AddE("relates", _t, _t2)
	assert.Equal(_err, err)

	_, err = g.AddEById("relates", _t.Id, _t2.Id)
	assert.Equal(_err, err)

	_, err = g.AddEWithProps("relates", _t, _t2, nil)
	assert.Equal(_err, err)

	_, err = g.AddEWithPropsById("relates", _t.Id, _t2.Id, nil)
	assert.Equal(_err, err)

	_, err = g.DropE("relates", _t, _t2)
	assert.Equal(_err, err)

	_, err = g.DropEById("relates", _t.Id, _t2.Id)
	assert.Equal(_err, err)
}

func TestGetPropVal(t *testing.T) {
	assert := assert.New(t)

	var props map[string]interface{}
	maps := []byte(`{"foo":"bar","biz":3}`)
	err := json.Unmarshal(maps, &props)
	assert.Nil(err)

	p, err := getPropertyValue(props)
	assert.Equal(nil, p)
	_err := errors.New("passed property cannot be cast")
	assert.Equal(_err, err)
}

func TestBuildProps(t *testing.T) {
	assert := assert.New(t)

	var props map[string]interface{}
	maps := []byte(`{"foo":"bar","biz":3}`)
	err := json.Unmarshal(maps, &props)
	assert.Nil(err)
	p, err := buildProps(props)
	assert.Nil(err)
	_p := ".property('foo', 'bar').property('biz', 3)"
	_p2 := ".property('biz', 3).property('foo', 'bar')"
	assert.True(_p == p || _p2 == p)
}

func TestClientClose(t *testing.T) {
	assert := assert.New(t)

	// Create test server with the mock handler.
	s := httptest.NewServer(http.HandlerFunc(mock))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// test connecting to the mock server
	g, errs := NewClient(NewClientConfig(u))
	assert.IsType(make(chan error), errs)
	assert.NotNil(g)
	assert.IsType(&Client{}, g)

	// setup err channel
	go func(chan error) {
		err := <-errs
		_err := &net.OpError{}
		_errWs := &websocket.CloseError{}
		assert.True(assert.IsType(_err, err) || assert.IsType(_errWs, err))
	}(errs)

	g.Close()
}

func TestAddEWithProps(t *testing.T) {
	assert := assert.New(t)

	// Create test server with the mock handler.
	s := httptest.NewServer(http.HandlerFunc(mock))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// test connecting to the mock server
	g, errs := NewClient(NewClientConfig(u))
	assert.IsType(make(chan error), errs)
	assert.NotNil(g)
	assert.IsType(&Client{}, g)

	// setup err channel
	go func(chan error) {
		err := <-errs
		assert.Nil(err)
	}(errs)

	_tUUID, _ := uuid.Parse("64795211-c4a1-4eac-9e0a-b674ced77461")
	_t := Test{
		Id: _tUUID,
		A:  "aa",
		B:  10,
		C:  20,
		D:  30,
		E:  40,
		F:  50,
		G:  0.06,
		H:  0.07,
		I:  80,
		J:  90,
		K:  100,
		L:  110,
		M:  120,
		N:  true,
		AA: []string{"aa", "aa"},
		BB: []int{10, 10},
		CC: []int8{20, 20},
		DD: []int16{30, 30},
		EE: []int32{40, 40},
		FF: []int64{50, 50},
		GG: []float32{0.06, 0.06},
		HH: []float64{0.07, 0.07},
		II: []uint{80, 80},
		JJ: []uint8{90, 90},
		KK: []uint16{100, 100},
		LL: []uint32{110, 110},
		MM: []uint64{120, 120},
		NN: []bool{true, true},
		X:  XXX(130),
		XX: []XXX{XXX(140), XXX(140)},
	}

	_t2UUID, _ := uuid.Parse("dafeafc6-63a7-42b2-8ac2-4b85c3e2e37a")
	_t2 := Test{
		Id: _t2UUID,
		A:  "a",
		B:  1,
		C:  2,
		D:  3,
		E:  4,
		F:  5,
		G:  0.6,
		H:  0.7,
		I:  8,
		J:  9,
		K:  10,
		L:  11,
		M:  12,
		N:  true,
		AA: []string{"a", "a"},
		BB: []int{1, 1},
		CC: []int8{2, 2},
		DD: []int16{3, 3},
		EE: []int32{4, 4},
		FF: []int64{5, 5},
		GG: []float32{0.6, 0.6},
		HH: []float64{0.7, 0.7},
		II: []uint{8, 8},
		JJ: []uint8{9, 9},
		KK: []uint16{10, 10},
		LL: []uint32{11, 11},
		MM: []uint64{12, 12},
		NN: []bool{true, true},
		X:  XXX(13),
		XX: []XXX{XXX(14), XXX(14)},
	}

	_resp := []interface{}([]interface{}{
		[]interface{}{
			map[string]interface{}{
				"outVLabel":  "test",
				"type":       "edge",
				"id":         "e623ef5c-01f9-44f1-9684-f33c2e6598ee",
				"inV":        "d014ab68-fa70-4a6c-8f11-33fd3eef0112",
				"inVLabel":   "test",
				"label":      "relates",
				"outV":       "e3ff8f7d-0b29-4f4e-854a-affa3544b12a",
				"properties": map[string]interface{}{"biz": float64(3), "foo": "bar"},
			}}})

	var props map[string]interface{}
	maps := []byte(`{"foo":"bar","biz":3}`)
	err := json.Unmarshal(maps, &props)
	assert.Nil(err)
	resp, err := g.AddEWithProps("relates", _t, _t2, props)
	assert.Nil(err)
	assert.Equal(_resp, resp)
}

func TestAddEWithPropsById(t *testing.T) {
	assert := assert.New(t)

	// Create test server with the mock handler.
	s := httptest.NewServer(http.HandlerFunc(mock))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// test connecting to the mock server
	g, errs := NewClient(NewClientConfig(u))
	assert.IsType(make(chan error), errs)
	assert.NotNil(g)
	assert.IsType(&Client{}, g)

	// setup err channel
	go func(chan error) {
		err := <-errs
		assert.Nil(err)
	}(errs)

	_tUUID, _ := uuid.Parse("64795211-c4a1-4eac-9e0a-b674ced77461")
	_t2UUID, _ := uuid.Parse("dafeafc6-63a7-42b2-8ac2-4b85c3e2e37a")

	_resp := []interface{}([]interface{}{
		[]interface{}{
			map[string]interface{}{
				"outVLabel":  "test",
				"type":       "edge",
				"id":         "e623ef5c-01f9-44f1-9684-f33c2e6598ee",
				"inV":        "d014ab68-fa70-4a6c-8f11-33fd3eef0112",
				"inVLabel":   "test",
				"label":      "relates",
				"outV":       "e3ff8f7d-0b29-4f4e-854a-affa3544b12a",
				"properties": map[string]interface{}{"biz": float64(3), "foo": "bar"},
			}}})

	var props map[string]interface{}
	maps := []byte(`{"baz":["foo","bar"]}`)
	err := json.Unmarshal(maps, &props)
	assert.Nil(err)
	resp, err := g.AddEWithPropsById("relates", _tUUID, _t2UUID, props)
	assert.Nil(err)
	assert.Equal(_resp, resp)

}
