package search_v3

import (
	"embed"
	"encoding/json"
	"fmt"
	v3 "github.com/KosyanMedia/delta/search/cmd/results-api/api/v3"
	"github.com/golang/protobuf/proto"
	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/require"
	"go-playground/protobuf/utils"
	"testing"
)

var (
	//go:embed results.json
	resultsJson embed.FS

	jsonIter = jsoniter.ConfigFastest
)

func TestBinarySize(t *testing.T) {
	dump := readDump()

	var data v3.SearchResults
	require.NoError(t, jsonIter.Unmarshal(dump, &data))

	marshalledJson := utils.Must2(jsonIter.Marshal(data))
	fmt.Printf("Json body length: %d\n", len(marshalledJson))
	utils.GzipAndPrint(marshalledJson)

	protoData := resultsToProto(data)
	bodyBytes := utils.Must2(proto.Marshal(protoData))
	fmt.Printf("Proto body length: %d\n", len(bodyBytes))
	utils.GzipAndPrint(bodyBytes)

	// Checking that original struct is equal to proto struct, converted by `resultsToProto`
	//requireDeepEqual(t, reflect.ValueOf(data), reflect.ValueOf(protoData.Chunks))
}

func BenchmarkObject_MarshalJSON(b *testing.B) {
	data := readDumpStruct()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		json.Marshal(data)
	}
}

func BenchmarkObject_UnmarshalJSON(b *testing.B) {
	bytes := readDump()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var target v3.SearchResults
		json.Unmarshal(bytes, &target)
	}
}

//func BenchmarkObject_MarshalEasyJSON(b *testing.B) {
//	data := readDumpStruct()
//	b.ReportAllocs()
//	b.ResetTimer()
//
//	for i := 0; i < b.N; i++ {
//		easyjson.Marshal(data)
//	}
//}
//
//func BenchmarkObject_UnmarshalEasyJSON(b *testing.B) {
//	bytes := readDump()
//	b.ReportAllocs()
//	b.ResetTimer()
//
//	for i := 0; i < b.N; i++ {
//		var target v3.SearchResults
//		easyjson.Unmarshal(bytes, &target)
//	}
//}

func BenchmarkObject_MarshalIterJSON(b *testing.B) {
	data := readDumpStruct()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		jsonIter.Marshal(data)
	}
}

func BenchmarkObject_UnmarshalIterJSON(b *testing.B) {
	bytes := readDump()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var target v3.SearchResults
		jsonIter.Unmarshal(bytes, &target)
	}
}

func BenchmarkObject_MarshalProto(b *testing.B) {
	data := resultsToProto(readDumpStruct())
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		proto.Marshal(data)
	}
}

func BenchmarkObject_UnmarshalProto(b *testing.B) {
	data := resultsToProto(readDumpStruct())
	bytes := utils.Must2(proto.Marshal(data))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var target SearchResults
		proto.Unmarshal(bytes, &target)
	}
}

func BenchmarkObject_MarshalVTProto(b *testing.B) {
	data := resultsToProto(readDumpStruct())
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data.MarshalVT()
	}
}

func BenchmarkObject_UnmarshalVTProto(b *testing.B) {
	data := resultsToProto(readDumpStruct())
	bytes := utils.Must2(proto.Marshal(data))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var target SearchResults
		target.UnmarshalVT(bytes)
	}
}

func readDump() []byte {
	return utils.Must2(resultsJson.ReadFile("results.json"))
}

func readDumpStruct() v3.SearchResults {
	dump := readDump()

	var data v3.SearchResults
	utils.Must(jsonIter.Unmarshal(dump, &data))
	return data
}
