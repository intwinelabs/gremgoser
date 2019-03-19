package gremgoser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

type gremlinResponse struct {
	RequestId string `json:"requestId"`
	Status    struct {
		Code       int `json:"code"`
		Attributes struct {
			XMsStatusCode                 int     `json:"x-ms-status-code"`
			XMsRequestCharge              float64 `json:"x-ms-request-charge"`
			XMsTotalRequestCharge         float64 `json:"x-ms-total-request-charge"`
			XMsCosmosdbGraphRequestCharge float64 `json:"x-ms-cosmosdb-graph-request-charge"`
			StorageRU                     float64 `json:"StorageRU"`
		} `json:"attributes"`
		Message string `json:"message"`
	} `json:"status"`
	Result struct {
		Data []interface{} `json:"data"`
		Meta struct {
		} `json:"meta"`
	} `json:"result"`
}

var upgrader = websocket.Upgrader{}

var gremV = `g.V()`

var gremGet = `g.V('64795211-c4a1-4eac-9e0a-b674ced77461')`

var gremV1 = `g.addV('test').property('id', '64795211-c4a1-4eac-9e0a-b674ced77461').property('a', 'aa').property('b', 10).property('c', 20).property('d', 30).property('e', 40).property('f', 50).property('g', 0.06).property('h', 0.07).property('i', 80).property('j', 90).property('k', 100).property('l', 110).property('m', 120).property('n', true).property('aa', 'aa').property('aa', 'aa').property('bb', 10).property('bb', 10).property('cc', 20).property('cc', 20).property('dd', 30).property('dd', 30).property('ee', 40).property('ee', 40).property('ff', 50).property('ff', 50).property('gg', 0.06).property('gg', 0.06).property('hh', 0.07).property('hh', 0.07).property('ii', 80).property('ii', 80).property('jj', 90).property('jj', 90).property('kk', 100).property('kk', 100).property('ll', 110).property('ll', 110).property('mm', 120).property('mm', 120).property('nn', true).property('nn', true).property('x', 130).property('xx', 140).property('xx', 140).property('z', '{"Id":"64795211-c4a1-4eac-9e0a-b674ced77461","A":"aa","B":10}').property('zz', '[{"Id":"64795211-c4a1-4eac-9e0a-b674ced77461","A":"aa","B":10},{"Id":"64795211-c4a1-4eac-9e0a-b674ced77461","A":"aa","B":10}]')`

var gremUpdateV1 = `g.V('64795211-c4a1-4eac-9e0a-b674ced77461').property('id', '64795211-c4a1-4eac-9e0a-b674ced77461').property('a', 'aa').property('b', 10).property('c', 20).property('d', 30).property('e', 40).property('f', 50).property('g', 0.06).property('h', 0.07).property('i', 80).property('j', 90).property('k', 100).property('l', 110).property('m', 120).property('n', true).property('aa', 'aa').property('aa', 'aa').property('bb', 10).property('bb', 10).property('cc', 20).property('cc', 20).property('dd', 30).property('dd', 30).property('ee', 40).property('ee', 40).property('ff', 50).property('ff', 50).property('gg', 0.06).property('gg', 0.06).property('hh', 0.07).property('hh', 0.07).property('ii', 80).property('ii', 80).property('jj', 90).property('jj', 90).property('kk', 100).property('kk', 100).property('ll', 110).property('ll', 110).property('mm', 120).property('mm', 120).property('nn', true).property('nn', true).property('x', 130).property('xx', 140).property('xx', 140).property('z', '{"Id":"64795211-c4a1-4eac-9e0a-b674ced77461","A":"aa","B":10}').property('zz', '[{"Id":"64795211-c4a1-4eac-9e0a-b674ced77461","A":"aa","B":10},{"Id":"64795211-c4a1-4eac-9e0a-b674ced77461","A":"aa","B":10}]')`

var gremDropV1 = `g.V('64795211-c4a1-4eac-9e0a-b674ced77461').drop()`

