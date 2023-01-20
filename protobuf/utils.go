package protobuf

import (
	"go-playground/protobuf/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

func genObjects(n int) []*JsonObject {
	result := make([]*JsonObject, 0, n)
	for i := 0; i < n; i++ {
		now := time.Now().Unix()
		result = append(result, &JsonObject{
			Id:       int32(i),
			Price:    float32(i) * 0.32,
			Datetime: &now,
			Data:     utils.RandomString(n),
		})
	}
	return result
}

func toProto(objects []*JsonObject) []*Object {
	result := make([]*Object, 0, len(objects))
	for _, object := range objects {
		result = append(result, objToProto(object))
	}
	return result
}

func objToProto(object *JsonObject) *Object {
	if object == nil {
		return nil
	}
	return &Object{
		Id:       object.Id,
		Price:    object.Price,
		Datetime: timestamppb.New(time.Unix(*object.Datetime, 0)),
		Data:     object.Data,
	}
}
