package msgspb

import (
	"google.golang.org/protobuf/encoding/protojson"
)

func (r *UserConfig) MarshalJSON() ([]byte, error) {
	marshaler := protojson.MarshalOptions{
		UseProtoNames: true,
	}
	return marshaler.Marshal(r)
}

func (r *UserConfig) UnmarshalJSON(data []byte) error {

	return protojson.Unmarshal(data, r)
}