var gremV2 = `g.addV('test').property('id', 'dafeafc6-63a7-42b2-8ac2-4b85c3e2e37a').property('a', 'a').property('b', 1).property('c', 2).property('d', 3).property('e', 4).property('f', 5).property('g', 0.6).property('h', 0.7).property('i', 8).property('j', 9).property('k', 10).property('l', 11).property('m', 12).property('n', true).property('aa', 'a').property('aa', 'a').property('bb', 1).property('bb', 1).property('cc', 2).property('cc', 2).property('dd', 3).property('dd', 3).property('ee', 4).property('ee', 4).property('ff', 5).property('ff', 5).property('gg', 0.6).property('gg', 0.6).property('hh', 0.7).property('hh', 0.7).property('ii', 8).property('ii', 8).property('jj', 9).property('jj', 9).property('kk', 10).property('kk', 10).property('ll', 11).property('ll', 11).property('mm', 12).property('mm', 12).property('nn', true).property('nn', true).property('x', 13).property('xx', 14).property('xx', 14)`

var gremE = `g.V('64795211-c4a1-4eac-9e0a-b674ced77461').addE('relates').to(g.V('dafeafc6-63a7-42b2-8ac2-4b85c3e2e37a'))`

var gremEWithProps = `g.V('64795211-c4a1-4eac-9e0a-b674ced77461').addE('relates').to(g.V('dafeafc6-63a7-42b2-8ac2-4b85c3e2e37a')).property('foo', 'bar').property('biz', 3)`

var gremEWithProps2 = `g.V('64795211-c4a1-4eac-9e0a-b674ced77461').addE('relates').to(g.V('dafeafc6-63a7-42b2-8ac2-4b85c3e2e37a')).property('biz', 3).property('foo', 'bar')`

var gremEWithPropsSlice = `g.V('64795211-c4a1-4eac-9e0a-b674ced77461').addE('relates').to(g.V('dafeafc6-63a7-42b2-8ac2-4b85c3e2e37a')).property('baz', 'foo').property('baz', 'bar')`

var gremDropE = `g.V('64795211-c4a1-4eac-9e0a-b674ced77461').outE('relates').and(inV().is('dafeafc6-63a7-42b2-8ac2-4b85c3e2e37a')).drop()`

var authResp = `{"requestId":"a48a4f3c-f356-4a74-82c2-bdc5980b6495","status":{"code":407,"attributes":{"x-ms-status-code":407},"message":"Graph Service requires Gremlin Client to provide SASL Authentication."},"result":{"data":null,"meta":{}}}`

var vResp = `{"requestId":"a48a4f3c-f356-4a74-82c2-bdc5980b6495","status":{"code":200,"attributes":{"x-ms-status-code":200,"x-ms-request-charge":2.6199999999999997,"x-ms-total-request-charge":2.6199999999999997,"x-ms-cosmosdb-graph-request-charge":2.6199999999999997,"StorageRU":2.6199999999999997},"message":""},"result":{"data":[],"meta":{}}}`

