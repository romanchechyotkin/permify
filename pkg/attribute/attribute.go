package attribute

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	base "github.com/Permify/permify/pkg/pb/base/v1"
)

const (
	ENTITY    = "%s:%s"
	ATTRIBUTE = "#%s"
	VALUE     = "%s:%s"
)

// Attribute function takes a string representation of an attribute and converts it back into the Attribute object.
func Attribute(attribute string) (*base.Attribute, error) {
	// Splitting the attribute string by "@" delimiter
	s := strings.Split(strings.TrimSpace(attribute), "|")
	if len(s) != 2 {
		// The attribute string should have exactly two parts
		return nil, ErrInvalidAttribute
	}

	// Splitting the entity part of the string by "#" delimiter
	e := strings.Split(s[0], "$")
	if len(e) != 2 {
		// The entity string should have exactly two parts
		return nil, ErrInvalidEntity
	}

	// Splitting the entity type and id by ":" delimiter
	et := strings.Split(e[0], ":")
	if len(et) != 2 {
		// The entity type and id should have exactly two parts
		return nil, ErrInvalidAttribute
	}

	// Splitting the attribute value part of the string by ":" delimiter
	v := strings.Split(s[1], ":")
	if len(v) != 2 {
		// The attribute value string should have exactly two parts
		return nil, ErrInvalidAttribute
	}

	// Declare a proto message to hold the attribute value
	var wrapped proto.Message

	// Parse the attribute value based on its type
	switch v[0] {
	case "boolean":
		boolVal, err := strconv.ParseBool(v[1])
		if err != nil {
			return nil, fmt.Errorf("failed to parse boolean: %v", err)
		}
		wrapped = &base.Boolean{Value: boolVal}
	case "boolean[]":
		var ba []bool
		val := strings.Split(v[1], ",")
		for _, value := range val {
			boolVal, err := strconv.ParseBool(value)
			if err != nil {
				return nil, fmt.Errorf("failed to parse boolean: %v", err)
			}
			ba = append(ba, boolVal)
		}
		wrapped = &base.BooleanArray{Values: ba}
	case "string":
		wrapped = &base.String{Value: v[1]}
	case "string[]":
		var sa []string
		val := strings.Split(v[1], ",")
		for _, value := range val {
			sa = append(sa, value)
		}
		wrapped = &base.StringArray{Values: sa}
	case "double":
		doubleVal, err := strconv.ParseFloat(v[1], 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse float: %v", err)
		}
		wrapped = &base.Double{Value: doubleVal}
	case "double[]":
		var da []float64
		val := strings.Split(v[1], ",")
		for _, value := range val {
			doubleVal, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse float: %v", err)
			}
			da = append(da, doubleVal)
		}
		wrapped = &base.DoubleArray{Values: da}
	case "integer":
		intVal, err := strconv.ParseInt(v[1], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("failed to parse integer: %v", err)
		}
		wrapped = &base.Integer{Value: int32(intVal)}
	case "integer[]":

		var ia []int32
		val := strings.Split(v[1], ",")
		for _, value := range val {
			intVal, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("failed to parse integer: %v", err)
			}
			ia = append(ia, int32(intVal))
		}
		wrapped = &base.IntegerArray{Values: ia}
	default:
		return nil, ErrInvalidValue
	}

	// Convert the wrapped attribute value into Any proto message
	value, err := anypb.New(wrapped)
	if err != nil {
		return nil, err
	}

	// Return the attribute object
	return &base.Attribute{
		Entity: &base.Entity{
			Type: et[0],
			Id:   et[1],
		},
		Attribute: e[1],
		Value:     value,
	}, nil
}

// ToString function takes an Attribute object and converts it into a string.
func ToString(attribute *base.Attribute) string {
	// Get the entity from the attribute
	entity := attribute.GetEntity()

	// Convert the entity to string
	strEntity := EntityToString(entity)

	// Create the string representation of the attribute
	result := fmt.Sprintf("%s$%s|%s:%s", strEntity, attribute.GetAttribute(), TypeUrlToString(attribute.GetValue().GetTypeUrl()), AnyToString(attribute.GetValue()))

	return result
}

// EntityToString function takes an Entity object and converts it into a string.
func EntityToString(entity *base.Entity) string {
	return fmt.Sprintf(ENTITY, entity.GetType(), entity.GetId())
}

func TypeUrlToString(url string) string {
	switch url {
	case "type.googleapis.com/base.v1.String":
		return "string"
	case "type.googleapis.com/base.v1.Boolean":
		return "boolean"
	case "type.googleapis.com/base.v1.Integer":
		return "integer"
	case "type.googleapis.com/base.v1.Double":
		return "double"
	case "type.googleapis.com/base.v1.StringArray":
		return "string[]"
	case "type.googleapis.com/base.v1.BooleanArray":
		return "boolean[]"
	case "type.googleapis.com/base.v1.IntegerArray":
		return "integer[]"
	case "type.googleapis.com/base.v1.DoubleArray":
		return "double[]"
	default:
		return ""
	}
}

