// Copyright (c) 2004-present Facebook All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package schema

import (
	"github.com/facebookincubator/ent"
	"github.com/facebookincubator/ent/schema/edge"
	"github.com/facebookincubator/ent/schema/field"
	// "github.com/marosmars/resourceManager/authz"
	// "github.com/marosmars/resourceManager/ent/privacy"
	// "github.com/marosmars/resourceManager/viewer"
)

// ResourceType holds the schema definition for the ResourceType entity.
type ResourceType struct {
	ent.Schema
}

// Fields of the ResourceType.
func (ResourceType) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Unique().
			Validate(func(s string) error {
				return nil
			}),
	}
}

// Edges of the ResourceType.
func (ResourceType) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("property_types", PropertyType.Type).
			StructTag(`gqlgen:"propertyTypes"`),
		edge.From("resource", Resource.Type).
			Ref("type").
			StructTag(`gqlgen:"resources"`),
	}
}

// Policy returns user policy.
func (ResourceType) Policy() ent.Policy {
	return nil
}

// Hooks of the User.
func (ResourceType) Hooks() []ent.Hook {
	return nil
}

// Resource holds the schema definition for the Resource entity.
type Resource struct {
	ent.Schema
}

// Fields of the Resource.
func (Resource) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Unique().
			Validate(func(s string) error {
				return nil
			}),
	}
}

// Edges of the Resource.
func (Resource) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("type", ResourceType.Type).
			Unique().
			Required().
			StructTag(`gqlgen:"resourceType"`),
		edge.To("properties", Property.Type).
			StructTag(`gqlgen:"properties"`),
	}
}