var addV1Resp = `{"requestId":"e90163ec-e725-403f-b576-3e5d78ee6074","status":{"code":200,"attributes":{"x-ms-status-code":200,"x-ms-request-charge":83.89,"x-ms-total-request-charge":83.89,"x-ms-cosmosdb-graph-request-charge":83.89,"StorageRU":83.89},"message":""},"result":{"data":[{"id":"c1f7a921-b767-4839-bbdc-6478eb5f3454","label":"test","type":"vertex","properties":{"a":[{"id":"15d0a33b-d369-4b61-b162-320ece53cfa1","value":"aa"}],"b":[{"id":"91df576d-3501-4303-9d89-1c8409ce6ff4","value":10}],"c":[{"id":"8e58d327-e06b-44e4-a5d9-75558cdca2dc","value":20}],"d":[{"id":"565d7400-e75b-4813-aa39-6c09cae781a8","value":30}],"e":[{"id":"bf35b756-0640-48c6-9601-aab77c6aa603","value":40}],"f":[{"id":"ed8cf6a7-d585-4575-a08c-cf4aa27f1491","value":50}],"g":[{"id":"2c6860ac-7151-48f9-b866-5b40a3488d1e","value":0.06}],"h":[{"id":"3e16edbc-5a77-4b97-bdb4-4695996d8915","value":0.07}],"i":[{"id":"2295b3de-fc5b-42c9-8b04-44b12fbe1346","value":80}],"j":[{"id":"e7235d64-4212-448d-869f-612cf2403b96","value":90}],"k":[{"id":"a435d9ad-9f8f-43d6-b108-a2fe1d5a95b9","value":100}],"l":[{"id":"e7c6ddee-729e-44a4-b977-7d3eafe47497","value":110}],"m":[{"id":"a4b507fe-bb16-4b4a-aa1d-cf922af67cd2","value":120}],"n":[{"id":"954cc7f9-d655-4123-a66d-e3e665cf7d49","value":true}],"aa":[{"id":"225ed5a7-b000-4a59-b6c3-332682a5216a","value":"aa"},{"id":"9cbee039-c5b4-4e75-a1b0-346a47e5dc36","value":"aa"}],"bb":[{"id":"b96f76ed-028a-4e2f-942e-2adf37f5bcb0","value":10},{"id":"7f010e2c-b764-4601-b190-4b34372203e7","value":10}],"cc":[{"id":"8f16c7cd-4125-4d29-b714-5d8f561bb8e4","value":20},{"id":"84f44c14-f038-47b3-a0b7-9f14dd11ddde","value":20}],"dd":[{"id":"51336fa2-ebf9-4d7a-9ec5-0128a6341ea6","value":30},{"id":"337fe24f-8ea3-40ea-b726-e7169883618b","value":30}],"ee":[{"id":"8d21dad3-5ed7-4925-bfae-12b345592a36","value":40},{"id":"6277ad55-a3ed-41bc-8ad2-1cfe6e60938b","value":40}],"ff":[{"id":"4c2afab4-df69-490b-8cf0-c8311808c0fc","value":50},{"id":"38e68ad4-b9f8-4256-9638-1e52cdbb989a","value":50}],"gg":[{"id":"a0a129f4-bfa4-4d1a-bc77-df5b63049197","value":0.06},{"id":"3c8b51c4-8b40-4d5f-a79d-6a9d74929837","value":0.06}],"hh":[{"id":"b2266aff-1f18-4391-98f8-5ad6a542a2e1","value":0.07},{"id":"709a88dd-8fc4-4c8d-bd31-61962feff9b2","value":0.07}],"ii":[{"id":"fe3a148a-4a80-4ca2-851b-5dc473f549e6","value":80},{"id":"021a9a1a-49c1-4ae3-aa40-3f5c33a12e9f","value":80}],"jj":[{"id":"84e17824-94be-47e9-b7bb-7e46ed5c065f","value":90},{"id":"20dffbe0-4c63-4b15-9b5d-43111bd10525","value":90}],"kk":[{"id":"740dca9e-c5d7-40d3-8d68-feb48090a638","value":100},{"id":"a0587e7d-6bf5-4cba-b911-0507d0469068","value":100}],"ll":[{"id":"ff4d7387-b3f7-41ec-9cee-912bb9220545","value":110},{"id":"519d212c-4774-49a1-bc4e-3715af929c38","value":110}],"mm":[{"id":"0487cc86-ee49-4649-8131-d43610235c40","value":120},{"id":"b6e9c4c6-8ac4-4124-92cd-e53acf0cfd12","value":120}],"nn":[{"id":"7af1d164-3966-4dde-93fa-511a936601f5","value":true},{"id":"90be7e5c-8bf6-4bfd-bd01-38be1697d9f8","value":true}],"x":[{"id":"56b71ade-d0aa-416f-8da3-517391fd7ee4","value":130}],"xx":[{"id":"9cf5c2a7-45eb-4e58-bf5a-9f15186c0819","value":140},{"id":"122191a7-5437-45ae-9ec6-1a73fee5c996","value":140}]}}],"meta":{}}}`

