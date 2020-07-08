package pools

import (
    "context"
    "fmt"
    "github.com/marosmars/resourceManager/authz"
    "github.com/marosmars/resourceManager/authz/models"
    "log"
    "reflect"
    "strings"

    "github.com/marosmars/resourceManager/ent"
    "github.com/pkg/errors"
)

// TxFunction is a WithTx lambda
type TxFunction func(tx *ent.Tx) (interface{}, error)

// WithTx function executes a lambda function within a tx
func WithTx(
    ctx context.Context,
    client *ent.Client,
    fn TxFunction) (interface{}, error) {
    tx, err := client.Tx(ctx)
    if err != nil {
        return nil, err
    }
    defer func() {
        if v := recover(); v != nil {
            tx.Rollback()
            panic(v)
        }
    }()
    retVal, err := fn(tx)
    if err != nil {
        if rerr := tx.Rollback(); rerr != nil {
            err = errors.Wrapf(err, "rolling back transaction: %v", rerr)
        }
        return nil, err
    }
    if err := tx.Commit(); err != nil {
        return nil, errors.Wrapf(err, "committing transaction: %v", err)
    }
    return retVal, nil
}

func GetContext() context.Context {
    ctx := context.Background()
    ctx = authz.NewContext(ctx, &models.PermissionSettings{
        CanWrite:        true,
        WorkforcePolicy: authz.NewWorkforcePolicy(true, true)})
    return ctx
}

func OpenTestDb(ctx context.Context) *ent.Client {
    client, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
    if err != nil {
        log.Fatalf("failed opening connection to sqlite: %v", err)
    }
    // run the auto migration tool.
    if err := client.Schema.Create(ctx); err != nil {
        log.Fatalf("failed creating schema resources: %v", err)
    }

    return client
}

func HasPropertyTypeExistingProperties(
	ctx context.Context,
	client *ent.Client,
	propertyTypeID int) bool {

    propertyType, err := client.PropertyType.Get(ctx, propertyTypeID)

    if err != nil {
        //TODO error handling
    	return true //not sure
    }
    exists, err2 := client.PropertyType.QueryProperties(propertyType).Exist(ctx)

    if err2 != nil {
        //TODO error handling
        return true //not sure
    }
    //TODO also check resource type association??

	return exists
}

//TODO refactor
func UpdatePropertyType(
    ctx context.Context,
    client *ent.Client,
    propertyTypeID int,
    name string,
    typeName interface{},
    initValue interface{}) error {

    propertyTypeName := fmt.Sprintf("%v", typeName)

    prop := client.PropertyType.UpdateOneID(propertyTypeID).
        SetName(name).
        SetType(propertyTypeName).
        SetMandatory(true)

    //TODO we support int, but we always get int64 instead of int
    if reflect.TypeOf(initValue).String() == "int64" {
        initValue = int(initValue.(int64))
    }

    in := []reflect.Value{reflect.ValueOf(initValue)}
    reflect.ValueOf(prop).MethodByName("Set" + strings.Title(propertyTypeName) + "Val").Call(in)

    return prop.Exec(ctx)
}

func CreatePropertyType(
	ctx context.Context,
	client *ent.Client,
    name string,
    typeName interface{},
    initValue interface{}) (*ent.PropertyType, error) {

    propertyTypeName := fmt.Sprintf("%v", typeName)

    prop := client.PropertyType.Create().
        SetName(name).
        SetType(propertyTypeName).
        SetMandatory(true)

    //TODO we support int, but we always get int64 instead of int
    if reflect.TypeOf(initValue).String() == "int64" {
        initValue = int(initValue.(int64))
    }

    in := []reflect.Value{reflect.ValueOf(initValue)}
    reflect.ValueOf(prop).MethodByName("Set" + strings.Title(propertyTypeName) + "Val").Call(in)

   return prop.Save(ctx)
}

func CheckIfPoolsExist(
    ctx context.Context,
    client *ent.Client,
    resourceTypeID int) (bool, *ent.ResourceType) {
    resourceType, err := client.ResourceType.Get(ctx, resourceTypeID)
    if err != nil {
        //TODO add annoying GO error handling
        return true, resourceType //fix we don't know
    }

    //there can't be any existing pools
    count, err2 := resourceType.QueryPools().Count(ctx)

    if err2 != nil || count > 0 {
        //TODO add annoying GO error handling
        return true, resourceType //fix we don't know
    }

    return false, resourceType
}