// AnyToString function takes an Any proto message and converts it into a string.
func AnyToString(any *anypb.Any) string {
	var str string

	// Convert the Any proto message into string based on its TypeUrl
	switch any.TypeUrl {
	case "type.googleapis.com/base.v1.Boolean":
		boolVal := &base.Boolean{}
		if err := any.UnmarshalTo(boolVal); err != nil {
			return "undefined"
		}
		str = strconv.FormatBool(boolVal.Value)
	case "type.googleapis.com/base.v1.BooleanArray":
		boolVal := &base.BooleanArray{}
		if err := any.UnmarshalTo(boolVal); err != nil {
			return "undefined"
		}
		var strs []string
		for _, b := range boolVal.GetValues() {
			strs = append(strs, strconv.FormatBool(b))
		}
		str = strings.Join(strs, ",")
	case "type.googleapis.com/base.v1.String":
		stringVal := &base.String{}
		if err := any.UnmarshalTo(stringVal); err != nil {
			return "undefined"
		}
		str = stringVal.Value
	case "type.googleapis.com/base.v1.StringArray":
		stringVal := &base.StringArray{}
		if err := any.UnmarshalTo(stringVal); err != nil {
			return "undefined"
		}
		var strs []string
		for _, v := range stringVal.GetValues() {
			strs = append(strs, v)
		}
		str = strings.Join(strs, ",")
	case "type.googleapis.com/base.v1.Double":
		doubleVal := &base.Double{}
		if err := any.UnmarshalTo(doubleVal); err != nil {
			return "undefined"
		}
		str = strconv.FormatFloat(doubleVal.Value, 'f', -1, 64)
	case "type.googleapis.com/base.v1.DoubleArray":
		doubleVal := &base.DoubleArray{}
		if err := any.UnmarshalTo(doubleVal); err != nil {
			return "undefined"
		}
		var strs []string
		for _, v := range doubleVal.GetValues() {
			strs = append(strs, strconv.FormatFloat(v, 'f', -1, 64))
		}
		str = strings.Join(strs, ",")
	case "type.googleapis.com/base.v1.Integer":
		intVal := &base.Integer{}
		if err := any.UnmarshalTo(intVal); err != nil {
			return "undefined"
		}
		str = strconv.Itoa(int(intVal.Value))
	case "type.googleapis.com/base.v1.IntegerArray":
		intVal := &base.IntegerArray{}
		if err := any.UnmarshalTo(intVal); err != nil {
			return "undefined"
		}
		var strs []string
		for _, v := range intVal.GetValues() {
			strs = append(strs, strconv.Itoa(int(v)))
		}
		str = strings.Join(strs, ",")
	default:
		return "undefined"
	}

	return str
}

// TypeToString function takes an AttributeType enum and converts it into a string.
func TypeToString(attributeType base.AttributeType) string {
	switch attributeType {
	case base.AttributeType_ATTRIBUTE_TYPE_INTEGER:
		return "integer"
	case base.AttributeType_ATTRIBUTE_TYPE_INTEGER_ARRAY:
		return "integer[]"
	case base.AttributeType_ATTRIBUTE_TYPE_DOUBLE:
		return "double"
	case base.AttributeType_ATTRIBUTE_TYPE_DOUBLE_ARRAY:
		return "double[]"
	case base.AttributeType_ATTRIBUTE_TYPE_STRING:
		return "string"
	case base.AttributeType_ATTRIBUTE_TYPE_STRING_ARRAY:
		return "string[]"
	case base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN:
		return "boolean"
	case base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN_ARRAY:
		return "boolean[]"
	default:
		return "undefined"
	}
}

// ValidateValue checks the validity of the 'any' parameter which is a protobuf 'Any' type,
// based on the attribute type provided.
//
// 'any' is a protobuf 'Any' type which should contain a value of a specific type.
// 'attributeType' is an enum indicating the expected type of the value within 'any'.
// The function returns an error if the value within 'any' is not of the expected type, or if unmarshalling fails.
//
// The function returns nil if the value is valid (i.e., it is of the expected type and can be successfully unmarshalled).
func ValidateValue(any *anypb.Any, attributeType base.AttributeType) error {
	// Declare a variable 'target' of type proto.Message to hold the unmarshalled value.
	var target proto.Message

	// Depending on the expected attribute type, assign 'target' a new instance of the corresponding specific type.
	switch attributeType {
	case base.AttributeType_ATTRIBUTE_TYPE_INTEGER:
		target = &base.Integer{}
	case base.AttributeType_ATTRIBUTE_TYPE_INTEGER_ARRAY:
		target = &base.IntegerArray{}
	case base.AttributeType_ATTRIBUTE_TYPE_DOUBLE:
		target = &base.Double{}
	case base.AttributeType_ATTRIBUTE_TYPE_DOUBLE_ARRAY:
		target = &base.DoubleArray{}
	case base.AttributeType_ATTRIBUTE_TYPE_STRING:
		target = &base.String{}
	case base.AttributeType_ATTRIBUTE_TYPE_STRING_ARRAY:
		target = &base.StringArray{}
	case base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN:
		target = &base.Boolean{}
	case base.AttributeType_ATTRIBUTE_TYPE_BOOLEAN_ARRAY:
		target = &base.BooleanArray{}
	default:
		// If attributeType doesn't match any of the known types, return an error indicating invalid argument.
		return errors.New(base.ErrorCode_ERROR_CODE_INVALID_ARGUMENT.String())
	}

	// Attempt to unmarshal the value in 'any' into 'target'.
	// If this fails, return the error from UnmarshalTo.
	if err := any.UnmarshalTo(target); err != nil {
		return err
	}

	// If the value was successfully unmarshalled and is of the expected type, return nil to indicate success.
	return nil
}