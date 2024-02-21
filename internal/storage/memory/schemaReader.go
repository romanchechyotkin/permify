package memory

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/go-memdb"
	"github.com/rs/xid"

	"github.com/Permify/permify/internal/schema"
	"github.com/Permify/permify/internal/storage"
	"github.com/Permify/permify/internal/storage/memory/constants"
	"github.com/Permify/permify/internal/storage/memory/utils"
	"github.com/Permify/permify/pkg/database"
	db "github.com/Permify/permify/pkg/database/memory"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// SchemaReader - Structure for Schema Reader
type SchemaReader struct {
	database *db.Memory
}

// NewSchemaReader - Creates a new SchemaReader
func NewSchemaReader(database *db.Memory) *SchemaReader {
	return &SchemaReader{
		database: database,
	}
}

// ReadSchema - Reads a new schema from repository
func (r *SchemaReader) ReadSchema(_ context.Context, tenantID, version string) (sch *base.SchemaDefinition, err error) {
	txn := r.database.DB.Txn(false)
	defer txn.Abort()
	var it memdb.ResultIterator
	it, err = txn.Get(constants.SchemaDefinitionsTable, "version", tenantID, version)
	if err != nil {
		return sch, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	var definitions []string
	for obj := it.Next(); obj != nil; obj = it.Next() {
		definitions = append(definitions, obj.(storage.SchemaDefinition).Serialized())
	}

	sch, err = schema.NewSchemaFromStringDefinitions(false, definitions...)
	if err != nil {
		return nil, err
	}

	return sch, nil
}

// ReadEntityDefinition - Reads a Entity Definition from repository
func (r *SchemaReader) ReadEntityDefinition(_ context.Context, tenantID, entityName, version string) (definition *base.EntityDefinition, v string, err error) {
	txn := r.database.DB.Txn(false)
	defer txn.Abort()
	var raw interface{}
	raw, err = txn.First(constants.SchemaDefinitionsTable, "id", tenantID, entityName, version)
	if err != nil {
		return nil, "", errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	def, ok := raw.(storage.SchemaDefinition)
	if ok {
		var sch *base.SchemaDefinition
		sch, err = schema.NewSchemaFromStringDefinitions(false, def.Serialized())
		if err != nil {
			return nil, "", err
		}
		definition, err = schema.GetEntityByName(sch, entityName)
		if err != nil {
			return nil, "", err
		}
		return definition, def.Version, err
	}

	return nil, "", errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_NOT_FOUND.String())
}

// ReadRuleDefinition - Reads a Rule Definition from repository
func (r *SchemaReader) ReadRuleDefinition(_ context.Context, tenantID, ruleName, version string) (definition *base.RuleDefinition, v string, err error) {
	txn := r.database.DB.Txn(false)
	defer txn.Abort()
	var raw interface{}
	raw, err = txn.First(constants.SchemaDefinitionsTable, "id", tenantID, ruleName, version)
	if err != nil {
		return nil, "", errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	def, ok := raw.(storage.SchemaDefinition)
	if ok {
		var sch *base.SchemaDefinition
		sch, err = schema.NewSchemaFromStringDefinitions(false, def.Serialized())
		if err != nil {
			return nil, "", err
		}
		definition, err = schema.GetRuleByName(sch, ruleName)
		if err != nil {
			return nil, "", err
		}
		return definition, def.Version, err
	}

	return nil, "", errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_NOT_FOUND.String())
}

// HeadVersion - Reads the latest version from the repository.
func (r *SchemaReader) HeadVersion(_ context.Context, tenantID string) (string, error) {
	mu.Lock()
	defer mu.Unlock()

	version, ok := headVersion[tenantID]
	if !ok {
		return "", errors.New(base.ErrorCode_ERROR_CODE_SCHEMA_NOT_FOUND.String())
	}

	return version, nil
}

// ListSchemas - List all Schemas
func (r *SchemaReader) ListSchemas(_ context.Context, tenantID string, pagination database.Pagination) (schemas []*base.SchemaList, ct database.EncodedContinuousToken, err error) {
	txn := r.database.DB.Txn(false)
	defer txn.Abort()

	var lowerBound string
	if pagination.Token() != "" {
		var t database.ContinuousToken
		t, err = utils.EncodedContinuousToken{Value: pagination.Token()}.Decode()
		if err != nil {
			return nil, nil, err
		}
		lowerBound = t.(utils.ContinuousToken).Value
	}

	var result memdb.ResultIterator
	result, err = txn.LowerBound(constants.SchemaDefinitionsTable, "version", tenantID, lowerBound)
	if err != nil {
		return nil, nil, errors.New(base.ErrorCode_ERROR_CODE_EXECUTION.String())
	}

	schemas = make([]*base.SchemaList, 0, pagination.PageSize()+1)
	uniqueVersions := make(map[string]bool)
	for obj := result.Next(); obj != nil; obj = result.Next() {
		s, ok := obj.(storage.SchemaDefinition)
		if !ok {
			return nil, nil, errors.New(base.ErrorCode_ERROR_CODE_TYPE_CONVERSATION.String())
		}
		id, err := xid.FromString(s.Version)
		if err != nil {
			return nil, nil, errors.New(base.ErrorCode_ERROR_CODE_INTERNAL.String())
		}
		createdAt := id.Time().String()
		_, ok = uniqueVersions[s.Version]
		if !ok {
			schemas = append(schemas, &base.SchemaList{Version: s.Version, CreatedAt: createdAt})
			uniqueVersions[s.Version] = true
		}
		if len(schemas) > int(pagination.PageSize()) {
			return schemas[:pagination.PageSize()], utils.NewContinuousToken(fmt.Sprintf("%s|%s", s.Name, s.Version)).Encode(), nil
		}
	}

	return schemas, database.NewNoopContinuousToken().Encode(), err
}