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
            Unique(),
    }
}

// Edges of the ResourceType.
func (ResourceType) Edges() []ent.Edge {
    return []ent.Edge{
        edge.To("property_types", PropertyType.Type),
        edge.To("pools", ResourcePool.Type),
    }
}

// Policy returns ResourceType policy.
func (ResourceType) Policy() ent.Policy {
    return nil
}

// Hooks of the ResourceType.
func (ResourceType) Hooks() []ent.Hook {
    return nil
}

// ResourcePool holds the schema definition for the Resource pool entity.
type ResourcePool struct {
    ent.Schema
}

// Fields of the ResourcePool.
func (ResourcePool) Fields() []ent.Field {
    return []ent.Field{
        field.String("name").
            NotEmpty().
            Unique(),
        field.Enum("pool_type").
            Values("singleton", "set"),
    }
}

// Edges of the ResourcePool.
func (ResourcePool) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("resource_type", ResourceType.Type).
            Ref("pools").
            Unique(),
        edge.To("claims", Resource.Type),
    }
}

// Resource holds the schema definition for the Resource entity.
type Resource struct {
    ent.Schema
}

// Fields of the Resource.
func (Resource) Fields() []ent.Field {
    return []ent.Field{
        field.String("scope").
            NotEmpty().
            Unique(),
    }
}

// Edges of the Resource.
func (Resource) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("pool", ResourcePool.Type).
            Ref("claims").
            Unique(),
        edge.To("properties", Property.Type),
    }
}
