package protobuf

import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	jsoniter "github.com/json-iterator/go"
	"github.com/mailru/easyjson"
	"github.com/stretchr/testify/require"
	"go-playground/protobuf/utils"
	"testing"
	"time"
)

const (
	longStringLen = 5000
	objectsCount  = 100
)

var jsonIter = jsoniter.ConfigFastest

// 5014
func Test_Json_LongString(t *testing.T) {
	body := JsonLongString{
		Payload: utils.RandomString(longStringLen),
	}
	bodyBytes := utils.Must2(json.Marshal(body))

	fmt.Printf("Json body length: %d\n", len(bodyBytes))
	utils.GzipAndPrint(bodyBytes)
	var bodyStruct JsonLongString
	utils.Must(json.Unmarshal(bodyBytes, &bodyStruct))

	require.Equal(t, longStringLen, len(bodyStruct.Payload))
}

// 15660
func Test_Json_LargeResponse(t *testing.T) {
	body := JsonLargeResponse{
		Data: genObjects(objectsCount),
	}
	bodyBytes := utils.Must2(json.Marshal(body))

	fmt.Printf("Json body length: %d\n", len(bodyBytes))
	utils.GzipAndPrint(bodyBytes)
	var bodyStruct JsonLargeResponse
	utils.Must(json.Unmarshal(bodyBytes, &bodyStruct))

	require.Equal(t, objectsCount, len(bodyStruct.Data))
}

// 5003
func Test_Protobuf_LongString(t *testing.T) {
	body := LongString{
		Payload: utils.RandomString(longStringLen),
	}
	bodyBytes := utils.Must2(proto.Marshal(&body))

	fmt.Printf("Protobuf body length: %d\n", len(bodyBytes))
	utils.GzipAndPrint(bodyBytes)
	var bodyStruct LongString
	utils.Must(proto.Unmarshal(bodyBytes, &bodyStruct))

	require.Equal(t, longStringLen, len(bodyStruct.Payload))
}

// 11893
func Test_Protobuf_LargeResponse(t *testing.T) {
	body := LargeResponse{
		Data: toProto(genObjects(objectsCount)),
	}
	bodyBytes := utils.Must2(proto.Marshal(&body))

	fmt.Printf("Protobuf body length: %d\n", len(bodyBytes))
	utils.GzipAndPrint(bodyBytes)
	var bodyStruct LargeResponse
	utils.Must(proto.Unmarshal(bodyBytes, &bodyStruct))

	require.Equal(t, objectsCount, len(bodyStruct.Data))
}

func Test_Protobuf_Optional(t *testing.T) {
	obj := &SimpleObject{
		Id:       111,
		Kek:      nil,
		Cheburek: nil,
	}

	bytes, err := proto.Marshal(obj)
	require.NoError(t, err)

	var fromBytes SimpleObject
	require.NoError(t, proto.Unmarshal(bytes, &fromBytes))

	require.Equal(t, obj.Id, fromBytes.Id)
	require.Nil(t, fromBytes.Kek)
	require.Nil(t, fromBytes.Cheburek)
}

var (
	jsonObject = &JsonObject{
		Id:       15123,
		Price:    0.412,
		Datetime: utils.Ptr(time.Date(2022, 12, 23, 4, 51, 24, 0, time.UTC).Unix()),
		Data:     utils.RandomString(10),
	}
	protoObject = objToProto(jsonObject)

	jsonLargeObject = &JsonLargeResponse{
		Data: genObjects(150),
	}
	protoLargeObject = &LargeResponse{
		Data: toProto(jsonLargeObject.Data),
	}

	protoSimpleObject = &SimpleObject{
		Id:    853528,
		Price: 416.7651454,
		Foo:   6851943,
		Bar:   97.00000432,
		Lol:   "Hello World!",
		Kek: &NestedObject{
			I:     666,
			Am:    999999,
			Groot: "Yes",
		},
		Cheburek: []*NestedObject{
			{
				I:     333,
				Am:    666666,
				Groot: "No",
			},
		},
	}
)

// JSON: 68 bytes
// Protobuf: 28 bytes
// Large JSON: 31046 bytes
// Large Protobuf: 25665 bytes
func Test_Bench_Obj_Size(t *testing.T) {
	bytes := utils.Must2(json.Marshal(jsonObject))
	fmt.Printf("Json object size: %d\n", len(bytes))

	bytes = utils.Must2(proto.Marshal(protoObject))
	fmt.Printf("Proto object size: %d\n", len(bytes))

	vtBytes := utils.Must2(protoObject.MarshalVT())
	require.Equal(t, bytes, vtBytes)

	largeBytes := utils.Must2(json.Marshal(jsonLargeObject))
	fmt.Printf("Large Json object size: %d\n", len(largeBytes))

	largeBytes = utils.Must2(proto.Marshal(protoLargeObject))
	fmt.Printf("Large proto object size: %d\n", len(largeBytes))
}

