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

	_tResp := []*GremlinData{
		&GremlinData{
			Id:         uuid.UUID{0xc1, 0xf7, 0xa9, 0x21, 0xb7, 0x67, 0x48, 0x39, 0xbb, 0xdc, 0x64, 0x78, 0xeb, 0x5f, 0x34, 0x54},
			Label:      "test",
			Type:       "vertex",
			InVLablel:  "",
			OutVLablel: "",
			InV:        uuid.UUID{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			OutV:       uuid.UUID{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			Properties: map[string]interface{}{
				"kk": []interface{}{map[string]interface{}{"value": float64(100), "id": "740dca9e-c5d7-40d3-8d68-feb48090a638"}, map[string]interface{}{"id": "a0587e7d-6bf5-4cba-b911-0507d0469068", "value": float64(100)}},
				"cc": []interface{}{map[string]interface{}{"id": "8f16c7cd-4125-4d29-b714-5d8f561bb8e4", "value": float64(20)}, map[string]interface{}{"id": "84f44c14-f038-47b3-a0b7-9f14dd11ddde", "value": float64(20)}},
				"jj": []interface{}{map[string]interface{}{"id": "84e17824-94be-47e9-b7bb-7e46ed5c065f", "value": float64(90)}, map[string]interface{}{"id": "20dffbe0-4c63-4b15-9b5d-43111bd10525", "value": float64(90)}},
				"i":  []interface{}{map[string]interface{}{"id": "2295b3de-fc5b-42c9-8b04-44b12fbe1346", "value": float64(80)}},
				"aa": []interface{}{map[string]interface{}{"id": "225ed5a7-b000-4a59-b6c3-332682a5216a", "value": "aa"}, map[string]interface{}{"id": "9cbee039-c5b4-4e75-a1b0-346a47e5dc36", "value": "aa"}},
				"c":  []interface{}{map[string]interface{}{"value": float64(20), "id": "8e58d327-e06b-44e4-a5d9-75558cdca2dc"}},
				"ii": []interface{}{map[string]interface{}{"id": "fe3a148a-4a80-4ca2-851b-5dc473f549e6", "value": float64(80)}, map[string]interface{}{"id": "021a9a1a-49c1-4ae3-aa40-3f5c33a12e9f", "value": float64(80)}},
				"m":  []interface{}{map[string]interface{}{"value": float64(120), "id": "a4b507fe-bb16-4b4a-aa1d-cf922af67cd2"}},
				"x":  []interface{}{map[string]interface{}{"id": "56b71ade-d0aa-416f-8da3-517391fd7ee4", "value": float64(130)}},
				"b":  []interface{}{map[string]interface{}{"value": float64(10), "id": "91df576d-3501-4303-9d89-1c8409ce6ff4"}},
				"dd": []interface{}{map[string]interface{}{"id": "51336fa2-ebf9-4d7a-9ec5-0128a6341ea6", "value": float64(30)}, map[string]interface{}{"id": "337fe24f-8ea3-40ea-b726-e7169883618b", "value": float64(30)}},
				"h":  []interface{}{map[string]interface{}{"id": "3e16edbc-5a77-4b97-bdb4-4695996d8915", "value": float64(0.07)}},
				"j":  []interface{}{map[string]interface{}{"id": "e7235d64-4212-448d-869f-612cf2403b96", "value": float64(90)}},
				"xx": []interface{}{map[string]interface{}{"id": "9cf5c2a7-45eb-4e58-bf5a-9f15186c0819", "value": float64(140)}, map[string]interface{}{"id": "122191a7-5437-45ae-9ec6-1a73fee5c996", "value": float64(140)}},
				"ee": []interface{}{map[string]interface{}{"value": float64(40), "id": "8d21dad3-5ed7-4925-bfae-12b345592a36"}, map[string]interface{}{"id": "6277ad55-a3ed-41bc-8ad2-1cfe6e60938b", "value": float64(40)}},
				"f":  []interface{}{map[string]interface{}{"id": "ed8cf6a7-d585-4575-a08c-cf4aa27f1491", "value": float64(50)}},
				"k":  []interface{}{map[string]interface{}{"id": "a435d9ad-9f8f-43d6-b108-a2fe1d5a95b9", "value": float64(100)}},
				"l":  []interface{}{map[string]interface{}{"id": "e7c6ddee-729e-44a4-b977-7d3eafe47497", "value": float64(110)}},
				"nn": []interface{}{map[string]interface{}{"id": "7af1d164-3966-4dde-93fa-511a936601f5", "value": true}, map[string]interface{}{"id": "90be7e5c-8bf6-4bfd-bd01-38be1697d9f8", "value": true}},
				"d":  []interface{}{map[string]interface{}{"id": "565d7400-e75b-4813-aa39-6c09cae781a8", "value": float64(30)}},
				"g":  []interface{}{map[string]interface{}{"value": float64(0.06), "id": "2c6860ac-7151-48f9-b866-5b40a3488d1e"}},
				"n":  []interface{}{map[string]interface{}{"id": "954cc7f9-d655-4123-a66d-e3e665cf7d49", "value": true}},
				"bb": []interface{}{map[string]interface{}{"id": "b96f76ed-028a-4e2f-942e-2adf37f5bcb0", "value": float64(10)}, map[string]interface{}{"id": "7f010e2c-b764-4601-b190-4b34372203e7", "value": float64(10)}},
				"hh": []interface{}{map[string]interface{}{"value": float64(0.07), "id": "b2266aff-1f18-4391-98f8-5ad6a542a2e1"}, map[string]interface{}{"id": "709a88dd-8fc4-4c8d-bd31-61962feff9b2", "value": float64(0.07)}},
				"ff": []interface{}{map[string]interface{}{"id": "4c2afab4-df69-490b-8cf0-c8311808c0fc", "value": float64(50)}, map[string]interface{}{"id": "38e68ad4-b9f8-4256-9638-1e52cdbb989a", "value": float64(50)}},
				"gg": []interface{}{map[string]interface{}{"value": float64(0.06), "id": "a0a129f4-bfa4-4d1a-bc77-df5b63049197"}, map[string]interface{}{"id": "3c8b51c4-8b40-4d5f-a79d-6a9d74929837", "value": float64(0.06)}},
				"ll": []interface{}{map[string]interface{}{"id": "ff4d7387-b3f7-41ec-9cee-912bb9220545", "value": float64(110)}, map[string]interface{}{"id": "519d212c-4774-49a1-bc4e-3715af929c38", "value": float64(110)}},
				"mm": []interface{}{map[string]interface{}{"id": "0487cc86-ee49-4649-8131-d43610235c40", "value": float64(120)}, map[string]interface{}{"id": "b6e9c4c6-8ac4-4124-92cd-e53acf0cfd12", "value": float64(120)}},
				"a":  []interface{}{map[string]interface{}{"value": "aa", "id": "15d0a33b-d369-4b61-b162-320ece53cfa1"}},
				"e":  []interface{}{map[string]interface{}{"id": "bf35b756-0640-48c6-9601-aab77c6aa603", "value": float64(40)}}},
		},
	}

	resp, err := g.AddV("test", _t)
	assert.Nil(err)
	assert.Equal(_tResp, resp)
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

	_tResp := []*GremlinData{
		&GremlinData{
			Id:         uuid.UUID{0xc1, 0xf7, 0xa9, 0x21, 0xb7, 0x67, 0x48, 0x39, 0xbb, 0xdc, 0x64, 0x78, 0xeb, 0x5f, 0x34, 0x54},
			Label:      "test",
			Type:       "vertex",
			InVLablel:  "",
			OutVLablel: "",
			InV:        uuid.UUID{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			OutV:       uuid.UUID{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
			Properties: map[string]interface{}{
				"kk": []interface{}{map[string]interface{}{"value": float64(100), "id": "740dca9e-c5d7-40d3-8d68-feb48090a638"}, map[string]interface{}{"id": "a0587e7d-6bf5-4cba-b911-0507d0469068", "value": float64(100)}},
				"cc": []interface{}{map[string]interface{}{"id": "8f16c7cd-4125-4d29-b714-5d8f561bb8e4", "value": float64(20)}, map[string]interface{}{"id": "84f44c14-f038-47b3-a0b7-9f14dd11ddde", "value": float64(20)}},
				"jj": []interface{}{map[string]interface{}{"id": "84e17824-94be-47e9-b7bb-7e46ed5c065f", "value": float64(90)}, map[string]interface{}{"id": "20dffbe0-4c63-4b15-9b5d-43111bd10525", "value": float64(90)}},
				"i":  []interface{}{map[string]interface{}{"id": "2295b3de-fc5b-42c9-8b04-44b12fbe1346", "value": float64(80)}},
				"aa": []interface{}{map[string]interface{}{"id": "225ed5a7-b000-4a59-b6c3-332682a5216a", "value": "aa"}, map[string]interface{}{"id": "9cbee039-c5b4-4e75-a1b0-346a47e5dc36", "value": "aa"}},
				"c":  []interface{}{map[string]interface{}{"value": float64(20), "id": "8e58d327-e06b-44e4-a5d9-75558cdca2dc"}},
				"ii": []interface{}{map[string]interface{}{"id": "fe3a148a-4a80-4ca2-851b-5dc473f549e6", "value": float64(80)}, map[string]interface{}{"id": "021a9a1a-49c1-4ae3-aa40-3f5c33a12e9f", "value": float64(80)}},
				"m":  []interface{}{map[string]interface{}{"value": float64(120), "id": "a4b507fe-bb16-4b4a-aa1d-cf922af67cd2"}},
				"x":  []interface{}{map[string]interface{}{"id": "56b71ade-d0aa-416f-8da3-517391fd7ee4", "value": float64(130)}},
				"b":  []interface{}{map[string]interface{}{"value": float64(10), "id": "91df576d-3501-4303-9d89-1c8409ce6ff4"}},
				"dd": []interface{}{map[string]interface{}{"id": "51336fa2-ebf9-4d7a-9ec5-0128a6341ea6", "value": float64(30)}, map[string]interface{}{"id": "337fe24f-8ea3-40ea-b726-e7169883618b", "value": float64(30)}},
				"h":  []interface{}{map[string]interface{}{"id": "3e16edbc-5a77-4b97-bdb4-4695996d8915", "value": float64(0.07)}},
				"j":  []interface{}{map[string]interface{}{"id": "e7235d64-4212-448d-869f-612cf2403b96", "value": float64(90)}},
				"xx": []interface{}{map[string]interface{}{"id": "9cf5c2a7-45eb-4e58-bf5a-9f15186c0819", "value": float64(140)}, map[string]interface{}{"id": "122191a7-5437-45ae-9ec6-1a73fee5c996", "value": float64(140)}},
				"ee": []interface{}{map[string]interface{}{"value": float64(40), "id": "8d21dad3-5ed7-4925-bfae-12b345592a36"}, map[string]interface{}{"id": "6277ad55-a3ed-41bc-8ad2-1cfe6e60938b", "value": float64(40)}},
				"f":  []interface{}{map[string]interface{}{"id": "ed8cf6a7-d585-4575-a08c-cf4aa27f1491", "value": float64(50)}},
				"k":  []interface{}{map[string]interface{}{"id": "a435d9ad-9f8f-43d6-b108-a2fe1d5a95b9", "value": float64(100)}},
				"l":  []interface{}{map[string]interface{}{"id": "e7c6ddee-729e-44a4-b977-7d3eafe47497", "value": float64(110)}},
				"nn": []interface{}{map[string]interface{}{"id": "7af1d164-3966-4dde-93fa-511a936601f5", "value": true}, map[string]interface{}{"id": "90be7e5c-8bf6-4bfd-bd01-38be1697d9f8", "value": true}},
				"d":  []interface{}{map[string]interface{}{"id": "565d7400-e75b-4813-aa39-6c09cae781a8", "value": float64(30)}},
				"g":  []interface{}{map[string]interface{}{"value": float64(0.06), "id": "2c6860ac-7151-48f9-b866-5b40a3488d1e"}},
				"n":  []interface{}{map[string]interface{}{"id": "954cc7f9-d655-4123-a66d-e3e665cf7d49", "value": true}},
				"bb": []interface{}{map[string]interface{}{"id": "b96f76ed-028a-4e2f-942e-2adf37f5bcb0", "value": float64(10)}, map[string]interface{}{"id": "7f010e2c-b764-4601-b190-4b34372203e7", "value": float64(10)}},
				"hh": []interface{}{map[string]interface{}{"value": float64(0.07), "id": "b2266aff-1f18-4391-98f8-5ad6a542a2e1"}, map[string]interface{}{"id": "709a88dd-8fc4-4c8d-bd31-61962feff9b2", "value": float64(0.07)}},
				"ff": []interface{}{map[string]interface{}{"id": "4c2afab4-df69-490b-8cf0-c8311808c0fc", "value": float64(50)}, map[string]interface{}{"id": "38e68ad4-b9f8-4256-9638-1e52cdbb989a", "value": float64(50)}},
				"gg": []interface{}{map[string]interface{}{"value": float64(0.06), "id": "a0a129f4-bfa4-4d1a-bc77-df5b63049197"}, map[string]interface{}{"id": "3c8b51c4-8b40-4d5f-a79d-6a9d74929837", "value": float64(0.06)}},
				"ll": []interface{}{map[string]interface{}{"id": "ff4d7387-b3f7-41ec-9cee-912bb9220545", "value": float64(110)}, map[string]interface{}{"id": "519d212c-4774-49a1-bc4e-3715af929c38", "value": float64(110)}},
				"mm": []interface{}{map[string]interface{}{"id": "0487cc86-ee49-4649-8131-d43610235c40", "value": float64(120)}, map[string]interface{}{"id": "b6e9c4c6-8ac4-4124-92cd-e53acf0cfd12", "value": float64(120)}},
				"a":  []interface{}{map[string]interface{}{"value": "aa", "id": "15d0a33b-d369-4b61-b162-320ece53cfa1"}},
				"e":  []interface{}{map[string]interface{}{"id": "bf35b756-0640-48c6-9601-aab77c6aa603", "value": float64(40)}}},
		},
	}

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

	_tResp := []*GremlinData(nil)
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

	_resp := []*GremlinData{
		&GremlinData{
			Id:         uuid.UUID{0xe6, 0x23, 0xef, 0x5c, 0x1, 0xf9, 0x44, 0xf1, 0x96, 0x84, 0xf3, 0x3c, 0x2e, 0x65, 0x98, 0xee},
			Label:      "relates",
			Type:       "edge",
			InVLablel:  "test",
			OutVLablel: "test",
			InV:        uuid.UUID{0xd0, 0x14, 0xab, 0x68, 0xfa, 0x70, 0x4a, 0x6c, 0x8f, 0x11, 0x33, 0xfd, 0x3e, 0xef, 0x1, 0x12},
			OutV:       uuid.UUID{0xe3, 0xff, 0x8f, 0x7d, 0xb, 0x29, 0x4f, 0x4e, 0x85, 0x4a, 0xaf, 0xfa, 0x35, 0x44, 0xb1, 0x2a},
			Properties: map[string]interface{}(nil),
		},
	}

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

	_resp := []*GremlinData(nil)

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

	_resp := []*GremlinData{
		&GremlinData{
			Id:         uuid.UUID{0xe6, 0x23, 0xef, 0x5c, 0x1, 0xf9, 0x44, 0xf1, 0x96, 0x84, 0xf3, 0x3c, 0x2e, 0x65, 0x98, 0xee},
			Label:      "relates",
			Type:       "edge",
			InVLablel:  "test",
			OutVLablel: "test",
			InV:        uuid.UUID{0xd0, 0x14, 0xab, 0x68, 0xfa, 0x70, 0x4a, 0x6c, 0x8f, 0x11, 0x33, 0xfd, 0x3e, 0xef, 0x1, 0x12},
			OutV:       uuid.UUID{0xe3, 0xff, 0x8f, 0x7d, 0xb, 0x29, 0x4f, 0x4e, 0x85, 0x4a, 0xaf, 0xfa, 0x35, 0x44, 0xb1, 0x2a},
			Properties: map[string]interface{}(nil),
		},
	}

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

	_resp := []*GremlinData(nil)

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
	_err := ErrorConnectionDisposed
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
	assert.Equal(ErrorCannotCastProperty, err)
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

	_resp := []*GremlinData{
		&GremlinData{
			Id:         uuid.UUID{0xe6, 0x23, 0xef, 0x5c, 0x1, 0xf9, 0x44, 0xf1, 0x96, 0x84, 0xf3, 0x3c, 0x2e, 0x65, 0x98, 0xee},
			Label:      "relates",
			Type:       "edge",
			InVLablel:  "test",
			OutVLablel: "test",
			InV:        uuid.UUID{0xd0, 0x14, 0xab, 0x68, 0xfa, 0x70, 0x4a, 0x6c, 0x8f, 0x11, 0x33, 0xfd, 0x3e, 0xef, 0x1, 0x12},
			OutV:       uuid.UUID{0xe3, 0xff, 0x8f, 0x7d, 0xb, 0x29, 0x4f, 0x4e, 0x85, 0x4a, 0xaf, 0xfa, 0x35, 0x44, 0xb1, 0x2a},
			Properties: map[string]interface{}{"foo": "bar", "biz": float64(3)},
		},
	}

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

	_resp := []*GremlinData{
		&GremlinData{
			Id:         uuid.UUID{0xe6, 0x23, 0xef, 0x5c, 0x1, 0xf9, 0x44, 0xf1, 0x96, 0x84, 0xf3, 0x3c, 0x2e, 0x65, 0x98, 0xee},
			Label:      "relates",
			Type:       "edge",
			InVLablel:  "test",
			OutVLablel: "test",
			InV:        uuid.UUID{0xd0, 0x14, 0xab, 0x68, 0xfa, 0x70, 0x4a, 0x6c, 0x8f, 0x11, 0x33, 0xfd, 0x3e, 0xef, 0x1, 0x12},
			OutV:       uuid.UUID{0xe3, 0xff, 0x8f, 0x7d, 0xb, 0x29, 0x4f, 0x4e, 0x85, 0x4a, 0xaf, 0xfa, 0x35, 0x44, 0xb1, 0x2a},
			Properties: map[string]interface{}{"biz": float64(3), "foo": "bar"},
		},
	}

	var props map[string]interface{}
	maps := []byte(`{"baz":["foo","bar"]}`)
	err := json.Unmarshal(maps, &props)
	assert.Nil(err)
	resp, err := g.AddEWithPropsById("relates", _tUUID, _t2UUID, props)
	assert.Nil(err)
	assert.Equal(_resp, resp)

}
