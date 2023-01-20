package protobuf

type (
	//easyjson:json
	JsonLongString struct {
		Payload string `json:"payload,omitempty"`
	}

	//easyjson:json
	JsonObject struct {
		Id       int32   `json:"id,omitempty"`
		Price    float32 `json:"price,omitempty"`
		Datetime *int64  `json:"datetime,omitempty"`
		Data     string  `json:"data,omitempty"`
	}

	//easyjson:json
	JsonLargeResponse struct {
		Data []*JsonObject `json:"data,omitempty"`
	}
)