var addV2Resp = `{"requestId":"25318b9c-aa70-41c8-a7ff-ef206d5d472f","status":{"code":200,"attributes":{"x-ms-status-code":200,"x-ms-request-charge":83.89,"x-ms-total-request-charge":83.89,"x-ms-cosmosdb-graph-request-charge":83.89,"StorageRU":83.89},"message":""},"result":{"data":[{"id":"d0cfb441-5e49-488c-9d35-cf5f227ebc35","label":"test","type":"vertex","properties":{"a":[{"id":"cbe825ea-162a-4e15-9c4b-445ba55dd3a0","value":"a"}],"b":[{"id":"3f5e82fe-ce63-4e6d-b2fb-810b9c37a1d4","value":1}],"c":[{"id":"3806f808-05ff-45db-9f63-21fe361068f5","value":2}],"d":[{"id":"9abab188-07f9-4177-b89f-5fe844ff0611","value":3}],"e":[{"id":"4399bf7e-3759-4aa4-b944-28ea387a8512","value":4}],"f":[{"id":"0b339725-3d3c-404d-bf33-7bc003760282","value":5}],"g":[{"id":"5ab6a030-aa52-43fa-abaf-754697601192","value":0.6}],"h":[{"id":"ba7fb15d-9b1a-4714-b2fe-e1d9033aec52","value":0.7}],"i":[{"id":"8825cd52-b735-401c-8d19-a7e8e7dfa7ce","value":8}],"j":[{"id":"4fa78253-661c-40ec-b01b-04b832e39ba6","value":9}],"k":[{"id":"fe30d938-7d68-4ade-b9da-1e9a870b3abf","value":10}],"l":[{"id":"c2097bd4-0bf6-4b33-9a10-255fa1c44afd","value":11}],"m":[{"id":"b1f3b5b2-bc93-43c1-a557-3fbb20053210","value":12}],"n":[{"id":"e0529791-402e-4678-b6b3-450a42da483b","value":true}],"aa":[{"id":"8a31d342-49b3-41c4-af81-cb0302700ca4","value":"a"},{"id":"ab2a4ac1-6b44-4248-839e-915f5564ee8c","value":"a"}],"bb":[{"id":"55409f20-c805-4a88-8779-e6f2752ef9c1","value":1},{"id":"07748ddd-e780-4535-8dcf-f6cb7048e721","value":1}],"cc":[{"id":"2032ce2b-e935-469a-8dfd-cfc14965aaaf","value":2},{"id":"f86a37d9-e29c-4b30-b469-8373c63dfc3e","value":2}],"dd":[{"id":"b223b0ba-0c13-4518-9f03-29939ca113e1","value":3},{"id":"1588fbe1-5cdc-403f-bf12-9e72f5098d6a","value":3}],"ee":[{"id":"2ab1d4d7-0707-4d87-92e0-6c3aa6dd3f85","value":4},{"id":"f786c09a-c0d9-4c75-9041-997fb63e115b","value":4}],"ff":[{"id":"cadfe54e-ba9a-4a32-8035-50ffe28c0ba7","value":5},{"id":"cbfbf694-a9da-45cc-beb2-d6bb36067357","value":5}],"gg":[{"id":"5b545ef7-f6d3-4fc5-b1ef-2681b2003725","value":0.6},{"id":"524ab3c0-f634-4809-b3f5-df3971ec337c","value":0.6}],"hh":[{"id":"7e431db6-7a30-48c1-b3e8-9f46da66f02f","value":0.7},{"id":"1c1d5486-f60c-4192-b539-1c9b55d6a766","value":0.7}],"ii":[{"id":"dc382d5b-8b36-41b2-9f5d-d2bf116ba7fc","value":8},{"id":"8d1d8c0e-6cfa-40b9-9df6-cc2ebd353ec7","value":8}],"jj":[{"id":"9d00030e-a74b-413e-a235-34e14e0965f2","value":9},{"id":"61f7edc5-d55c-4682-87d8-d967c989fbd5","value":9}],"kk":[{"id":"396bab91-feec-467b-94b9-5c7ca3560a52","value":10},{"id":"29550cee-d97b-4a09-86b5-7448bf3332bc","value":10}],"ll":[{"id":"4edccd96-7a12-45a0-8e6e-44af77ee363d","value":11},{"id":"a58929a5-04e2-45b0-b2e9-7fbba4a86482","value":11}],"mm":[{"id":"a75fb6b6-7a0f-4b73-b2b8-cc32642115ef","value":12},{"id":"394e4399-4c82-4da1-bbb5-98d1f9ca5bb1","value":12}],"nn":[{"id":"699700f4-4dee-4338-ae17-d41728092f9b","value":true},{"id":"5ffb6551-c84b-4efd-82c1-912dd43c2f83","value":true}],"x":[{"id":"8334df77-a621-4199-a7c2-c24cbcbe9ae4","value":13}],"xx":[{"id":"72a80e12-28d4-4662-ad31-94c9add5d7ce","value":14},{"id":"129673e2-1d3d-49d5-aa75-12169fbbbdd5","value":14}]}}],"meta":{}}}`

