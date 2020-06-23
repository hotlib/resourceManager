package main

import (
    "context"

    "github.com/marosmars/resourceManager/ent"
    resource "github.com/marosmars/resourceManager/ent/resource"
    resourcePool "github.com/marosmars/resourceManager/ent/resourcepool"
    "github.com/pkg/errors"
)

// Scope is a unique identifier of a resource claim
type Scope struct {
    scope string
}

// Pool is a resource provider
type Pool interface {
    claimResource(scope Scope) (*ent.Resource, error)
    freeResource(scope Scope) (*ent.Resource, error)
    queryResource(scope Scope) (*ent.Resource, error)
}

// SingletonPool provides only a single resource that can be reclaimed under various scopes
type SingletonPool struct {
    *ent.ResourcePool
    
    ctx    context.Context
    client *ent.Client
}

func (pool SingletonPool) claimResource(scope Scope) (*ent.Resource, error) {
    claim, err := WithTx(pool.ctx, pool.client, func(tx *ent.Tx) (interface{}, error) {
        return pool.claimResourceInner(scope)
    })

    if err != nil {
        return nil, errors.Wrapf(err, "Unable to claim a resource")
    }

    return claim.(*ent.Resource), nil
}

func (pool SingletonPool) claimResourceInner(scope Scope) (*ent.Resource, error) {
    res, err := pool.queryResource(scope)

    // Resource exists for the scope, return the same one
    if res != nil {
        return res, nil
    }

    // Allocate new resource for this scope
    if ent.IsNotFound(err) {
        return pool.client.Resource.Create().
            SetPool(pool.ResourcePool).
            SetScope(scope.scope).
            Save(pool.ctx)

        // TODO properties
        // TODO do not allocate a new resource here, it should be predefined for singleton
    }

    return res, err
}

func (pool SingletonPool) freeResource(scope Scope) {
    pool.client.Resource.Delete().
        Where(resource.HasPoolWith(resourcePool.ID(pool.ID))).
        Where(resource.ScopeEQ(scope.scope)).
        Exec(pool.ctx)
}

func (pool SingletonPool) queryResource(scope Scope) (*ent.Resource, error) {
    resource, err := pool.client.Resource.Query().
        Where(resource.HasPoolWith(resourcePool.ID(pool.ID))).
        Where(resource.ScopeEQ(scope.scope)).
        Only(pool.ctx)

    return resource, err
}