func BenchmarkObject_MarshalJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		json.Marshal(jsonObject)
	}
}

func BenchmarkObject_UnmarshalJSON(b *testing.B) {
	bytes := utils.Must2(json.Marshal(jsonObject))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var target JsonObject
		json.Unmarshal(bytes, &target)
	}
}

func BenchmarkObject_MarshalEasyJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		easyjson.Marshal(jsonObject)
	}
}

func BenchmarkObject_UnmarshalEasyJSON(b *testing.B) {
	bytes := utils.Must2(easyjson.Marshal(jsonObject))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var target JsonObject
		easyjson.Unmarshal(bytes, &target)
	}
}

// Iter
func BenchmarkObject_MarshalIterJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		jsonIter.Marshal(jsonObject)
	}
}

func BenchmarkObject_UnmarshalIterJSON(b *testing.B) {
	bytes := utils.Must2(jsonIter.Marshal(jsonObject))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var target JsonObject
		jsonIter.Unmarshal(bytes, &target)
	}
}

func BenchmarkObject_MarshalProto(b *testing.B) {
	for i := 0; i < b.N; i++ {
		proto.Marshal(protoObject)
	}
}

func BenchmarkObject_UnmarshalProto(b *testing.B) {
	bytes := utils.Must2(proto.Marshal(protoObject))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var target Object
		proto.Unmarshal(bytes, &target)
	}
}

func BenchmarkObject_MarshalVTProto(b *testing.B) {
	for i := 0; i < b.N; i++ {
		protoObject.MarshalVT()
	}
}

func BenchmarkObject_UnmarshalVTProto(b *testing.B) {
	bytes := utils.Must2(protoObject.MarshalVT())
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var target Object
		target.UnmarshalVT(bytes)
	}
}

// Large object
func BenchmarkLargeObject_MarshalJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		json.Marshal(jsonLargeObject)
	}
}

func BenchmarkLargeObject_UnmarshalJSON(b *testing.B) {
	bytes := utils.Must2(json.Marshal(jsonLargeObject))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var target JsonLargeResponse
		json.Unmarshal(bytes, &target)
	}
}

func BenchmarkLargeObject_MarshalEasyJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		easyjson.Marshal(jsonLargeObject)
	}
}

func BenchmarkLargeObject_UnmarshalEasyJSON(b *testing.B) {
	bytes := utils.Must2(easyjson.Marshal(jsonLargeObject))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var target JsonLargeResponse
		easyjson.Unmarshal(bytes, &target)
	}
}

func BenchmarkLargeObject_MarshalIterJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		jsonIter.Marshal(jsonLargeObject)
	}
}

func BenchmarkLargeObject_UnmarshalIterJSON(b *testing.B) {
	bytes := utils.Must2(jsonIter.Marshal(jsonLargeObject))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var target JsonLargeResponse
		jsonIter.Unmarshal(bytes, &target)
	}
}

func BenchmarkLargeObject_MarshalProto(b *testing.B) {
	for i := 0; i < b.N; i++ {
		proto.Marshal(protoLargeObject)
	}
}

func BenchmarkLargeObject_UnmarshalProto(b *testing.B) {
	bytes := utils.Must2(proto.Marshal(protoLargeObject))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var target LargeResponse
		proto.Unmarshal(bytes, &target)
	}
}

func BenchmarkLargeObject_MarshalVTProto(b *testing.B) {
	for i := 0; i < b.N; i++ {
		protoLargeObject.MarshalVT()
	}
}

func BenchmarkLargeObject_UnmarshalVTProto(b *testing.B) {
	bytes := utils.Must2(protoLargeObject.MarshalVT())
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var target LargeResponse
		target.UnmarshalVT(bytes)
	}
}

// Simple object
func BenchmarkSimpleObject_MarshalProto(b *testing.B) {
	for i := 0; i < b.N; i++ {
		proto.Marshal(protoSimpleObject)
	}
}

func BenchmarkSimpleObject_UnmarshalProto(b *testing.B) {
	bytes := utils.Must2(proto.Marshal(protoSimpleObject))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var target SimpleObject
		proto.Unmarshal(bytes, &target)
	}
}

func BenchmarkSimpleObject_MarshalVTProto(b *testing.B) {
	for i := 0; i < b.N; i++ {
		protoSimpleObject.MarshalVT()
	}
}

func BenchmarkSimpleObject_UnmarshalVTProto(b *testing.B) {
	bytes := utils.Must2(protoSimpleObject.MarshalVT())
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var target SimpleObject
		target.UnmarshalVT(bytes)
	}
}