var addEResp = `{"requestId":"9fdfe0ec-5329-4d3a-9566-e275b4734ff5","status":{"code":200,"attributes":{"x-ms-status-code":200,"x-ms-request-charge":14.35,"x-ms-total-request-charge":14.35,"x-ms-cosmosdb-graph-request-charge":14.35,"StorageRU":14.35},"message":""},"result":{"data":[{"id":"e623ef5c-01f9-44f1-9684-f33c2e6598ee","label":"relates","type":"edge","inVLabel":"test","outVLabel":"test","inV":"d014ab68-fa70-4a6c-8f11-33fd3eef0112","outV":"e3ff8f7d-0b29-4f4e-854a-affa3544b12a"}],"meta":{}}}`

var addEWithPropsResp = `{"requestId":"9fdfe0ec-5329-4d3a-9566-e275b4734ff5","status":{"code":200,"attributes":{"x-ms-status-code":200,"x-ms-request-charge":14.35,"x-ms-total-request-charge":14.35,"x-ms-cosmosdb-graph-request-charge":14.35,"StorageRU":14.35},"message":""},"result":{"data":[{"id":"e623ef5c-01f9-44f1-9684-f33c2e6598ee","label":"relates","type":"edge","inVLabel":"test","outVLabel":"test","inV":"d014ab68-fa70-4a6c-8f11-33fd3eef0112","outV":"e3ff8f7d-0b29-4f4e-854a-affa3544b12a","properties":{"foo":"bar","biz":3}}],"meta":{}}}`

