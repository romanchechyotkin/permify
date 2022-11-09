package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/Permify/permify/pkg/database"
	base "github.com/Permify/permify/pkg/pb/base/v1"
	"github.com/Permify/permify/pkg/token"
)

// RelationshipWriter is an autogenerated mock type for the RelationshipWriter type
type RelationshipWriter struct {
	mock.Mock
}

// WriteRelationships -
func (_m *SchemaReader) WriteRelationships(ctx context.Context, collection database.ITupleCollection) (token.EncodedSnapToken, error) {
	ret := _m.Called(collection)

	var r0 token.EncodedSnapToken
	if rf, ok := ret.Get(0).(func(context.Context, database.ITupleCollection) token.EncodedSnapToken); ok {
		r0 = rf(ctx, collection)
	} else {
		r0 = ret.Get(0).(token.EncodedSnapToken)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, database.ITupleCollection) error); ok {
		r1 = rf(ctx, collection)
	} else {
		if e, ok := ret.Get(1).(error); ok {
			r1 = e
		} else {
			r1 = nil
		}
	}

	return r0, r1
}

// DeleteRelationships -
func (_m *SchemaReader) DeleteRelationships(ctx context.Context, filter *base.TupleFilter) (token.EncodedSnapToken, error) {
	ret := _m.Called(filter)

	var r0 token.EncodedSnapToken
	if rf, ok := ret.Get(0).(func(context.Context, *base.TupleFilter) token.EncodedSnapToken); ok {
		r0 = rf(ctx, filter)
	} else {
		r0 = ret.Get(0).(token.EncodedSnapToken)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *base.TupleFilter) error); ok {
		r1 = rf(ctx, filter)
	} else {
		if e, ok := ret.Get(1).(error); ok {
			r1 = e
		} else {
			r1 = nil
		}
	}

	return r0, r1
}