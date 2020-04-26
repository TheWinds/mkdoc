package mkdoc

// ObjectType describe a object
//
// Name is one of below:
// object,
// string,
// bool,
// byte,
// interface{},
// int,int8,int16,int32,int64,
// uint,uint8,uint16,uint32,uint64,
// float,float32,float64
//
// Ref describe which object to reference
//
// IsRepeated will be true if that is a array/slice type
type ObjectType struct {
	Name       string
	Ref        string
	IsRepeated bool
}