var getResp = `{"requestId":"53d49b1d-74e1-473e-8130-762901557daf","status":{"code":200,"attributes":{"x-ms-status-code":200,"x-ms-request-charge":3.64,"x-ms-total-request-charge":3.64,"x-ms-cosmosdb-graph-request-charge":3.64,"StorageRU":3.64},"message":""},"result":{"data":[{"id":"64795211-c4a1-4eac-9e0a-b674ced77461","label":"test","type":"vertex","properties":{"a":[{"id":"15d0a33b-d369-4b61-b162-320ece53cfa1","value":"aa"}],"b":[{"id":"91df576d-3501-4303-9d89-1c8409ce6ff4","value":10}],"c":[{"id":"8e58d327-e06b-44e4-a5d9-75558cdca2dc","value":20}],"d":[{"id":"565d7400-e75b-4813-aa39-6c09cae781a8","value":30}],"e":[{"id":"bf35b756-0640-48c6-9601-aab77c6aa603","value":40}],"f":[{"id":"ed8cf6a7-d585-4575-a08c-cf4aa27f1491","value":50}],"g":[{"id":"2c6860ac-7151-48f9-b866-5b40a3488d1e","value":0.06}],"h":[{"id":"3e16edbc-5a77-4b97-bdb4-4695996d8915","value":0.07}],"i":[{"id":"2295b3de-fc5b-42c9-8b04-44b12fbe1346","value":80}],"j":[{"id":"e7235d64-4212-448d-869f-612cf2403b96","value":90}],"k":[{"id":"a435d9ad-9f8f-43d6-b108-a2fe1d5a95b9","value":100}],"l":[{"id":"e7c6ddee-729e-44a4-b977-7d3eafe47497","value":110}],"m":[{"id":"a4b507fe-bb16-4b4a-aa1d-cf922af67cd2","value":120}],"n":[{"id":"954cc7f9-d655-4123-a66d-e3e665cf7d49","value":true}],"aa":[{"id":"225ed5a7-b000-4a59-b6c3-332682a5216a","value":"aa"},{"id":"9cbee039-c5b4-4e75-a1b0-346a47e5dc36","value":"aa"}],"bb":[{"id":"b96f76ed-028a-4e2f-942e-2adf37f5bcb0","value":10},{"id":"7f010e2c-b764-4601-b190-4b34372203e7","value":10}],"cc":[{"id":"8f16c7cd-4125-4d29-b714-5d8f561bb8e4","value":20},{"id":"84f44c14-f038-47b3-a0b7-9f14dd11ddde","value":20}],"dd":[{"id":"51336fa2-ebf9-4d7a-9ec5-0128a6341ea6","value":30},{"id":"337fe24f-8ea3-40ea-b726-e7169883618b","value":30}],"ee":[{"id":"8d21dad3-5ed7-4925-bfae-12b345592a36","value":40},{"id":"6277ad55-a3ed-41bc-8ad2-1cfe6e60938b","value":40}],"ff":[{"id":"4c2afab4-df69-490b-8cf0-c8311808c0fc","value":50},{"id":"38e68ad4-b9f8-4256-9638-1e52cdbb989a","value":50}],"gg":[{"id":"a0a129f4-bfa4-4d1a-bc77-df5b63049197","value":0.06},{"id":"3c8b51c4-8b40-4d5f-a79d-6a9d74929837","value":0.06}],"hh":[{"id":"b2266aff-1f18-4391-98f8-5ad6a542a2e1","value":0.07},{"id":"709a88dd-8fc4-4c8d-bd31-61962feff9b2","value":0.07}],"ii":[{"id":"fe3a148a-4a80-4ca2-851b-5dc473f549e6","value":80},{"id":"021a9a1a-49c1-4ae3-aa40-3f5c33a12e9f","value":80}],"jj":[{"id":"84e17824-94be-47e9-b7bb-7e46ed5c065f","value":90},{"id":"20dffbe0-4c63-4b15-9b5d-43111bd10525","value":90}],"kk":[{"id":"740dca9e-c5d7-40d3-8d68-feb48090a638","value":100},{"id":"a0587e7d-6bf5-4cba-b911-0507d0469068","value":100}],"ll":[{"id":"ff4d7387-b3f7-41ec-9cee-912bb9220545","value":110},{"id":"519d212c-4774-49a1-bc4e-3715af929c38","value":110}],"mm":[{"id":"0487cc86-ee49-4649-8131-d43610235c40","value":120},{"id":"b6e9c4c6-8ac4-4124-92cd-e53acf0cfd12","value":120}],"nn":[{"id":"7af1d164-3966-4dde-93fa-511a936601f5","value":true},{"id":"90be7e5c-8bf6-4bfd-bd01-38be1697d9f8","value":true}],"x":[{"id":"56b71ade-d0aa-416f-8da3-517391fd7ee4","value":130}],"xx":[{"id":"9cf5c2a7-45eb-4e58-bf5a-9f15186c0819","value":140},{"id":"122191a7-5437-45ae-9ec6-1a73fee5c996","value":140}],"z":[{"id":"f97cc2c0-dacf-4167-89cd-9df0bff9756b","value":"{\"Id\":\"96f7cacd-01fd-469e-a14c-5178903a39b6\",\"A\":\"aa\",\"B\":10}"}],"zz":[{"id":"33ac5504-4369-4b1a-af93-5544e1790670","value":"[{\"Id\":\"4f9a2f8b-3b3a-4028-a1e0-fa55e0dd1543\",\"A\":\"aa\",\"B\":10},{\"Id\":\"593db951-3456-450a-a779-44f1a4bd330d\",\"A\":\"aa\",\"B\":10}]"}]}}],"meta":{}}}`

