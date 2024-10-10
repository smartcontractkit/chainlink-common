package ocr3cap

var encoderTypes []Encoder

func init() {
	for _, v := range enumValues_Encoder {
		encoderTypes = append(encoderTypes, Encoder(v.(string)))
	}
}

func Encoders() []Encoder {
	cpy := make([]Encoder, len(encoderTypes))
	copy(cpy, encoderTypes)
	return cpy
}
