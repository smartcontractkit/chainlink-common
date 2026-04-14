package chipingress

type RawCodec struct{}

func (RawCodec) Name() string {
	return "raw"
}

func (RawCodec) Marshal(v interface{}) ([]byte, error) {
	return v.([]byte), nil
}

func (RawCodec) Unmarshal(data []byte, v interface{}) error {
	ptr := v.(*[]byte)
	*ptr = data
	return nil
}