func nows(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `nows`)
}

func pong(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()
	err = c.WriteMessage(websocket.PongMessage, []byte{})
	if err != nil {
		return
	}
	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			break
		}
		err = c.WriteMessage(websocket.PongMessage, []byte{})
		if err != nil {
			break
		}
	}
}

func mock(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			break
		}
		mimeType := []byte("!application/vnd.gremlin-v2.0+json")
		msg := bytes.SplitAfter(message, mimeType)
		if len(msg) == 2 {
			var req GremlinRequest
			err := json.Unmarshal(msg[1], &req)
			if err != nil {
				break
			}
			gremlin := req.Args["gremlin"]
			fmt.Printf("------>   Mock Server Request: %s\n", gremlin)
			switch gremlin {
			case string(gremV): // query the whole graph and return a empty graph
				var resp GremlinResponse
				err := json.Unmarshal([]byte(vResp), &resp)
				if err != nil {
					break
				}
				resp.RequestId = req.RequestId
				respMessage, err := json.Marshal(resp)
				if err != nil {
					break
				}
				err = c.WriteMessage(mt, respMessage)
				if err != nil {
					break
				}
			case string(gremE): // query add edge
				var resp GremlinResponse
				err := json.Unmarshal([]byte(addEResp), &resp)
				if err != nil {
					break
				}
				resp.RequestId = req.RequestId
				respMessage, err := json.Marshal(resp)
				if err != nil {
					break
				}
				err = c.WriteMessage(mt, respMessage)
				if err != nil {
					break
				}
			case string(gremEWithProps): // query add edge with props
				var resp GremlinResponse
				err := json.Unmarshal([]byte(addEWithPropsResp), &resp)
				if err != nil {
					break
				}
				resp.RequestId = req.RequestId
				respMessage, err := json.Marshal(resp)
				if err != nil {
					break
				}
				err = c.WriteMessage(mt, respMessage)
				if err != nil {
					break
				}
			case string(gremEWithProps2): // query add edge with props with alternate map order
				var resp GremlinResponse
				err := json.Unmarshal([]byte(addEWithPropsResp), &resp)
				if err != nil {
					break
				}
				resp.RequestId = req.RequestId
				respMessage, err := json.Marshal(resp)
				if err != nil {
					break
				}
				err = c.WriteMessage(mt, respMessage)
				if err != nil {
					break
				}
			case string(gremEWithPropsSlice): // query add edge with props
				var resp GremlinResponse
				err := json.Unmarshal([]byte(addEWithPropsResp), &resp)
				if err != nil {
					break
				}
				resp.RequestId = req.RequestId
				respMessage, err := json.Marshal(resp)
				if err != nil {
					break
				}
				err = c.WriteMessage(mt, respMessage)
				if err != nil {
					break
				}
			case string(gremDropE): // query drop edge
				var resp GremlinResponse
				err := json.Unmarshal([]byte(vResp), &resp)
				if err != nil {
					break
				}
				resp.RequestId = req.RequestId
				respMessage, err := json.Marshal(resp)
				if err != nil {
					break
				}
				err = c.WriteMessage(mt, respMessage)
				if err != nil {
					break
				}
			case string(gremV1): // query add vertex
				var resp GremlinResponse
				err := json.Unmarshal([]byte(addV1Resp), &resp)
				if err != nil {
					break
				}
				resp.RequestId = req.RequestId
				respMessage, err := json.Marshal(resp)
				if err != nil {
					break
				}
				err = c.WriteMessage(mt, respMessage)
				if err != nil {
					break
				}
			case string(gremUpdateV1): // query update vertex
				var resp GremlinResponse
				err := json.Unmarshal([]byte(addV1Resp), &resp)
				if err != nil {
					break
				}
				resp.RequestId = req.RequestId
				respMessage, err := json.Marshal(resp)
				if err != nil {
					break
				}
				err = c.WriteMessage(mt, respMessage)
				if err != nil {
					break
				}
			case string(gremDropV1): // query drop vertex
				var resp GremlinResponse
				err := json.Unmarshal([]byte(vResp), &resp)
				if err != nil {
					break
				}
				resp.RequestId = req.RequestId
				respMessage, err := json.Marshal(resp)
				if err != nil {
					break
				}
				err = c.WriteMessage(mt, respMessage)
				if err != nil {
					break
				}
			case string(gremGet): // query get vertex
				var resp GremlinResponse
				err := json.Unmarshal([]byte(getResp), &resp)
				if err != nil {
					break
				}
				resp.RequestId = req.RequestId
				respMessage, err := json.Marshal(resp)
				if err != nil {
					break
				}
				err = c.WriteMessage(mt, respMessage)
				if err != nil {
					break
				}
			default:
				err = c.WriteMessage(mt, []byte("FOOBAR"))
				if err != nil {
					return
				}
			}
		} else if mt == websocket.PingMessage {
			c.WriteMessage(websocket.PongMessage, []byte{})
			break
		} else {
			c.WriteMessage(mt, []byte("TEST"))
			break
		}
	}
}

func TestWsConnection(t *testing.T) {
	assert := assert.New(t)

	// Create test server with the mock handler.
	s := httptest.NewServer(http.HandlerFunc(mock))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// test connecting to the mock server
	g, _ := NewClient(NewClientConfig(u))
	ws := g.conn.(*Ws)

	// test to see if connectend
	connected := ws.isConnected()
	assert.True(connected)

	// test to see if dispoased
	disposed := ws.isDisposed()
	assert.False(disposed)

	// test ping
	ws.pingInterval = time.Duration(10) * time.Millisecond
	errs := make(chan error)
	go ws.ping(errs)
	go func() {
		time.Sleep(time.Duration(15) * time.Millisecond)
		close(ws.quit)
		close(errs)
	}()
	err := <-errs
	assert.Nil(err)
	time.Sleep(time.Duration(15) * time.Millisecond)

	// test close
	g2, _ := NewClient(NewClientConfig(u))
	ws2 := g2.conn.(*Ws)
	err = ws2.close()
	assert.Nil(err)
}

func TestWsConnectionError(t *testing.T) {
	assert := assert.New(t)

	// Create test server with the hello handler.
	s := httptest.NewServer(http.HandlerFunc(nows))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// test connecting to the hello server
	g, errs := NewClient(NewClientConfig(u))
	err := <-errs
	assert.Equal(ErrorWSConnectionNil, err)
	assert.NotNil(g)
}

func TestWsConnectiongPongHandler(t *testing.T) {
	assert := assert.New(t)

	// Create test server with the ping handler.
	s := httptest.NewServer(http.HandlerFunc(pong))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.
	u := "ws" + strings.TrimPrefix(s.URL, "http")

	// test connecting to the ping server
	ws := Ws{uri: u}
	err := ws.connect()
	assert.Equal(nil, err)

	err = ws.pongHandler("")
	assert.Nil(err)
}
